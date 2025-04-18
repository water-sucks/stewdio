package server

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	cmdUtils "stewdio/internal/cmd/utils"
	"stewdio/internal/refs"
	"stewdio/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
)

type ServerOpts struct {
	Port         int
	DataLocation string
}

type Server struct {
	DataDir string
	Router  *chi.Mux
}

func NewServer(dataDir string) *Server {
	s := &Server{
		DataDir: dataDir,
		Router:  chi.NewRouter(),
	}

	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", s.ListProjectsHandler)
		r.Post("/projects", s.CreateProjectHandler)
		r.Delete("/projects/{project}", s.DeleteProjectHandler)
		r.Get("/projects/{project}", s.GetProjectHandler)
		r.Get("/projects/{project}/pins", s.HandleGetVersionList)
		r.Post("/projects/{project}/pins", s.HandleUploadPin)
		r.Get("/projects/{project}/pins/{version}", s.HandleFetchVersion)
		r.Get("/projects/{project}/pins/{version}/file", s.HandleFetchFile)
	})

	return s
}

func ServerCommand() *cobra.Command {
	opts := ServerOpts{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a sync server for hosting Stew projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdUtils.CommandErrorHandler(serverMain(&opts))
		},
	}

	cmd.Flags().IntVarP(&opts.Port, "port", "p", 6969, "Port to listen on")
	cmd.Flags().StringVarP(&opts.DataLocation, "data", "d", "./stewdio-data", "Directory in which to store data")

	return cmd
}

func serverMain(opts *ServerOpts) error {
	dataDir := opts.DataLocation

	s := NewServer(dataDir)

	addr := fmt.Sprintf(":%d", opts.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}

	fmt.Println("hello, cruel world!")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("server listening on %s\n", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-done

	fmt.Println("\nshutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	fmt.Println("goodbye, cruel world!")

	return nil
}

// Handler methods
func (s *Server) ListProjectsHandler(w http.ResponseWriter, r *http.Request) {
	dirs, err := os.ReadDir(s.DataDir)
	if err != nil {
		fmt.Printf("error listing projects: %v\n", err)
		http.Error(w, "Unable to list projects", http.StatusInternalServerError)
		return
	}

	var projects []string
	for _, entry := range dirs {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	_ = json.NewEncoder(w).Encode(projects)
}

type createProjectRequest struct {
	Name string `json:"name"`
}

func (s *Server) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Missing project name", http.StatusBadRequest)
		return
	}

	path := filepath.Join(s.DataDir, req.Name)
	if utils.PathExists(path) {
		http.Error(w, "Project already exists", http.StatusConflict)
		return
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("Project created"))
}

func (s *Server) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")

	path := filepath.Join(s.DataDir, project)
	if err := os.RemoveAll(path); err != nil {
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("Project deleted"))
}

type projectInfoRes struct {
	Project      string    `json:"project"`
	LastModified time.Time `json:"lastModified"`
}

func (s *Server) GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	path := filepath.Join(s.DataDir, project)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Project not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error accessing project", http.StatusInternalServerError)
		}
		return
	}

	res := projectInfoRes{
		Project:      project,
		LastModified: info.ModTime(),
	}

	_ = json.NewEncoder(w).Encode(res)
}

type PinMetadata struct {
	Version string `json:"version"`
}

func (s *Server) HandleUploadPin(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	metaField := r.FormValue("meta")
	var meta PinMetadata
	if err := json.Unmarshal([]byte(metaField), &meta); err != nil {
		http.Error(w, "Invalid metadata JSON", http.StatusBadRequest)
		return
	}

	if meta.Version == "" {
		http.Error(w, "Missing version", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	projectDir := filepath.Join(s.DataDir, "projects", project, "objects", meta.Version)
	if utils.PathExists(projectDir) {
		http.Error(w, "Pin already exists", http.StatusConflict)
		return
	}

	err = os.MkdirAll(projectDir, 0o755)
	if err != nil {
		http.Error(w, "Failed to create project directory", http.StatusInternalServerError)
		return
	}

	dstPath := filepath.Join(projectDir, refs.ObjectTarName)

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Could not write file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = dst.Close() }()

	_, err = dst.ReadFrom(file)
	if err != nil {
		http.Error(w, "Failed to write data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("Pin uploaded"))
}

func (s *Server) HandleGetVersionList(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")

	versions, err := os.ReadDir(filepath.Join(s.DataDir, "projects", project, "objects"))
	if err != nil {
		http.Error(w, "Error accessing project", http.StatusInternalServerError)
		return
	}

	var versionsList []string
	for _, entry := range versions {
		if entry.IsDir() {
			versionsList = append(versionsList, entry.Name())
		}
	}

	sortVersionNumbers(versionsList)

	_ = json.NewEncoder(w).Encode(versionsList)
}

func (s *Server) HandleFetchVersion(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	version := chi.URLParam(r, "version")

	pinPath := filepath.Join(s.DataDir, "projects", project, "objects", version, refs.ObjectTarName)

	file, err := os.Open(pinPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("version not found: %v", err), http.StatusNotFound)
		return
	}
	defer func() { _ = file.Close() }()

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.tar.gz"`, version))

	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "failed to stream file", http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleFetchFile(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	version := chi.URLParam(r, "version")

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Missing file parameter", http.StatusBadRequest)
		return
	}

	tarPath := filepath.Join(s.DataDir, "projects", project, "objects", version, refs.ObjectTarName)

	f, err := os.Open(tarPath)
	if err != nil {
		http.Error(w, "Failed to open archive: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		http.Error(w, "Failed to read gzip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			http.Error(w, "File not found in archive", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Failed to read tar: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if strings.HasPrefix(hdr.Name, "files/") && hdr.Name == "files/"+filename {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "inline; filename=\""+path.Base(filename)+"\"")

			if _, err := io.Copy(w, tr); err != nil {
				log.Printf("Failed to write file to response: %v", err)
			}
			return
		}
	}
}

func sortVersionNumbers(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		mi := refs.ParseVersion(versions[i])
		mj := refs.ParseVersion(versions[j])

		if mi.Major != mj.Major {
			return mi.Major < mj.Major
		}

		return mi.Minor < mj.Minor
	})
}

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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
	})

	return s
}

func ServerCommand() *cobra.Command {
	opts := ServerOpts{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a sync server for hosting Stew projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serverMain(&opts)
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
	exists, err := pathExists(path)
	if err != nil {
		http.Error(w, "Error accessing project", http.StatusInternalServerError)
		return
	}
	if exists {
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

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

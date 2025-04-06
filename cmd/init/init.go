package init_cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"

	"stewdio/internal/config"
	"stewdio/internal/pin"
	"stewdio/internal/refs"
	tar_utils "stewdio/internal/tar"

	cmdUtils "stewdio/internal/cmd/utils"

	"github.com/spf13/cobra"
)

type initOpts struct {
	Name   string
	Remote string
}

func InitCommand() *cobra.Command {
	opts := initOpts{}

	cmd := cobra.Command{
		Use:          "init {PROJECT_NAME} -r {REMOTE}",
		Short:        "Initialize current directory as stewdio project with specified remote",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			opts.Name = args[0]

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdUtils.CommandErrorHandler(initMain(&opts))
		},
	}

	cmd.Flags().StringVarP(&opts.Remote, "remote", "r", "", "Remote repository URL")
	_ = cmd.MarkFlagRequired("remote")

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [PROJECT_NAME]   Name of audio project
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func initMain(opts *initOpts) error {
	fmt.Println("Initializing project ", opts.Name)
	fmt.Println("Using remote ", opts.Remote)

	cwd, _ := os.Getwd()

	createDotStew()

	cfg, err := config.CreateConfig(cwd, opts.Name, opts.Remote)
	if err != nil {
		fmt.Println("error creating config:", err)
		return err
	}
	_ = cfg

	err = pin_utils.Push(cwd, cfg.Remote, "0.1")
	if err != nil {
		fmt.Println("error pushing initial pin version 0.1:", err)
		return err
	}

	return nil
}

func createDotStew() {
	err := os.Mkdir(".stew", 0o775)
	if err != nil {
		fmt.Println("Could not initialize .stew directory:", err)
		return
	}

	versionFile, err := os.Create(".stew/version")
	if err != nil {
		fmt.Println("Could not create version file:", err)
		return
	}
	defer func() { _ = versionFile.Close() }()

	_, err = versionFile.WriteString("0.1")
	if err != nil {
		fmt.Println("Could not write to version file:", err)
		return
	}

	err = os.MkdirAll(".stew/objects/0.1", 0o775)
	if err != nil {
		fmt.Println("Could not create objects directory:", err)
		return
	}

	refsFile, err := os.Create(".stew/objects/0.1/refs")
	if err != nil {
		fmt.Println("Could not create refs file:", err)
		return
	}
	defer func() { _ = refsFile.Close() }()

	var wavFiles []string
	snapshot := make(map[string]bool)

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".wav" {
			snapshot[path] = true
			wavFiles = append(wavFiles, path)
			_, err = refsFile.WriteString(path + "\n")
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Could not walk through directory to find .wav files:", err)
		return
	}

	createInitialArchive(refs.Version{Major: 0, Minor: 1}, snapshot)
}

func createInitialArchive(version refs.Version, snapshot map[string]bool) {
	dir := fmt.Sprintf(".stew/objects/%d.%d", version.Major, version.Minor)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	tarPath := filepath.Join(dir, refs.ObjectTarName)
	tarFile, err := os.Create(tarPath)
	if err != nil {
		panic(err)
	}
	defer tarFile.Close()

	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Write initial message
	message := fmt.Sprintf("Initial version %d.%d", version.Major, version.Minor)
	tar_utils.AddStringToTar(tarWriter, "message", message)

	// Write empty diffs.json
	tar_utils.AddBytesToTar(tarWriter, "diffs.json", []byte("[]"))

	// Add all .wav files into files/
	for file := range snapshot {
		tar_utils.AddFileToTar(tarWriter, file, filepath.Join("files", file))
	}
}

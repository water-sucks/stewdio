package pin

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cmdUtils "stewdio/internal/cmd/utils"
	"stewdio/internal/config"
	pin_utils "stewdio/internal/pin"
	"stewdio/internal/refs"
	"stewdio/internal/tar"
)

type PinOpts struct {
	Message string
}

func PinCommand() *cobra.Command {
	opts := PinOpts{}

	cmd := cobra.Command{
		Use:   "pin",
		Short: "Snapshot current audio project to a new version and sync with remote",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(0)(cmd, args); err != nil {
				return err
			}

			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdUtils.CommandErrorHandler(pinMain(&opts))
		},
	}

	cmd.Flags().String("message", "", "Version message")

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [VERSION]   The version to checkout
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func pinMain(opts *PinOpts) error {
	cwd, _ := os.Getwd()

	if !refs.IsStewRepo(cwd) {
		msg := "error: current directory is not a stewdio project"
		fmt.Println(msg)
		return fmt.Errorf("%s", msg)
	}

	fmt.Println("Pinning current project...")

	version := refs.ReadVersion(cwd)
	version.Minor++

	refs.WriteVersion(cwd, version)

	snapshot := createSnapshot()

	diffs := computeDiffs(snapshot, version)

	storeSnapshotAndDiffs(version, snapshot, diffs, opts.Message)

	fmt.Println("Project pinned to version", version)

	cfg, err := config.ParseConfig(cwd)
	if err != nil {
		fmt.Println("error parsing config:", err)
		fmt.Println("unable to push pin, the repo is fucked")
		return err
	}

	err = pin_utils.Push(cwd, cfg.Remote, version.String())
	if err != nil {
		fmt.Printf("error pushing pin %v: %v\n", version.String(), err)
		fmt.Println("unable to push pin, the repo is fucked")
		return err
	}

	return nil
}

func createSnapshot() map[string]bool {
	snapshot := make(map[string]bool)
	_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".wav" {
			snapshot[path] = true
		}
		return nil
	})
	return snapshot
}

func computeDiffs(snapshot map[string]bool, version refs.Version) []refs.Diff {
	var diffs []refs.Diff
	previousSnapshot := readPreviousSnapshot(version)

	// Detect added files
	for file := range snapshot {
		if _, exists := previousSnapshot[file]; !exists {
			diffs = append(diffs, refs.Diff{
				File: file,
				Type: "added",
			})
		}
	}

	// Detect removed files
	for file := range previousSnapshot {
		if _, exists := snapshot[file]; !exists {
			diffs = append(diffs, refs.Diff{
				File: file,
				Type: "removed",
			})
		}
	}

	return diffs
}

func readPreviousSnapshot(version refs.Version) map[string]bool {
	if version.Minor == 1 {
		return make(map[string]bool)
	}

	// TODO: this sucks. Doesn't track minor versions
	prevVersion := refs.Version{Major: version.Major, Minor: version.Minor - 1}

	data, err := os.ReadFile(fmt.Sprintf(".stew/objects/%d.%d/refs", prevVersion.Major, prevVersion.Minor))
	if err != nil {
		return make(map[string]bool)
	}

	refs := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			refs[line] = true
		}
	}
	return refs
}

func storeSnapshotAndDiffs(version refs.Version, snapshot map[string]bool, diffs []refs.Diff, message string) {
	dir := fmt.Sprintf(".stew/objects/%d.%d", version.Major, version.Minor)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	tarFilePath := filepath.Join(dir, refs.ObjectTarName)
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		panic(err)
	}
	defer tarFile.Close()

	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	if message == "" {
		message = fmt.Sprintf("Pinned version %d.%d", version.Major, version.Minor)
	}

	tar_utils.AddStringToTar(tarWriter, "message", message)

	// 2. Write diffs.json
	diffBytes, err := json.MarshalIndent(diffs, "", "  ")
	if err != nil {
		panic(err)
	}
	tar_utils.AddBytesToTar(tarWriter, "diffs.json", diffBytes)

	// 3. Add added files under files/
	for _, diff := range diffs {
		if diff.Type == "added" {
			tar_utils.AddFileToTar(tarWriter, diff.File, filepath.Join("files", diff.File))
		}
	}
}

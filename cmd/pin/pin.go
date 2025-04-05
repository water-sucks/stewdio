package pin

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	cmdUtils "stewdio/internal/cmd/utils"
)

type pinOps struct{}

func PinCommand() *cobra.Command {
	opts := pinOps{}

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

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [VERSION]   The version to checkout
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func pinMain(opts *pinOps) error {
	fmt.Println("Pinning current project...")

	// Increment version
	version := readVersion()
	version.Minor++
	writeVersion(version)

	// Create snapshot
	snapshot := createSnapshot()

	// Compute diffs
	diffs := computeDiffs(snapshot, version)

	// Store diffs and snapshot
	storeSnapshotAndDiffs(version, snapshot, diffs)

	fmt.Println("Project pinned to version", version)

	return nil
}

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

type Diff struct {
	File string `json:"file"`
	Type string `json:"type"`
}

func readVersion() Version {
	data, err := os.ReadFile(".stew/version")
	if err != nil {
		panic(err)
	}
	versionStr := strings.TrimSpace(string(data))
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) != 2 {
		panic("Invalid version format")
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		panic(err)
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		panic(err)
	}
	return Version{
		Major: major,
		Minor: minor,
	}
}

func writeVersion(version Version) {
	versionStr := fmt.Sprintf("%d.%d", version.Major, version.Minor)
	if err := os.WriteFile(".stew/version", []byte(versionStr), 0o644); err != nil {
		panic(err)
	}
}

func createSnapshot() map[string]bool {
	snapshot := make(map[string]bool)
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
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

func computeDiffs(snapshot map[string]bool, version Version) []Diff {
	var diffs []Diff
	previousSnapshot := readPreviousSnapshot(version)

	// Detect added files
	for file := range snapshot {
		if _, exists := previousSnapshot[file]; !exists {
			diffs = append(diffs, Diff{
				File: file,
				Type: "added",
			})
		}
	}

	// Detect removed files
	for file := range previousSnapshot {
		if _, exists := snapshot[file]; !exists {
			diffs = append(diffs, Diff{
				File: file,
				Type: "removed",
			})
		}
	}

	return diffs
}

func readPreviousSnapshot(version Version) map[string]bool {
	if version.Minor == 1 {
		return make(map[string]bool)
	}
	prevVersion := Version{Major: version.Major, Minor: version.Minor - 1}
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

func storeSnapshotAndDiffs(version Version, snapshot map[string]bool, diffs []Diff) {
	dir := fmt.Sprintf(".stew/objects/%d.%d", version.Major, version.Minor)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	// Store diffs
	data, err := json.Marshal(diffs)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "diffs.json"), data, 0o644); err != nil {
		panic(err)
	}

	// Store refs
	var refs []string
	for file := range snapshot {
		refs = append(refs, file)
	}
	refsData := strings.Join(refs, "\n")
	if err := os.WriteFile(filepath.Join(dir, "refs"), []byte(refsData), 0o644); err != nil {
		panic(err)
	}

	// Create tar.gz for added files only if there are added files
	addedFiles := []string{}
	for _, diff := range diffs {
		if diff.Type == "added" {
			addedFiles = append(addedFiles, diff.File)
		}
	}
	if len(addedFiles) > 0 {
		tarFile, err := os.Create(filepath.Join(dir, "added_files.tar.gz"))
		if err != nil {
			panic(err)
		}
		defer tarFile.Close()
		gzWriter := gzip.NewWriter(tarFile)
		defer gzWriter.Close()
		tarWriter := tar.NewWriter(gzWriter)
		defer tarWriter.Close()
		for _, file := range addedFiles {
			addFileToTar(tarWriter, file)
		}
	}
}

func addFileToTar(tarWriter *tar.Writer, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		panic(err)
	}
	header.Name = filePath
	if err := tarWriter.WriteHeader(header); err != nil {
		panic(err)
	}
	if _, err := io.Copy(tarWriter, file); err != nil {
		panic(err)
	}
}

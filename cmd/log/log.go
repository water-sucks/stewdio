package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	cmdUtils "stewdio/internal/cmd/utils"
)

type logOpts struct {
	Limit int
}

func LogCmd() *cobra.Command {
	opts := logOpts{}

	cmd := cobra.Command{
		Use:   "log",
		Short: "Show pin history or version log",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return logMain(cmd, &opts)
		},
	}

	// Limit flag for specified amount of versions
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "Limit the number of log entries shown")

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  (None)      Shows the version history up to the limit specified
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func logMain(cmd *cobra.Command, opts *logOpts) error {
	// Path to the versions directory
	versionsDir := ".stew/versions"

	// Read version files (filtered for .tar.gz)
	files, err := readVersionFiles(versionsDir)
	if err != nil {
		return err
	}

	// Apply limit
	if opts.Limit > 0 && opts.Limit < len(files) {
		files = files[:opts.Limit]
	}

	// Print version files
	fmt.Printf("Showing last %d versions:\n", len(files))
	for _, file := range files {
		fmt.Println(file)
	}

	return nil
}

func readVersionFiles(versionsDir string) ([]string, error) {
	var versionFiles []string

	// Use os.ReadDir to read the directory
	dirEntries, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read versions directory: %v", err)
	}

	// Sort files in reverse order (most recent first)
	sort.Sort(sort.Reverse(sort.StringSlice(getDirs(dirEntries))))

	// Iterate over the directories (version directories)
	for _, entry := range dirEntries {
		// Only process directories
		if entry.IsDir() {
			// Construct the path to the version directory
			versionPath := filepath.Join(versionsDir, entry.Name())

			// Read files inside the version directory (filter for .tar.gz)
			versionFilesInDir, err := readTarGzFiles(versionPath)
			if err != nil {
				return nil, err
			}

			// Append valid files to the list
			versionFiles = append(versionFiles, versionFilesInDir...)
		}
	}

	return versionFiles, nil
}

// Helper function to extract directories from os.DirEntry
func getDirs(entries []os.DirEntry) []string {
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs
}

func readTarGzFiles(versionDir string) ([]string, error) {
	var tarGzFiles []string

	// Use os.ReadDir to read the version directory
	dirEntries, err := os.ReadDir(versionDir)
	if err != nil {
		return nil, fmt.Errorf("could not read version directory %s: %v", versionDir, err)
	}

	// Filter for .tar.gz files
	for _, entry := range dirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tar.gz") {
			tarGzFiles = append(tarGzFiles, filepath.Join(versionDir, entry.Name()))
		}
	}

	return tarGzFiles, nil
}


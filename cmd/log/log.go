package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

	// Limit flag for specified amount of bitches
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "Limit the number of log entries shown")

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  (None)      Shows the pin history up to the limit specified
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func logMain(cmd *cobra.Command, opts *logOpts) error {
	// Path to the bitches
	versionsDir := ".stew/versions"

	// Read them bitches
	files, err := readVersionFiles(versionsDir)
	if err != nil {
		return err
	}

	// Limit them bitches
	if opts.Limit > 0 && opts.Limit < len(files) {
		files = files[:opts.Limit]
	}

	// Print them bitches out
	fmt.Printf("Showing last %d versions:\n", len(files))
	for _, file := range files {
		// TODO: Adjust based on file type
		fmt.Println(file)
	}

	return nil
}

func readVersionFiles(versionsDir string) ([]string, error) {
	var versionFiles []string

	// Open the directory
	dir, err := os.Open(versionsDir)
	if err != nil {
		return nil, fmt.Errorf("could not open versions directory: %v", err)
	}
	defer dir.Close()

	// Read all bitches in the dir
	files, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("could not read files in versions directory: %v", err)
	}

	// Sort bitches to most recent on top
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, file := range files {
		// Possibly filter bitches for specific file types
		versionFiles = append(versionFiles, file)
	}

	return versionFiles, nil
}

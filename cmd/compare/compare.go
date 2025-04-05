package compare

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cmdUtils "stewdio/internal/cmd/utils"
)

type compareOpts struct {
	OldFile string
	NewFile string
	Output  string
}

func CompareCmd() *cobra.Command {
	opts := compareOpts{}

	cmd := &cobra.Command{
		Use:   "compare {OLD_FILE} {NEW_FILE} {OUTPUT}",
		Short: "Compare two audio files and produce binary diffs",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.OldFile = args[0]
			opts.NewFile = args[1]
			opts.Output = args[2]
			return cmdUtils.CommandErrorHandler(compareMain(&opts))
		},
	}

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [OLD_FILE]   The path to the old audio file
  [NEW_FILE]   The path to the new audio file
  [OUTPUT]     The path to the output directory
`)
	cmdUtils.SetHelpFlagText(cmd)

	return cmd
}

func compareMain(opts *compareOpts) error {
	oldFile, err := os.Open(opts.OldFile)
	if err != nil {
		return err
	}
	defer oldFile.Close()

	newFile, err := os.Open(opts.NewFile)
	if err != nil {
		return err
	}
	defer newFile.Close()

	// Generate diffs
	err = GenerateDiffs(oldFile, newFile, opts.Output)
	if err != nil {
		return err
	}

	fmt.Println("Diffs generated successfully")

	return nil
}

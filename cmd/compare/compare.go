package compare

import (
	"fmt"
	"os"

	"github.com/go-audio/wav"
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

	err = generateDiffs(oldFile, newFile, opts.Output)
	if err != nil {
		return err
	}

	fmt.Println("Diffs generated successfully")
	return nil
}

func generateDiffs(oldFile, newFile *os.File, outputPath string) error {
	oldDecoder := wav.NewDecoder(oldFile)
	oldAudioBuf, err := oldDecoder.FullPCMBuffer()
	if err != nil {
		return err
	}

	newDecoder := wav.NewDecoder(newFile)
	newAudioBuf, err := newDecoder.FullPCMBuffer()
	if err != nil {
		return err
	}

	if oldDecoder.BitDepth != newDecoder.BitDepth {
		return fmt.Errorf("bit depth mismatch: old file is %d-bit, new file is %d-bit", oldDecoder.BitDepth, newDecoder.BitDepth)
	}

	additions, subtractions := calculateDiffs(oldAudioBuf.Data, newAudioBuf.Data, int(oldDecoder.BitDepth), getFormat(oldDecoder))

	err = os.WriteFile(fmt.Sprintf("%s/additions.bin", outputPath), additions, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/subtractions.bin", outputPath), subtractions, 0644)
	if err != nil {
		return err
	}

	return nil
}

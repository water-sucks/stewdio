package patchCommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"stewdio/cmd/pin"
	cmdUtils "stewdio/internal/cmd/utils"
)

type patchOpts struct {
	TargetFile string
	PatchFile  string
}

func PatchCmd() *cobra.Command {
	opts := patchOpts{}

	cmd := cobra.Command{
		Use:   "patch {TARGET_FILE} {PATCH_FILE}",
		Short: "Apply a patch to a target file",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}

			opts.TargetFile = args[0]
			opts.PatchFile = args[1]

			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdUtils.CommandErrorHandler(patchMain(cmd, &opts))
		},
	}

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [TARGET_FILE]   The target file to patch
  [PATCH_FILE]    The patch file to apply
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func patchMain(cmd *cobra.Command, opts *patchOpts) error {
	err := pin.ApplyPatch(opts.TargetFile, opts.PatchFile)
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}
	fmt.Println("Patch applied successfully")
	return nil
}

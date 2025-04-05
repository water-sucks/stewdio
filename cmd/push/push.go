package push

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdUtils "stewdio/internal/cmd/utils"
)

type PushOpts struct {
	Repository string
}

func PushCmd() *cobra.Command {
	opts := PushOpts{}

	cmd := cobra.Command{
		Use:   "push {REPOSITORY}",
		Short: "Fetch a Repository and make it active",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			opts.Repository = args[0]

			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return PushMain(cmd, &opts)
		},
	}

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [REPOSITORY]   The Repository to push to
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func PushMain(cmd *cobra.Command, opts *PushOpts) error {
	fmt.Println("push:", opts.Repository)

	return nil
}

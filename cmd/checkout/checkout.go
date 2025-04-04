package checkout

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdUtils "stewdio/internal/cmd/utils"
)

type checkoutOpts struct {
	Version string
}

func CheckoutCmd() *cobra.Command {
	opts := checkoutOpts{}

	cmd := cobra.Command{
		Use:   "checkout {VERSION}",
		Short: "Fetch a version and make it active",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			opts.Version = args[0]

			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkoutMain(cmd, &opts)
		},
	}

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [VERSION]   The version to checkout
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func checkoutMain(cmd *cobra.Command, opts *checkoutOpts) error {
	fmt.Println("checkout:", opts.Version)

	return nil
}

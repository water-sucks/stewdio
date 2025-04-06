package cmd

import (
	"os"

	"stewdio/cmd/checkout"
	"stewdio/cmd/compare"
	"stewdio/cmd/init"
	patchCommand "stewdio/cmd/patch"
	"stewdio/cmd/pin"
	"stewdio/cmd/server"

	"github.com/spf13/cobra"
)

func MainCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "stewdio {command} [flags]",
		Short: "Stewdio CLI",
		Long:  "Stewdio CLI - i like soup!",
	}

	cmd.SetErrPrefix("error:")

	cmd.Flags().BoolP("help", "h", false, "Show this help menu")

	cmd.SetHelpCommand(&cobra.Command{
		Hidden:       true,
		SilenceUsage: true,
	})
	cmd.CompletionOptions.HiddenDefaultCmd = true

	cmd.AddCommand(checkout.CheckoutCmd())
	cmd.AddCommand(init_cmd.InitCommand())
	cmd.AddCommand(pin.PinCommand())
	cmd.AddCommand(server.ServerCommand())
	cmd.AddCommand(compare.CompareCmd())
	cmd.AddCommand(patchCommand.PatchCmd())

	return &cmd
}

func Execute() {
	cmd := MainCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

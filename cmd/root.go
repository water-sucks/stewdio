package cmd

import (
	"os"

	"stewdio/cmd/checkout"
	"stewdio/cmd/init"
	"stewdio/cmd/log"
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
	cmd.AddCommand(init_cmd.InitCMD())
	cmd.AddCommand(pin.PinCMD())
	cmd.AddCommand(server.ServerCommand())
	cmd.AddCommand(log.LogCmd())

	return &cmd
}

func Execute() {
	cmd := MainCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

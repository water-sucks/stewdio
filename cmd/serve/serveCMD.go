package serve

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdUtils "stewdio/internal/cmd/utils"
)

func ServeCmd() *cobra.Command {

	cmd := cobra.Command{
		Use:   "serve",
		Short: "Run a server for hosting project",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(0)(cmd, args); err != nil {
				return err
			}

			return nil
		},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveMain(cmd)
		},
	}

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
    This command takes no arguments
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func serveMain(cmd *cobra.Command) error {
	fmt.Println("Starting server.")

	return nil
}

package init_cmd

import (
	"fmt"
	"os"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/spf13/cobra"
	cmdUtils "stewdio/internal/cmd/utils"
)

type initOpts struct {
	Name   string
	Remote string
}

func InitCMD() *cobra.Command {
	opts := initOpts{}

	cmd := cobra.Command{
		Use:          "init {PROJECT_NAME} -r {REMOTE}",
		Short:        "Initialize current directory as stewdio project with specified remote",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			opts.Name = args[0]

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return initMain(cmd, &opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Remote, "remote", "r", "", "Remote repository URL")
	cmd.MarkFlagRequired("remote")

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Arguments:
  [PROJECT_NAME]   Name of audio project
`)
	cmdUtils.SetHelpFlagText(&cmd)

	return &cmd
}

func initMain(cmd *cobra.Command, opts *initOpts) error {
	fmt.Println("Initializing project ", opts.Name)
	fmt.Println("Using remote ", opts.Remote)

	createDotStew(opts.Name, opts.Remote)

	return nil
}

func createDotStew(projectName string, remoteURL string) {
	err := os.Mkdir(".stew", 0775)
	if err != nil {
		fmt.Println("Could not initialize .stew directory:", err)
		return
	}

	k := koanf.New(".")
	remote := map[string]string{
		"server":  remoteURL,
		"project": projectName,
	}
	k.Load(rawbytes.Provider([]byte{}), nil)
	k.Set("remote", remote)

	tomlBytes, err := k.Marshal(toml.Parser())
	if err != nil {
		fmt.Println("Error marshalling to TOML:", err)
		return
	}

	file, err := os.Create(".stew/thamizh.toml")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(tomlBytes)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Configuration saved to .stew/thamizh.toml")
}

package init_cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	cmdUtils "stewdio/internal/cmd/utils"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/spf13/cobra"
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

	createDotStew()
	createTOML(opts.Name, opts.Remote)

	return nil
}

func createDotStew() {
	err := os.Mkdir(".stew", 0775)
	if err != nil {
		fmt.Println("Could not initialize .stew directory:", err)
		return
	}

	versionFile, err := os.Create(".stew/version")
	if err != nil {
		fmt.Println("Could not create version file:", err)
		return
	}
	defer versionFile.Close()
	_, err = versionFile.WriteString("0.1")
	if err != nil {
		fmt.Println("Could not write to version file:", err)
		return
	}

	err = os.MkdirAll(".stew/objects/0.1", 0775)
	if err != nil {
		fmt.Println("Could not create objects directory:", err)
		return
	}

	refsFile, err := os.Create(".stew/objects/0.1/refs")
	if err != nil {
		fmt.Println("Could not create refs file:", err)
		return
	}
	defer refsFile.Close()

	var wavFiles []string
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".wav" {
			wavFiles = append(wavFiles, path)
			_, err = refsFile.WriteString(path + "\n")
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Could not walk through directory to find .wav files:", err)
		return
	}

	err = compressFilesToTarGz(".stew/objects/0.1/audio_files.tar.gz", wavFiles)
	if err != nil {
		fmt.Println("Could not compress audio files:", err)
		return
	}
}

func compressFilesToTarGz(dest string, files []string) error {
	tarfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	gzwriter := gzip.NewWriter(tarfile)
	defer gzwriter.Close()

	tarwriter := tar.NewWriter(gzwriter)
	defer tarwriter.Close()

	for _, file := range files {
		err := addFileToTarWriter(file, tarwriter)
		if err != nil {
			return err
		}
	}
	return nil
}

func addFileToTarWriter(filename string, tw *tar.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}
	header.Name = filename

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}
	return nil
}

func createTOML(projectName string, remoteURL string) {
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

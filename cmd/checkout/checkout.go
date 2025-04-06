package checkout

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"stewdio/internal/config"
	"stewdio/internal/utils"
)

type CheckoutOpts struct {
	Version string
}

func CheckoutCmd() *cobra.Command {
	opts := CheckoutOpts{}

	return &cobra.Command{
		Use:   "checkout [project] [version]",
		Short: "Checkout a specific version of a project",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			opts.Version = args[0]

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			return checkoutMain(&opts, version)
		},
	}
}

func checkoutMain(opts *CheckoutOpts, version string) error {
	cwd, _ := os.Getwd()

	if !utils.PathExists(filepath.Join(cwd, ".stew")) {
		msg := "error: this is not a stewdio repository"
		fmt.Println(msg)
		return fmt.Errorf("%s", msg)
	}

	cfg, err := config.ParseConfig(cwd)
	if err != nil {
		fmt.Println("error parsing git config:", err)
		return err
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/pins/%v", cfg.Remote.Server, cfg.Remote.Project, opts.Version)

	// Make the GET request to fetch the version
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	// Create a directory to store the downloaded version
	versionDir := filepath.Join(".stew", cfg.Remote.Project, "versions", version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for version: %v", err)
	}

	// Create the file to save the downloaded tarball
	tarballPath := filepath.Join(versionDir, fmt.Sprintf("%s.tar.gz", version))
	outFile, err := os.Create(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy the response body to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	fmt.Printf("Successfully checked out version %s of project %s\n", version, cfg.Remote.Project)
	return nil
}

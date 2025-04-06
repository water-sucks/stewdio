package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

var k = koanf.New(".")

type RemoteConfig struct {
	Remote Remote `koanf:"remote"`
}

type Remote struct {
	Server  string `koanf:"server"`
	Project string `koanf:"project"`
}

func CreateConfig(path string, projectName string, remoteURL string) (*RemoteConfig, error) {
	cfg := RemoteConfig{
		Remote: Remote{
			Server:  remoteURL,
			Project: projectName,
		},
	}

	if err := k.Load(structs.Provider(cfg, "koanf"), nil); err != nil {
		return nil, fmt.Errorf("failed to load config struct: %w", err)
	}

	tomlBytes, err := k.Marshal(toml.Parser())
	if err != nil {
		return nil, fmt.Errorf("error marshalling to TOML: %w", err)
	}

	// Ensure config directory exists
	configPath := filepath.Join(path, ".stew")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	// Write to file
	configFile := filepath.Join(configPath, "thamizh.toml")
	if err := os.WriteFile(configFile, tomlBytes, 0o644); err != nil {
		return nil, fmt.Errorf("error writing to file: %w", err)
	}

	fmt.Println("Configuration saved to", configFile)
	return &cfg, nil
}

func ParseConfig(path string) (*RemoteConfig, error) {
	configFile := filepath.Join(path, ".stew", "thamizh.toml")

	// Load the TOML config file
	if err := k.Load(file.Provider(configFile), toml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	var cfg RemoteConfig
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &cfg, nil
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Editor      string   `json:"editor"`
	SearchPaths []string `json:"search_paths"`
}

func DefaultConfig() (*Config, error) {

	var homeDir, err = os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return &Config{
		Editor: "nvim",
		SearchPaths: []string{
			filepath.Join(homeDir, "dev"), filepath.Join(homeDir, "projects"), filepath.Join(homeDir, "repos"), filepath.Join(homeDir, "workspaces")},
	}, nil

}

func ConfigPath() (string, error) {

	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	return filepath.Join(configPath, "gf", "config.json"), nil
}

func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	var data []byte
	data, err = os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg, err := DefaultConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			err = cfg.Save()
			if err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	var data []byte
	data, err = json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

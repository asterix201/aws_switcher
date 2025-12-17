package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Account represents an AWS account with role information
type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RoleName string `json:"role_name"`
}

// Config holds the application configuration
type Config struct {
	StartURL  string    `json:"start_url"`
	SSORegion string    `json:"sso_region"`
	Accounts  []Account `json:"accounts"`
}

// configPath returns the path to the config file
var configPath = func() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	awsDir := filepath.Join(homeDir, ".aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(awsDir, "switcher-config.json"), nil
}

// Load loads the configuration from file
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		SSORegion: "us-east-1", // default
		Accounts:  make([]Account, 0),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // return default config if file doesn't exist
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

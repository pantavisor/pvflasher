package util

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Theme string `json:"theme"` // "light" or "dark"
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Theme: "system",
	}
}

// GetConfigDir returns the path to the configuration directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".pvflasher")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return DefaultConfig(), err
	}

	configFile := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return DefaultConfig(), err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644)
}

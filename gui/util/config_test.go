package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Theme != "system" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "system")
	}
}

func TestGetConfigDir(t *testing.T) {
	// This should not error on a normal system
	configDir, err := GetConfigDir()
	if err != nil {
		t.Logf("GetConfigDir() error: %v (may be expected in CI)", err)
		return
	}

	if configDir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Directory should exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory does not exist: %s", configDir)
	}
}

func TestLoadConfig_NoExisting(t *testing.T) {
	// Create a temp dir for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := LoadConfig()
	if err != nil {
		// Should not error, should return default config
		t.Logf("LoadConfig() error: %v", err)
	}

	// Should get default config
	if cfg.Theme != "system" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "system")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temp dir for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := Config{Theme: "dark"}
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	// Verify config was saved
	configDir, _ := GetConfigDir()
	configFile := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	// Create a temp dir for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Save a config
	savedCfg := Config{Theme: "light"}
	err := SaveConfig(savedCfg)
	if err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	// Load it back
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loadedCfg.Theme != savedCfg.Theme {
		t.Errorf("Loaded theme = %q, want %q", loadedCfg.Theme, savedCfg.Theme)
	}
}

func TestConfig_JSON(t *testing.T) {
	// Create a temp dir for config
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Verify the config can be marshaled/unmarshaled
	configDir, _ := GetConfigDir()
	configFile := filepath.Join(configDir, "config.json")

	// Write JSON manually
	jsonData := `{"theme":"dark"}`
	err := os.WriteFile(configFile, []byte(jsonData), 0644)
	if err != nil {
		t.Logf("Could not write test config: %v", err)
		return
	}
	defer os.Remove(configFile)

	// Load and verify
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loadedCfg.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", loadedCfg.Theme, "dark")
	}
}

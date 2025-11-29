package userconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Telemetry {
		t.Error("expected Telemetry to default to true")
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.toml")

	cfg, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Telemetry {
		t.Error("expected default Telemetry=true when file missing")
	}
}

func TestLoadExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.toml")

	// Write config with telemetry disabled
	err := os.WriteFile(path, []byte("telemetry = false\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Telemetry {
		t.Error("expected Telemetry=false from file")
	}
}

func TestLoadInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.toml")

	// Write invalid TOML
	err := os.WriteFile(path, []byte("this is not valid toml [[["), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = loadFromPath(path)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "config.toml")

	cfg := &Config{Telemetry: false}
	if err := cfg.saveToPath(path); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if loaded.Telemetry != false {
		t.Error("expected Telemetry=false after save/load")
	}
}

func TestGetTelemetry(t *testing.T) {
	cfg := &Config{Telemetry: true}
	val, ok := cfg.Get("telemetry")
	if !ok {
		t.Error("expected telemetry key to exist")
	}
	if val != "true" {
		t.Errorf("expected 'true', got %q", val)
	}

	cfg.Telemetry = false
	val, ok = cfg.Get("telemetry")
	if !ok {
		t.Error("expected telemetry key to exist")
	}
	if val != "false" {
		t.Errorf("expected 'false', got %q", val)
	}
}

func TestGetUnknownKey(t *testing.T) {
	cfg := DefaultConfig()
	_, ok := cfg.Get("unknown")
	if ok {
		t.Error("expected unknown key to return false")
	}
}

func TestSetTelemetry(t *testing.T) {
	cfg := DefaultConfig()

	if err := cfg.Set("telemetry", "false"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Telemetry {
		t.Error("expected Telemetry=false")
	}

	if err := cfg.Set("telemetry", "true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Telemetry {
		t.Error("expected Telemetry=true")
	}

	// Test case insensitivity
	if err := cfg.Set("TELEMETRY", "false"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Telemetry {
		t.Error("expected Telemetry=false (case insensitive)")
	}
}

func TestSetInvalidValue(t *testing.T) {
	cfg := DefaultConfig()

	err := cfg.Set("telemetry", "invalid")
	if err == nil {
		t.Error("expected error for invalid boolean value")
	}
}

func TestSetUnknownKey(t *testing.T) {
	cfg := DefaultConfig()

	err := cfg.Set("unknown", "value")
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestAvailableKeys(t *testing.T) {
	keys := AvailableKeys()
	if _, ok := keys["telemetry"]; !ok {
		t.Error("expected telemetry in available keys")
	}
}

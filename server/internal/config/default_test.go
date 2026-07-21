package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ensures that the default configuration passes config validation
func TestDefaultConfigIsValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pocketry.toml")

	if err := os.WriteFile(path, DefaultConfig(), 0600); err != nil {
		t.Fatalf("write default config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Version != CurrentVersion {
		t.Fatalf("expected version %d, got %d", CurrentVersion, cfg.Version)
	}
}
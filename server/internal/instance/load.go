package instance

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CDX-1/pocketry/internal/config"
)

func Load(path string) (*Instance, error) {
	if path == "" {
		return nil, errors.New("Pocketry server directory is required")
	}

	rootDir, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve server directory: %w", err)
	}

	info, err := os.Stat(rootDir)
	if err != nil {
		return nil, fmt.Errorf("inspect server directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("server path is not a directory: %s", rootDir)
	}

	marker, err := loadMarker(filepath.Join(rootDir, markerFilename))
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(filepath.Join(rootDir, configFilename))
	if err != nil {
		return nil, fmt.Errorf("load Pocketry config: %w", err)
	}

	inst := &Instance{
		RootDir: rootDir,
		Marker:  marker,
		Config:  cfg,
	}

	if err := inst.validateRequiredFiles(); err != nil {
		return nil, err
	}

	return inst, nil
}

func loadMarker(path string) (Marker, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Marker{}, errors.New("directory is not a Pocketry server instance")
	}
	if err != nil {
		return Marker{}, fmt.Errorf("read instance marker: %w", err)
	}

	var marker Marker

	if err := json.Unmarshal(data, &marker); err != nil {
		return Marker{}, fmt.Errorf("decode instance marker: %w", err)
	}

	if err := validateMarker(marker); err != nil {
		return Marker{}, err
	}

	return marker, nil
}

func validateMarker(marker Marker) error {
	if marker.Format != formatName {
		return fmt.Errorf(
			"expected marker format: %q, got %q",
			formatName,
			marker.Format,
		)
	}

	if marker.FormatVersion != formatVersion {
		return fmt.Errorf(
			"expected format version: %d, got %d",
			formatVersion,
			marker.FormatVersion,
		)
	}

	if marker.InstanceID == "" {
		return fmt.Errorf("expected instance ID to be non-empty")
	}

	if marker.CreatedAt.IsZero() {
		return fmt.Errorf("expected created_at to be non-empty")
	}

	return nil
}

func (i *Instance) validateRequiredFiles() error {
	requiredFiles := []struct {
		name string
		path string
	}{
		{
			name: "database",
			path: i.DatabasePath(),
		},
		{
			name: "access token secret",
			path: i.AccessTokenSecretPath(),
		},
		{
			name: "OPAQUE server key material",
			path: i.OpaqueKeyMaterialPath(),
		},
	}

	for _, file := range requiredFiles {
		info, err := os.Stat(file.path)
		if err != nil {
			return fmt.Errorf("required file not found: %s (%s): %w", file.name, file.path, err)
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s path is not a regular file: %s", file.name, file.path)
		}
	}

	return nil
}

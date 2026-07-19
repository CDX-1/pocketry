package instance

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CDX-1/pocketry/internal/auth"
	"github.com/CDX-1/pocketry/internal/config"
	"github.com/CDX-1/pocketry/internal/db"
	"github.com/google/uuid"
)

const (
	formatName              = "pocketry-server"
	formatVersion           = 1
	accessTokenSecretHeader = "POCKETRY-ACCESS-TOKEN-SECRET-V1"
)

type Instance struct {
	RootDir string
	Marker  Marker
	Config  config.Config
}

type Marker struct {
	Format        string    `json:"format"`
	FormatVersion int       `json:"format_version"`
	InstanceID    string    `json:"instance_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func Initialize(path string) (*Instance, error) {
	rootDir, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve directory: %w", err)
	}

	if err := createTarget(rootDir); err != nil {
		return nil, err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(rootDir)
		}
	}()

	if err := createDirectories(rootDir); err != nil {
		return nil, err
	}

	if err := writeAccessTokenSecret(
		filepath.Join(rootDir, accessTokenRelativePath),
	); err != nil {
		return nil, fmt.Errorf("create access token key: %w", err)
	}

	if err := writeOpaqueServerKeyMaterial(
		filepath.Join(rootDir, opaqueRelativePath),
	); err != nil {
		return nil, fmt.Errorf("create OPAQUE server key material: %w", err)
	}

	if err := writeDefaultConfig(rootDir); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	if err := initializeDatabase(rootDir); err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	if err := writeMarker(rootDir); err != nil {
		return nil, fmt.Errorf("write instance marker: %w", err)
	}

	cleanup = false

	return &Instance{RootDir: rootDir}, nil
}

func createTarget(path string) error {
	_, err := os.Lstat(path)

	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return fmt.Errorf("inspect target directory: %w", err)
	default:
		return fmt.Errorf("target already exists: %s", path)
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	return nil
}

func createDirectories(rootDir string) error {
	directories := []string{
		filepath.Join(rootDir, "data"),
		filepath.Join(rootDir, "secrets"),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0700); err != nil {
			return fmt.Errorf("create directory: %s: %w", directory, err)
		}
	}

	return nil
}

func writeAccessTokenSecret(path string) error {
	secret, err := auth.GenerateAccessTokenSecret()
	if err != nil {
		return fmt.Errorf("generate access token secret: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(secret)
	content := []byte(accessTokenSecretHeader + "\n" + encoded + "\n")

	if err := os.WriteFile(path, content, 0600); err != nil {
		return fmt.Errorf("write access token secret: %w", err)
	}

	return nil
}

func writeOpaqueServerKeyMaterial(path string) error {
	encodedSKM, err := auth.GenerateOpaqueServerKeyMaterial(
		auth.DefaultOpaqueServerIdentity,
	)
	if err != nil {
		return fmt.Errorf("generate OPAQUE server key material: %w", err)
	}

	if err := os.WriteFile(path, encodedSKM, 0600); err != nil {
		return fmt.Errorf("write OPAQUE server key material: %w", err)
	}

	return nil
}

func writeDefaultConfig(rootDir string) error {
	path := filepath.Join(rootDir, configFilename)

	err := os.WriteFile(
		path,
		config.DefaultConfig(),
		0600,
	)
	if err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}

func initializeDatabase(rootDir string) error {
	databasePath := filepath.Join(rootDir, databaseRelativePath)

	store, err := db.Open(databasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := store.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}

func writeMarker(rootDir string) error {
	marker := Marker{
		Format:         formatName,
		FormatVersion:  formatVersion,
		InstanceID:     uuid.NewString(),
		CreatedAt:      time.Now().UTC(),
	}

	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return fmt.Errorf("encode instance marker: %w", err)
	}
	data = append(data, '\n')

	path := filepath.Join(rootDir, markerFilename)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write instance marker: %w", err)
	}

	return nil
}

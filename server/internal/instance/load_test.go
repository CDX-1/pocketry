package instance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupInitializedInstance(t *testing.T) string {
	t.Helper()

	targetDir := filepath.Join(t.TempDir(), "server")

	if _, err := Initialize(targetDir); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	return targetDir
}

func writeTestMarker(t *testing.T, rootDir string, marker Marker) {
	t.Helper()

	data, err := json.Marshal(marker)
	if err != nil {
		t.Fatalf("marshal marker: %v", err)
	}

	path := filepath.Join(rootDir, markerFilename)

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write marker: %v", err)
	}
}

// ensures that Load() successfully loads an initialized instance
func TestLoadValidInstance(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	inst, err := Load(targetDir)
	if err != nil {
		t.Fatalf("Load returned err: %v", err)
	}

	expectedRoot, err := filepath.Abs(targetDir)
	if err != nil {
		t.Fatalf("filepath.Abs returned error: %v", err)
	}

	if inst.RootDir != expectedRoot {
		t.Fatalf("expected root directory %q, got %q", expectedRoot, inst.RootDir)
	}

	if inst.Marker.Format != formatName {
		t.Fatalf("expected marker format %q, got %q", formatName, inst.Marker.Format)
	}

	if inst.Marker.FormatVersion != formatVersion {
		t.Fatalf("expected marker version %d, got %d", formatVersion, inst.Marker.FormatVersion)
	}

	if inst.Config.Version == 0 {
		t.Fatal("expected config to be loaded")
	}
}

// ensures that Load rejects an empty path
func TestLoadRejectsEmptyPath(t *testing.T) {
	_, err := Load("")
	if err == nil {
		t.Fatal("expected Load to reject an empty path")
	}

	if !strings.Contains(err.Error(), "Pocketry server directory is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a path that does not exist
func TestLoadRejectsMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected Load to reject a missing directory")
	}

	if !strings.Contains(err.Error(), "inspect server directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a path that is not a directory
func TestLoadRejectsFilePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server")

	if err := os.WriteFile(path, []byte("not a directory"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected Load to reject a file path")
	}

	if !strings.Contains(
		err.Error(),
		"server path is not a directory",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with a missing marker file
func TestLoadRejectsMissingMarker(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	if err := os.Remove(
		filepath.Join(targetDir, markerFilename),
	); err != nil {
		t.Fatalf("remove marker: %v", err)
	}

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject a missing marker")
	}

	if !strings.Contains(
		err.Error(),
		"directory is not a Pocketry server instance",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with a malformed marker file
func TestLoadRejectsMalformedMarker(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	path := filepath.Join(targetDir, markerFilename)

	if err := os.WriteFile(
		path,
		[]byte(`{"format":`),
		0600,
	); err != nil {
		t.Fatalf("write malformed marker: %v", err)
	}

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject malformed marker JSON")
	}

	if !strings.Contains(err.Error(), "decode instance marker") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with the wrong marker format
func TestLoadRejectsWrongMarkerFormat(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	writeTestMarker(t, targetDir, Marker{
		Format:        "not-pocketry",
		FormatVersion: formatVersion,
		InstanceID:    "test-instance-id",
		CreatedAt:     time.Now().UTC(),
	})

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject wrong marker format")
	}

	if !strings.Contains(err.Error(), "expected marker format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsUnsupportedFormatVersion(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	writeTestMarker(t, targetDir, Marker{
		Format:        formatName,
		FormatVersion: formatVersion + 1,
		InstanceID:    "test-instance-id",
		CreatedAt:     time.Now().UTC(),
	})

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject unsupported format version")
	}

	if !strings.Contains(err.Error(), "expected format version") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with an empty instance ID in the marker
func TestLoadRejectsEmptyInstanceID(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	writeTestMarker(t, targetDir, Marker{
		Format:        formatName,
		FormatVersion: formatVersion,
		InstanceID:    "",
		CreatedAt:     time.Now().UTC(),
	})

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject empty instance ID")
	}

	if !strings.Contains(
		err.Error(),
		"expected instance ID to be non-empty",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with a missing creation time in the marker
func TestLoadRejectsMissingCreationTime(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	writeTestMarker(t, targetDir, Marker{
		Format:        formatName,
		FormatVersion: formatVersion,
		InstanceID:    "test-instance-id",
		CreatedAt:     time.Time{},
	})

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject missing creation time")
	}

	if !strings.Contains(
		err.Error(),
		"expected created_at to be non-empty",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with a missing config file
func TestLoadRejectsMissingConfig(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	if err := os.Remove(
		filepath.Join(targetDir, configFilename),
	); err != nil {
		t.Fatalf("remove config: %v", err)
	}

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject missing config")
	}

	if !strings.Contains(err.Error(), "load Pocketry config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with an invalid config file
func TestLoadRejectsInvalidConfig(t *testing.T) {
	targetDir := setupInitializedInstance(t)

	path := filepath.Join(targetDir, configFilename)

	if err := os.WriteFile(
		path,
		[]byte(`this is not valid toml = [`),
		0600,
	); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	_, err := Load(targetDir)
	if err == nil {
		t.Fatal("expected Load to reject invalid config")
	}

	if !strings.Contains(err.Error(), "load Pocketry config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that Load rejects a directory with a missing required file
func TestLoadRejectsMissingRequiredFile(t *testing.T) {
	tests := []struct {
		name            string
		pathForInstance func(*Instance) string
		expectedError   string
	}{
		{
			name: "database",
			pathForInstance: func(inst *Instance) string {
				return inst.DatabasePath()
			},
			expectedError: "required file not found: database",
		},
		{
			name: "access token secret",
			pathForInstance: func(inst *Instance) string {
				return inst.AccessTokenSecretPath()
			},
			expectedError: "required file not found: access token secret",
		},
		{
			name: "OPAQUE key material",
			pathForInstance: func(inst *Instance) string {
				return inst.OpaqueKeyMaterialPath()
			},
			expectedError: "required file not found: OPAQUE server key material",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetDir := setupInitializedInstance(t)

			inst, err := Load(targetDir)
			if err != nil {
				t.Fatalf("initial Load returned error: %v", err)
			}

			path := tt.pathForInstance(inst)

			if err := os.Remove(path); err != nil {
				t.Fatalf("remove required file: %v", err)
			}

			_, err = Load(targetDir)
			if err == nil {
				t.Fatal("expected Load to reject missing required file")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf(
					"expected error containing %q, got %v",
					tt.expectedError,
					err,
				)
			}
		})
	}
}

// ensures that Load rejects a required file that is a directory
func TestLoadRejectsRequiredFileThatIsDirectory(t *testing.T) {
	tests := []struct {
		name            string
		pathForInstance func(*Instance) string
		expectedError   string
	}{
		{
			name: "database",
			pathForInstance: func(inst *Instance) string {
				return inst.DatabasePath()
			},
			expectedError: "database path is not a regular file",
		},
		{
			name: "access token secret",
			pathForInstance: func(inst *Instance) string {
				return inst.AccessTokenSecretPath()
			},
			expectedError: "access token secret path is not a regular file",
		},
		{
			name: "OPAQUE key material",
			pathForInstance: func(inst *Instance) string {
				return inst.OpaqueKeyMaterialPath()
			},
			expectedError: "OPAQUE server key material path is not a regular file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetDir := setupInitializedInstance(t)

			inst, err := Load(targetDir)
			if err != nil {
				t.Fatalf("initial Load returned error: %v", err)
			}

			path := tt.pathForInstance(inst)

			if err := os.Remove(path); err != nil {
				t.Fatalf("remove required file: %v", err)
			}

			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatalf("create directory at required path: %v", err)
			}

			_, err = Load(targetDir)
			if err == nil {
				t.Fatal(
					"expected Load to reject directory at required file path",
				)
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf(
					"expected error containing %q, got %v",
					tt.expectedError,
					err,
				)
			}
		})
	}
}
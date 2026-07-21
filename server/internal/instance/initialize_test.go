package instance

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CDX-1/pocketry/internal/db"
)

// Checks that Initialize creates the correct directory structure and files
func TestInitializeCreatesPocketryServer(t *testing.T) {
	parentDir := t.TempDir()
	targetDir := filepath.Join(parentDir, "server")

	inst, err := Initialize(targetDir)
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	expectedRootDir, err := filepath.Abs(targetDir)
	if err != nil {
		t.Fatalf("filepath.Abs returned error: %v", err)
	}

	if inst.RootDir != expectedRootDir {
		t.Fatalf(
			"expected root directory %q, got %q",
			expectedRootDir,
			inst.RootDir,
		)
	}

	expectedPaths := []string{
		filepath.Join(targetDir, "pocketry.instance"),
		filepath.Join(targetDir, "pocketry.toml"),
		filepath.Join(targetDir, "data"),
		filepath.Join(targetDir, "data", "pocketry.db"),
		filepath.Join(targetDir, "secrets"),
		filepath.Join(targetDir, "secrets", "access-token.key"),
		filepath.Join(targetDir, "secrets", "opaque.key"),
	}

	for _, expectedPath := range expectedPaths {
		if _, err := os.Stat(expectedPath); err != nil {
			t.Fatalf("expected %q to exist: %v", expectedPath, err)
		}
	}
}

// Checks that Initialize writes a valid marker file
func TestInitializeWritesValidMarker(t *testing.T) {
	targetDir := filepath.Join(t.TempDir(), "server")

	if _, err := Initialize(targetDir); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	data, err := os.ReadFile(
		filepath.Join(targetDir, "pocketry.instance"),
	)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}

	var marker Marker
	if err := json.Unmarshal(data, &marker); err != nil {
		t.Fatalf("decode marker: %v", err)
	}

	if marker.Format != formatName {
		t.Fatalf(
			"expected marker format: %q, got %q",
			formatName,
			marker.Format,
		)
	}

	if marker.FormatVersion != formatVersion {
		t.Fatalf(
			"expected format version: %d, got %d",
			formatVersion,
			marker.FormatVersion,
		)
	}

	if marker.InstanceID == "" {
		t.Fatalf("expected instance ID to be non-empty")
	}

	if marker.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be non-empty")
	}
}

// Checks that Initialize writes a valid access token secret
func TestInitializeWritesValidAccessTokenSecret(t *testing.T) {
	targetDir := filepath.Join(t.TempDir(), "server")

	if _, err := Initialize(targetDir); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	data, err := os.ReadFile(
		filepath.Join(targetDir, "secrets", "access-token.key"),
	)
	if err != nil {
		t.Fatalf("read access token secret: %v", err)
	}

	lines := strings.Split(
		strings.TrimSpace(string(data)),
		"\n",
	)

	if len(lines) != 2 {
		t.Fatalf(
			"expected secret file to contain two lines, got %d",
			len(lines),
		)
	}

	if lines[0] != accessTokenSecretHeader {
		t.Fatalf(
			"expected header %q, got %q",
			accessTokenSecretHeader,
			lines[0],
		)
	}

	secret, err := base64.RawURLEncoding.DecodeString(lines[1])
	if err != nil {
		t.Fatalf("decode access token secret: %v", err)
	}

	if len(secret) < 32 {
		t.Fatalf(
			"expected at least 32 byte secret, got %d",
			len(secret),
		)
	}
}

// Checks that Initialize creates a usable database
func TestInitializeCreatesUsableDatabase(t *testing.T) {
	targetDir := filepath.Join(t.TempDir(), "server")

	if _, err := Initialize(targetDir); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	store, err := db.Open(
		filepath.Join(targetDir, "data", "pocketry.db"),
	)
	if err != nil {
		t.Fatalf("open initialized database: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})

	if err := store.DB.PingContext(t.Context()); err != nil {
		t.Fatalf("ping initialized database: %v", err)
	}

	if store.Q == nil {
		t.Fatal("expected initialize database queries")
	}
}

// Checks that Initialize rejects an existing target directory
func TestInitializeRejectsExistingTarget(t *testing.T) {
	targetDir := filepath.Join(t.TempDir(), "server")

	if err := os.Mkdir(targetDir, 0700); err != nil {
		t.Fatalf("create existing target: %v", err)
	}

	_, err := Initialize(targetDir)
	if err == nil {
		t.Fatal("expected Initialize to reject existing target")
	}

	if !strings.Contains(err.Error(), "target already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}
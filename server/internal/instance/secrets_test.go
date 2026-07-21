package instance

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSecretTestInstance(t *testing.T) *Instance {
	t.Helper()

	rootDir := t.TempDir()

	if err := os.MkdirAll(
		filepath.Join(rootDir, "secrets"), 0700,
	); err != nil {
		t.Fatalf("create secrets directory: %v", err)
	}

	return &Instance{
		RootDir: rootDir,
	}
}

// ensures that LoadAccessTokenSecret correctly loads the access token secret
func TestLoadAccessTokenSecret(t *testing.T) {
	inst := setupSecretTestInstance(t)

	want := []byte("0123456789abcdef0123456789abcdef")
	encoded := base64.RawURLEncoding.EncodeToString(want)

	content := []byte(
		accessTokenSecretHeader + "\n" +
			encoded + "\n",
	)

	if err := os.WriteFile(
		inst.AccessTokenSecretPath(),
		content,
		0600,
	); err != nil {
		t.Fatalf("write access token secret: %v", err)
	}

	got, err := inst.LoadAccessTokenSecret()
	if err != nil {
		t.Fatalf("LoadAccessTokenSecret returned error: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"unexpected secret: got %q, want %q",
			got,
			want,
		)
	}
}

// ensures that LoadAccessTokenSecret rejects files with unexpected formats
func TestLoadAccessTokenSecretRejectsInvalidFiles(t *testing.T) {
	validSecret := base64.RawURLEncoding.EncodeToString(
		[]byte("0123456789abcdef0123456789abcdef"),
	)

	tests := []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name:          "empty file",
			content:       "",
			expectedError: "invalid access token secret file format",
		},
		{
			name:          "header only",
			content:       accessTokenSecretHeader + "\n",
			expectedError: "invalid access token secret file format",
		},
		{
			name: "too many lines",
			content: accessTokenSecretHeader + "\n" +
				validSecret + "\nextra\n",
			expectedError: "invalid access token secret file format",
		},
		{
			name: "wrong header",
			content: "WRONG-HEADER\n" +
				validSecret + "\n",
			expectedError: "unsupported access token secret format",
		},
		{
			name: "invalid base64",
			content: accessTokenSecretHeader + "\n" +
				"not-valid-base64***\n",
			expectedError: "decode access token secret",
		},
		{
			name: "secret too short",
			content: accessTokenSecretHeader + "\n" +
				base64.RawURLEncoding.EncodeToString(
					[]byte("too-short"),
				) + "\n",
			expectedError: "must contain at least 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := setupSecretTestInstance(t)

			if err := os.WriteFile(
				inst.AccessTokenSecretPath(),
				[]byte(tt.content),
				0600,
			); err != nil {
				t.Fatalf("write access token secret: %v", err)
			}

			_, err := inst.LoadAccessTokenSecret()
			if err == nil {
				t.Fatal("expected LoadAccessTokenSecret to return an error")
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

// ensures that LoadAccessTokenSecret rejects a missing secret file
func TestLoadAccessTokenSecretRejectsMissingFile(t *testing.T) {
	inst := setupSecretTestInstance(t)

	_, err := inst.LoadAccessTokenSecret()
	if err == nil {
		t.Fatal("expected missing access token secret error")
	}

	if !strings.Contains(err.Error(), "read access token secret") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that LoadOpaqueServerKeyMaterial successfully loads the OPAQUE key material
func TestLoadOpaqueServerKeyMaterial(t *testing.T) {
	inst := setupSecretTestInstance(t)

	want := []byte("encoded-opaque-server-key-material")

	if err := os.WriteFile(
		inst.OpaqueKeyMaterialPath(),
		want,
		0600,
	); err != nil {
		t.Fatalf("write OPAQUE key material: %v", err)
	}

	got, err := inst.LoadOpaqueServerKeyMaterial()
	if err != nil {
		t.Fatalf(
			"LoadOpaqueServerKeyMaterial returned error: %v",
			err,
		)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"unexpected OPAQUE key material: got %q, want %q",
			got,
			want,
		)
	}
}

// ensures that LoadOpaqueServerKeyMaterial rejects a missing key material file
func TestLoadOpaqueServerKeyMaterialRejectsMissingFile(t *testing.T) {
	inst := setupSecretTestInstance(t)

	_, err := inst.LoadOpaqueServerKeyMaterial()
	if err == nil {
		t.Fatal("expected missing OPAQUE key material error")
	}

	if !strings.Contains(
		err.Error(),
		"read OPAQUE server key material",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ensures that LoadOpaqueServerKeyMaterial rejects an empty key material file
func TestLoadOpaqueServerKeyMaterialRejectsEmptyFile(t *testing.T) {
	inst := setupSecretTestInstance(t)

	if err := os.WriteFile(
		inst.OpaqueKeyMaterialPath(),
		nil,
		0600,
	); err != nil {
		t.Fatalf("write empty OPAQUE key material: %v", err)
	}

	_, err := inst.LoadOpaqueServerKeyMaterial()
	if err == nil {
		t.Fatal("expected empty OPAQUE key material error")
	}

	if !strings.Contains(
		err.Error(),
		"OPAQUE server key material is empty",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}
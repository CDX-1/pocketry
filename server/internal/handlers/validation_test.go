package handlers

import "testing"

// checks that validateAuthRequest accepts valid auth requests
func TestValidateAuthRequestAcceptsValidRequest(t *testing.T) {
	req := AuthRequest{
		Username: "username",
		Password: "StrongPassword123!",
	}

	if err := validateAuthRequest(&req); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}
}

// checks that validateAuthRequest trims leading/trailing whitespace from username
func TestValidateAuthRequestTrimsUsername(t *testing.T) {
	req := AuthRequest{
		Username: "  testuser  ",
		Password: "StrongPassword123!",
	}

	if err := validateAuthRequest(&req); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	if req.Username != "testuser" {
		t.Fatalf("expected trimmed username, got %q", req.Username)
	}
}

// checks that validateAuthRequest rejects empty username
func TestValidateAuthRequestRejectsEmptyUsername(t *testing.T) {
	req := AuthRequest{
		Username: "",
		Password: "StrongPassword123!",
	}

	if err := validateAuthRequest(&req); err == nil {
		t.Fatal("expected error for empty username")
	}
}

// checks that validateAuthRequest rejects short password
func TestValidateAuthRequestRejectsShortPassword(t *testing.T) {
	req := AuthRequest{
		Username: "testuser",
		Password: "short",
	}

	if err := validateAuthRequest(&req); err == nil {
		t.Fatal("expected error for short password")
	}
}

// checks that validateVaultRequest accepts valid vault requests
func TestValidateVaultRequestAcceptsValidBlob(t *testing.T) {
	req := VaultRequest{
		EncryptedBlob: "fake-encrypted-vault-data",
	}

	if err := validateVaultRequest(&req); err != nil {
		t.Fatalf("expected valid vault request, got error: %v", err)
	}
}

// checks that validateVaultRequest rejects empty blob
func TestValidateVaultRequestRejectsEmptyBlob(t *testing.T) {
	req := VaultRequest{
		EncryptedBlob: "",
	}

	if err := validateVaultRequest(&req); err == nil {
		t.Fatal("expected error for empty encrypted_blob")
	}
}
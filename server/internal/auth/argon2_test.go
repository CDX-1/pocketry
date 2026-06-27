package auth

import "testing"

// checks that verifying password hash returns true with correct password
func TestHashPasswordAndVerifyPassword(t *testing.T) {
	password := "StrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatal("expected hash to not be empty")
	}

	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}

	if !match {
		t.Fatal("expected password to match hash")
	}
}

// checks that verifying password hash returns false with wrong password
func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("StrongPassword123!")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	match, err := VerifyPassword("WrongPassword123!", hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}

	if match {
		t.Fatal("expected password to not match hash")
	}
}

// checks that verifying password hash returns false with invalid hash format
func TestVerifyPasswordWithInvalidHash(t *testing.T) {
	match, err := VerifyPassword("StrongPassword123!", "invalid-hash")
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}

	if match {
		t.Fatal("expected password to not match hash")
	}
}
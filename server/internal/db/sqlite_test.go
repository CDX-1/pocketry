package db

import (
	"path/filepath"
	"testing"
)

// checks that InitDB initializes the database schema and queries (Q)
func TestInitDBCreatesTables(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}

	if DB == nil {
		t.Fatal("expected DB to be initialized")
	}

	if Q == nil {
		t.Fatal("expected Q to be initialized")
	}

	t.Cleanup(func() {
		_ = DB.Close()
	})
}

// checks that CreateUser creates a new user and GetUserByUsername retrieves it
func TestCreateAndGetUser(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}

	t.Cleanup(func() {
		_ = DB.Close()
	})

	err := Q.CreateUser(
		t.Context(),
		CreateUserParams{
			Username: 		"username",
			PasswordHash: 	"password-hash",
		},
	)
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	user, err := Q.GetUserByUsername(t.Context(), "username")
	if err != nil {
		t.Fatalf("GetUserByUsername returned error: %v", err)
	}

	if user.ID == 0 {
		t.Fatal("expected user ID to be set")
	}

	if user.PasswordHash != "password-hash" {
		t.Fatalf("expected password hash: %q", user.PasswordHash)
	}
}
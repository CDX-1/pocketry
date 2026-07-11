package db

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}

	t.Cleanup(func() {
		if DB != nil {
			if err := DB.Close(); err != nil {
				t.Errorf("close test database: %v", err)
			}
		}

		DB = nil
		Q = nil
	})
}

// tests database & query initialization & pinging
func TestInitDBInitializeDatabaseAndQueries(t *testing.T) {
	setupTestDB(t)

	if DB == nil {
		t.Fatal("expected DB to be initialized")
	}

	if Q == nil {
		t.Fatal("expected Q to be initialized")
	}

	if err := DB.PingContext(t.Context()); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}
}

// ensures database creates all tables
func TestInitDBCreatesExpectedTables(t *testing.T) {
	setupTestDB(t)

	expectedTables := []string{
		"users",
		"pending_registrations",
		"pending_logins",
		"vaults",
	}

	for _, tableName := range expectedTables {
		t.Run(tableName, func(t *testing.T) {
			var actualName string

			err := DB.QueryRowContext(
				t.Context(),
				`
					SELECT name
					FROM sqlite_master
					WHERE type = 'table'
					AND name = ?
				`,
				tableName,
			).Scan(&actualName)

			if err != nil {
				t.Fatalf(
					"expected table %q to exist: %v",
					tableName,
					err,
				)
			}

			if actualName != tableName {
				t.Fatalf(
					"expected table name %q, got %q",
					tableName,
					actualName,
				)
			}
		})
	}
}

// ensures database has foreign keys enabled
func TestInitDBEnablesForeignKeys(t *testing.T) {
	setupTestDB(t)

	var enabled int

	if err := DB.QueryRowContext(
		t.Context(),
		"PRAGMA foreign_keys;",
	).Scan(&enabled); err != nil {
		t.Fatalf("read foreign_keys pragma: %v", err)
	}

	if enabled != 1 {
		t.Fatalf("expected foreign_keys to be enabled, got %d", enabled)
	}
}

// ensures database has WAL mode enabled
func TestInitDBEnablesWALMode(t *testing.T) {
	setupTestDB(t)

	var journalMode string

	if err := DB.QueryRowContext(
		t.Context(),
		"PRAGMA journal_mode;",
	).Scan(&journalMode); err != nil {
		t.Fatalf("read journal_mode pragma: %v", err)
	}

	if journalMode != "wal" {
		t.Fatalf("expected journal mode %q, got %q", "wal", journalMode)
	}
}

// ensures database set busy timeout correctly
func TestInitDBSetsBusyTimeout(t *testing.T) {
	setupTestDB(t)

	var timeout int

	if err := DB.QueryRowContext(
		t.Context(),
		"PRAGMA busy_timeout;",
	).Scan(&timeout); err != nil {
		t.Fatalf("read busy_timeout pragma: %v", err)
	}

	if timeout != 5000 {
		t.Fatalf("expected busy timeout 5000ms, got %d", timeout)
	}
}

// verifies that the database can create and retrieve users using normalized usernames
func TestCreateAndGetUserByNormalizedUsername(t *testing.T) {
	setupTestDB(t)

	registrationRecord := []byte("opaque-registration-record")

	err := Q.CreateUser(
		t.Context(),
		CreateUserParams{
			Username:               "Username",
			UsernameNormalized:     "username",
			OpaqueRegistrationRecord: registrationRecord,
		},
	)
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	user, err := Q.GetUserByUsernameNormalized(
		t.Context(),
		"username",
	)
	if err != nil {
		t.Fatalf(
			"GetUserByUsernameNormalized returned error: %v",
			err,
		)
	}

	if user.ID <= 0 {
		t.Fatalf("expected positive user ID, got %d", user.ID)
	}

	if user.Username != "Username" {
		t.Fatalf(
			"expected username %q, got %q",
			"Username",
			user.Username,
		)
	}

	if user.UsernameNormalized != "username" {
		t.Fatalf(
			"expected username %q, got %q",
			"username",
			user.UsernameNormalized,
		)
	}

	if !bytes.Equal(
		user.OpaqueRegistrationRecord,
		registrationRecord,
	) {
		t.Fatalf(
			"unexpected OPAQUE registration record: got %q, want %q",
			user.OpaqueRegistrationRecord,
			registrationRecord,
		)
	}

	if user.CryptoPolicyVersion != 1 {
		t.Fatalf(
			"expected crypto policy version 1, got %d",
			user.CryptoPolicyVersion,
		)
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be non-empty")
	}

	if user.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be non-empty")
	}
}

// verifies that the database can retrieve users by ID after creation
func TestGetUserByID(t *testing.T) {
	setupTestDB(t)

	err := Q.CreateUser(
		t.Context(),
		CreateUserParams{
			Username: 				  "Username",
			UsernameNormalized: 	  "username",
			OpaqueRegistrationRecord: []byte("opaque-registration-record"),
		},
	)
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	createdUser, err := Q.GetUserByUsernameNormalized(
		t.Context(),
		"username",
	)
	if err != nil {
		t.Fatalf(
			"GetUserByNormalized returned error: %v",
			err,
		)
	}

	user, err := Q.GetUserByID(
		t.Context(),
		createdUser.ID,
	)
	if err != nil {
		t.Fatalf("GetUserByID returned error: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			createdUser.ID,
			user.ID,
		)
	}

	if user.Username != "Username" {
		t.Fatalf(
			"expected username %q, got %q",
			"Username",
			user.Username,
		)
	}
}

// verifies that the database can check for username existence
func TestUsernameExists(t *testing.T) {
	setupTestDB(t)

	exists, err := Q.UsernameExists(
		t.Context(),
		"username",
	)
	if err != nil {
		t.Fatalf("UsernameExists returned error: %v", err)
	}

	if exists {
		t.Fatal("expected username not to exist before creation")
	}

	err = Q.CreateUser(
		t.Context(),
		CreateUserParams{
			Username:                 "Username",
			UsernameNormalized:       "username",
			OpaqueRegistrationRecord: []byte("registration-record"),
		},
	)
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	exists, err = Q.UsernameExists(
		t.Context(),
		"username",
	)
	if err != nil {
		t.Fatalf("UsernameExists returned error: %v", err)
	}

	if !exists {
		t.Fatal("expected username to exist after creation")
	}
}

// verifies that the database rejects duplicate normalized usernames
func TestCreateUserRejectsDuplicateNormalizedUsername(t *testing.T) {
	setupTestDB(t)

	first := CreateUserParams{
		Username:                 "Username",
		UsernameNormalized:       "username",
		OpaqueRegistrationRecord: []byte("first-record"),
	}

	if err := Q.CreateUser(t.Context(), first); err != nil {
		t.Fatalf("first CreateUser returned error: %v", err)
	}

	second := CreateUserParams{
		Username:                 "USERNAME",
		UsernameNormalized:       "username",
		OpaqueRegistrationRecord: []byte("second-record"),
	}

	err := Q.CreateUser(t.Context(), second)
	if err == nil {
		t.Fatal("expected duplicate normalized username error")
	}
}

// verifies that the database returns sql.ErrNoRows for missing users
func TestGetMissingUserReturnsNoRows(t *testing.T) {
	setupTestDB(t)

	_, err := Q.GetUserByUsernameNormalized(
		t.Context(),
		"missinguser",
	)

	if err == nil {
		t.Fatal("expected error for missing user")
	}

	if err != sql.ErrNoRows {
		t.Fatalf(
			"expected sql.ErrNoRows, got %T: %v",
			err,
			err,
		)
	}
}
package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	DB *sql.DB
	Q  *Queries
}

func Open(filepath string) (*Store, error) {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := configure(database); err != nil {
		_ = database.Close()
		return nil, err
	}

	if _, err := database.Exec(schemaSQL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("apply database schema: %w", err)
	}

	return &Store{
		DB: database,
		Q:  New(database),
	}, nil
}

func configure(database *sql.DB) error {
	statements := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
	}

	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			return fmt.Errorf("execute %q: %w", statement, err)
		}
	}

	return nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}
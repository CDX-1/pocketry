package db

import (
	"database/sql"
	_ "embed"
	_ "github.com/ncruces/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB
var Q *Queries

func InitDB(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		return err
	}

	if _, err = DB.Exec(schemaSQL); err != nil {
		return err
	}

	Q = New(DB)
	return nil
}
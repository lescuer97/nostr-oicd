package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectDB opens the sqlite database and returns a *sql.DB handle.
// It uses a local file named nostr_oidc.db in the project root.
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./nostr_oidc.db")
	if err != nil {
		return nil, err
	}
	// Set reasonable limits
	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Ensure the database is reachable
	if err := db.Ping(); err != nil {
		log.Printf("db ping failed: %v", err)
		return nil, err
	}
	return db, nil
}

package db

import (
	"database/sql"
	"os"
)

func RunMigrations(db *sql.DB) error {
	// Phase 1: users table
	if !tableExists(db, "users") {
		if err := execSQLFromFile(db, "./migrations/001_create_users_table.sql"); err != nil {
			return err
		}
	}
	// Phase 1: sessions table
	if !tableExists(db, "sessions") {
		if err := execSQLFromFile(db, "./migrations/002_create_sessions_table.sql"); err != nil {
			return err
		}
	}
	// Phase 2: auth_challenges table
	if !tableExists(db, "auth_challenges") {
		if err := execSQLFromFile(db, "./migrations/003_create_auth_challenges.sql"); err != nil {
			return err
		}
	}
	return nil
}

func tableExists(db *sql.DB, name string) bool {
	var count int
	row := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", name)
	if err := row.Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func execSQLFromFile(db *sql.DB, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(b))
	return err
}

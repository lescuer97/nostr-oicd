package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func RunMigrations(db *sql.DB) error {
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	// Phase 1: users table
	if !tableExists(db, "users") {
		if err := execSQLFromFile(db, filepath.Join(migrationsDir, "001_create_users_table.sql")); err != nil {
			return err
		}
	}
	// Phase 1: sessions table
	if !tableExists(db, "sessions") {
		if err := execSQLFromFile(db, filepath.Join(migrationsDir, "002_create_sessions_table.sql")); err != nil {
			return err
		}
	}
	// Phase 2: auth_challenges table
	if !tableExists(db, "auth_challenges") {
		if err := execSQLFromFile(db, filepath.Join(migrationsDir, "003_create_auth_challenges.sql")); err != nil {
			return err
		}
	}
	return nil
}

func findMigrationsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(wd, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", fmt.Errorf("migrations directory not found")
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

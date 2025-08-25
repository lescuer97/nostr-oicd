package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// PRAGMA for WAL mode and busy timeout
	_, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Printf("warning: failed to set pragmas: %v", err)
	}
	return db, nil
}

func RunMigrations(db *sql.DB, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	for _, f := range files {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		queries := string(b)
		_, err = db.Exec(queries)
		if err != nil {
			return fmt.Errorf("migration %s failed: %w", f, err)
		}
		time.Sleep(20 * time.Millisecond) // tiny pause between migrations
	}
	return nil
}

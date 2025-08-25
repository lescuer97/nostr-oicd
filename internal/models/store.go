package models

import (
	"context"
	"database/sql"
	"time"
)

// EnsureUser finds a user by public key or creates one atomically using a transaction.
func EnsureUser(ctx context.Context, db *sql.DB, pubkey string) (int64, error) {
	// Start a transaction so the find-or-create is atomic.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var id int64
	row := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE public_key = ?`, pubkey)
	switch err := row.Scan(&id); err {
	case nil:
		// found
		if err := tx.Commit(); err != nil {
			return 0, err
		}
		return id, nil
	case sql.ErrNoRows:
		// insert
		res, err := tx.ExecContext(ctx, `INSERT INTO users (public_key, is_admin, created_at, updated_at) VALUES (?, false, ?, ?)`, pubkey, time.Now().Unix(), time.Now().Unix())
		if err != nil {
			return 0, err
		}
		last, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		if err := tx.Commit(); err != nil {
			return 0, err
		}
		return last, nil
	default:
		return 0, err
	}
}

// CreateSession inserts a session row within a transaction and returns its id.
func CreateSession(ctx context.Context, db *sql.DB, userID int64, tokenHash string, expiresAt time.Time) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx, `INSERT INTO sessions (user_id, token_hash, created_at, expires_at, active) VALUES (?, ?, ?, ?, 1)`, userID, tokenHash, time.Now().Unix(), expiresAt.Unix())
	if err != nil {
		return 0, err
	}
	last, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return last, nil
}

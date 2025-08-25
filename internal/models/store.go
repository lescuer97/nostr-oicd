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
	tlast, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return tlast, nil
}

// GetSessionByHash looks up a session row by token_hash and checks active/expiry.
func GetSessionByHash(ctx context.Context, db *sql.DB, tokenHash string) (*Session, error) {
	row := db.QueryRowContext(ctx, `SELECT id, user_id, token_hash, created_at, expires_at, active FROM sessions WHERE token_hash = ? LIMIT 1`, tokenHash)
	var s Session
	var createdAtUnix, expiresAtUnix int64
	if err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &createdAtUnix, &expiresAtUnix, &s.Active); err != nil {
		return nil, err
	}
	s.CreatedAt = time.Unix(createdAtUnix, 0)
	s.ExpiresAt = time.Unix(expiresAtUnix, 0)
	// check active and expiry
	if !s.Active || time.Now().After(s.ExpiresAt) {
		return nil, sql.ErrNoRows
	}
	return &s, nil
}

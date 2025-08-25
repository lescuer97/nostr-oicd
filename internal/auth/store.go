package auth

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

var (
	mu         sync.Mutex
	challenges = make(map[string]time.Time)
	// challenge TTL
	ttl = 5 * time.Minute
)

func SaveChallenge(ch string) {
	mu.Lock()
	defer mu.Unlock()
	challenges[ch] = time.Now()
}

func ValidateAndDeleteChallenge(ch string) bool {
	mu.Lock()
	defer mu.Unlock()
	t, ok := challenges[ch]
	if !ok {
		return false
	}
	if time.Since(t) > ttl {
		delete(challenges, ch)
		return false
	}
	delete(challenges, ch)
	return true
}

// DeactivateSessionByHash sets active = false for the session with the given token_hash
func DeactivateSessionByHash(ctx context.Context, db *sql.DB, tokenHash string) error {
	res, err := db.ExecContext(ctx, `UPDATE sessions SET active = 0 WHERE token_hash = ?`, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return nil
	}
	return nil
}

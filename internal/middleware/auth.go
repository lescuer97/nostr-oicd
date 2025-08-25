package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/lescuer97/nostr-oicd/internal/config"
	"github.com/lescuer97/nostr-oicd/internal/models"
)

type contextKey string

const ContextUserKey = contextKey("user")

// AuthMiddleware validates the session cookie token by computing HMAC(token)
// and looking up the session in the DB. If valid, it loads the user and stores
// it in the request context. Otherwise it returns 401.
func AuthMiddleware(cfg *config.Config, db *sql.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := cookie.Value
			// compute hmac
			signKey := []byte(cfg.SessionSigningKey)
			if len(signKey) == 0 {
				signKey = []byte(cfg.JWTSecret)
			}
			h := hmac.New(sha256.New, signKey)
			h.Write([]byte(token))
			tokenHash := hex.EncodeToString(h.Sum(nil))

			// find session
			sess, err := models.GetSessionByHash(r.Context(), db, tokenHash)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// load user
			row := db.QueryRowContext(r.Context(), `SELECT id, public_key, is_admin, created_at, updated_at FROM users WHERE id = ?`, sess.UserID)
			var u models.User
			var createdAtUnix, updatedAtUnix int64
			if err := row.Scan(&u.ID, &u.PublicKey, &u.IsAdmin, &createdAtUnix, &updatedAtUnix); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			u.CreatedAt = time.Unix(createdAtUnix, 0)
			u.UpdatedAt = time.Unix(updatedAtUnix, 0)

			// attach user to context
			ctx := context.WithValue(r.Context(), ContextUserKey, &u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

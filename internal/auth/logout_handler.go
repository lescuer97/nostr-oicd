package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"

	"github.com/lescuer97/nostr-oicd/internal/config"
)

// LogoutHandler invalidates the current session token (sets active=false in DB) and clears the cookie.
func LogoutHandler(cfg *config.Config, db *sql.DB, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cfg.CookieName)
	if err != nil {
		// no cookie — nothing to do
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	token := cookie.Value
	// compute HMAC like in middleware
	signKey := []byte(cfg.SessionSigningKey)
	if len(signKey) == 0 {
		signKey = []byte(cfg.JWTSecret)
	}
	h := hmac.New(sha256.New, signKey)
	h.Write([]byte(token))
	tokenHash := hex.EncodeToString(h.Sum(nil))

	// deactivate session
	if err := DeactivateSessionByHash(r.Context(), db, tokenHash); err != nil {
		// If failing to deactivate, still clear cookie and redirect
		// (log server-side in real app)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		MaxAge:   -1,
	})

	// redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

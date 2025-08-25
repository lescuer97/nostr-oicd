package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lescuer97/nostr-oicd/internal/config"
	"github.com/lescuer97/nostr-oicd/internal/models"
	"github.com/lescuer97/nostr-oicd/templates/fragments"
	"github.com/nbd-wtf/go-nostr"
)

// generateRandomToken returns a hex token of nBytes length (2*n hex chars)
func generateRandomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hmacHash returns hex-encoded HMAC-SHA256 of data using key
func hmacHash(key []byte, data string) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// renderLoginError renders the LoginError fragment with the given message.
func renderLoginError(ctx context.Context, w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = fragments.LoginError(msg).Render(ctx, w)
}

// LoginHandler handles signed nostr event login. It receives the app config and DB via closure
func LoginHandler(cfg *config.Config, db *sql.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Expect signed_event in POST form
	if err := r.ParseForm(); err != nil {
		renderLoginError(ctx, w, "invalid request")
		return
	}
	signed := r.FormValue("signed_event")
	if signed == "" {
		renderLoginError(ctx, w, "missing signed_event")
		return
	}
	// Parse signed event JSON
	var ev nostr.Event
	if err := json.Unmarshal([]byte(signed), &ev); err != nil {
		renderLoginError(ctx, w, "invalid event")
		return
	}
	// Validate signature using event method
	ok, err := ev.CheckSignature()
	if err != nil {
		renderLoginError(ctx, w, "invalid signature")
		return
	}
	if !ok {
		renderLoginError(ctx, w, "signature verification failed")
		return
	}
	// The event content should be the challenge
	challenge := ev.Content
	if !ValidateAndDeleteChallenge(challenge) {
		renderLoginError(ctx, w, "invalid or expired challenge")
		return
	}
	// Ensure user exists
	userID, err := models.EnsureUser(ctx, db, ev.PubKey)
	if err != nil {
		renderLoginError(ctx, w, "failed to create user")
		return
	}

	// Generate an opaque session token (random) and store its HMAC in DB
	token, err := generateRandomToken(32) // 32 bytes -> 64 hex chars
	if err != nil {
		renderLoginError(ctx, w, "failed to generate token")
		return
	}
	// Use SESSION_SIGNING_KEY if provided, else fallback to JWT secret
	signKey := []byte(cfg.SessionSigningKey)
	if len(signKey) == 0 {
		signKey = []byte(cfg.JWTSecret)
	}
	hash := hmacHash(signKey, token)

	expiresAt := time.Now().Add(15 * time.Minute)
	if _, err := models.CreateSession(ctx, db, userID, hash, expiresAt); err != nil {
		renderLoginError(ctx, w, "failed to create session")
		return
	}

	// Set cookie to the opaque token value
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	// Render login success fragment
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := fragments.LoginSuccess().Render(ctx, w); err != nil {
		// As a fallback, write plain text
		http.Error(w, "failed to render fragment", http.StatusInternalServerError)
	}
}

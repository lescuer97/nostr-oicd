package auth

import (
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

// LoginHandler handles signed nostr event login. It receives the app config and DB via closure
func LoginHandler(cfg *config.Config, db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Expect signed_event in POST form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	signed := r.FormValue("signed_event")
	if signed == "" {
		http.Error(w, "missing signed_event", http.StatusBadRequest)
		return
	}
	// Parse signed event JSON
	var ev nostr.Event
	if err := json.Unmarshal([]byte(signed), &ev); err != nil {
		http.Error(w, "invalid event", http.StatusBadRequest)
		return
	}
	// Validate signature using event method
	ok, err := ev.CheckSignature()
	if err != nil || !ok {
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}
	// The event content should be the challenge
	challenge := ev.Content
	if !ValidateAndDeleteChallenge(challenge) {
		http.Error(w, "invalid or expired challenge", http.StatusUnauthorized)
		return
	}
	// Ensure user exists
	userID, err := models.EnsureUser(r.Context(), db, ev.PubKey)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate an opaque session token (random) and store its HMAC in DB
	token, err := generateRandomToken(32) // 32 bytes -> 64 hex chars
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	// Use SESSION_SIGNING_KEY if provided, else fallback to JWT secret
	signKey := []byte(cfg.SessionSigningKey)
	if len(signKey) == 0 {
		signKey = []byte(cfg.JWTSecret)
	}
	hash := hmacHash(signKey, token)

	expiresAt := time.Now().Add(15 * time.Minute)
	if _, err := models.CreateSession(r.Context(), db, userID, hash, expiresAt); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
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

	// Return success fragment (simple text for now)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<div class=\"p-4 bg-green-100\">Logged in</div>"))
}

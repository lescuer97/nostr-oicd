package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lescuer97/nostr-oicd/internal/config"
	"github.com/lescuer97/nostr-oicd/internal/models"
	"github.com/nbd-wtf/go-nostr"
)

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
	if err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	if !ok {
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
	// Create JWT and session
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": ev.PubKey,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}
	// Store session in DB
	hash := fmt.Sprintf("hash_%s", tokenStr) // placeholder; use real hashing
	expiresAt := time.Now().Add(15 * time.Minute)
	if _, err := models.CreateSession(r.Context(), db, userID, hash, expiresAt); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CookieName,
		Value:    tokenStr,
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

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
	"github.com/lescuer97/nostr-oicd/internal/ui"
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

// renderLoginError renders an inline script that shows a toast popup with the provided message.
func renderLoginError(ctx context.Context, w http.ResponseWriter, msg string) {
	// Prefer sending an HX-Trigger header so HTMX can trigger a client-side notify event
	// Prefer returning an OOB fragment to update the stable snackbar via hx-swap-oob
	// Use helper to render snackbar and target #htmx-snackbar
	_ = ui.RenderSnackbar(ctx, w, msg, "error", "3s")
	return
}

// extractChallengeFromEvent returns the challenge string present either in the event content
// or in a tag of the form ["challenge", "<value>"]. If none found, returns empty string.
func extractChallengeFromEvent(ev nostr.Event) string {
	if ev.Content != "" {
		return ev.Content
	}
	for _, t := range ev.Tags {
		if len(t) >= 2 && t[0] == "challenge" {
			return t[1]
		}
	}
	return ""
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
	// Ensure the event is the expected kind
	if ev.Kind != 2222 {
		renderLoginError(ctx, w, "unexpected event kind")
		return
	}
	// Extract challenge from content or tags
	challenge := extractChallengeFromEvent(ev)
	if challenge == "" {
		renderLoginError(ctx, w, "missing challenge in event")
		return
	}
	if !ValidateAndDeleteChallenge(challenge) {
		renderLoginError(ctx, w, "invalid or expired challenge")
		return
	}
	// Ensure user exists (only pre-registered users are allowed to login)
	userID, err := models.GetUserByPubKey(ctx, db, ev.PubKey)
	if err != nil {
		if err == sql.ErrNoRows {
			renderLoginError(ctx, w, "Your key is not authorized. Contact an admin.")
			return
		}
		renderLoginError(ctx, w, "failed to lookup user")
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

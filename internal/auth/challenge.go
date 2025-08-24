package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

func ChallengeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pub := r.URL.Query().Get("public_key_hex")
		if pub == "" {
			http.Error(w, "missing public_key_hex", http.StatusBadRequest)
			return
		}
		ch, err := GenerateChallenge()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		expiresAt := time.Now().Add(10 * time.Minute)
		if err := SaveChallenge(db, pub, ch, expiresAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{"challenge": ch, "expires_at": expiresAt.Format(time.RFC3339)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func SaveChallenge(db *sql.DB, publicKeyHex, challenge string, expiresAt time.Time) error {
	_, err := db.Exec("INSERT INTO auth_challenges (public_key_hex, challenge, created_at, expires_at) VALUES (?, ?, CURRENT_TIMESTAMP, ?)", publicKeyHex, challenge, expiresAt)
	return err
}

func GenerateChallenge() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

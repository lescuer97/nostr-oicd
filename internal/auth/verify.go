package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type VerifyRequest struct {
	PublicKeyHex string `json:"public_key_hex"`
	Challenge    string `json:"challenge"`
	Signature    string `json:"signature"`
	Client       string `json:"client,omitempty"`
}

func VerifyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.PublicKeyHex) == "" || strings.TrimSpace(req.Challenge) == "" || strings.TrimSpace(req.Signature) == "" {
			http.Error(w, "missing fields", http.StatusBadRequest)
			return
		}

		// Check challenge expiry
		var expiryStr string
		if err := db.QueryRow("SELECT expires_at FROM auth_challenges WHERE public_key_hex = ? AND challenge = ?", req.PublicKeyHex, req.Challenge).Scan(&expiryStr); err != nil {
			http.Error(w, "invalid challenge", http.StatusUnauthorized)
			return
		}
		expiresAt, err := time.Parse(time.RFC3339, expiryStr)
		if err != nil || time.Now().After(expiresAt) {
			http.Error(w, "challenge expired", http.StatusUnauthorized)
			return
		}

		ok, verErr := VerifySignature(req.PublicKeyHex, req.Challenge, req.Signature)
		if verErr != nil || !ok {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// Create or fetch user by public_key_hex
		userID := 0
		if err := db.QueryRow("SELECT id FROM users WHERE public_key_hex = ?", req.PublicKeyHex).Scan(&userID); err != nil {
			if err == sql.ErrNoRows {
				res, insErr := db.Exec("INSERT INTO users (public_key_hex, active) VALUES (?, ?)", req.PublicKeyHex, true)
				if insErr != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				if id64, _ := res.LastInsertId(); id64 != 0 {
					userID = int(id64)
				}
			} else {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		// Create session
		sidBytes := make([]byte, 32)
		if _, err := rand.Read(sidBytes); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		sessionID := hex.EncodeToString(sidBytes)
		expires := time.Now().Add(30 * 24 * time.Hour)

		if _, err := db.Exec("INSERT INTO sessions (session_id, user_id, client, active, expiry_timestamp) VALUES (?, ?, ?, ?, ?)", sessionID, userID, req.Client, true, expires); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "nostr_session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"expires_at": expires.Format(time.RFC3339),
		})
	}
}

func VerifySignature(pubKeyHex, message, sigHex string) (bool, error) {
	pub, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false, err
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, err
	}
	if len(pub) != ed25519.PublicKeySize {
		return false, nil
	}
	if len(sig) != ed25519.SignatureSize {
		return false, nil
	}
	return ed25519.Verify(ed25519.PublicKey(pub), []byte(message), sig), nil
}

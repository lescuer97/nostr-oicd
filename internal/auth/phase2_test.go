package auth

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	dbpkg "nostr-oidc-service/internal/db"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbConn, err := dbpkg.ConnectDB()
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	if err := dbpkg.RunMigrations(dbConn); err != nil {
		t.Fatalf("migrations: %v", err)
	}
	t.Cleanup(func() { dbConn.Close() })
	return dbConn
}

func TestChallengeAndVerify_Success(t *testing.T) {
	dbConn := setupTestDB(t)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	pubHex := hex.EncodeToString(pub)

	// Challenge
	req := httptest.NewRequest("GET", "/auth/challenge?public_key_hex="+pubHex, nil)
	w := httptest.NewRecorder()
	ChallengeHandler(dbConn).ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("challenge status: %d body: %s", w.Code, w.Body.String())
	}
	var ch struct {
		Challenge string `json:"challenge"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &ch); err != nil {
		t.Fatalf("decode: %v", err)
	}

	sig := ed25519.Sign(priv, []byte(ch.Challenge))
	sigHex := hex.EncodeToString(sig)

	// Verify
	payload := map[string]string{"public_key_hex": pubHex, "challenge": ch.Challenge, "signature": sigHex}
	body, _ := json.Marshal(payload)
	req2 := httptest.NewRequest("POST", "/auth/verify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	VerifyHandler(dbConn).ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Fatalf("verify status: %d body: %s", w2.Code, w2.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse resp: %v", err)
	}
	if s, ok := resp["session_id"].(string); !ok || s == "" {
		t.Fatalf("missing session_id in response: %v", resp)
	}
}

func TestChallengeAndVerify_WrongSignature(t *testing.T) {
	dbConn := setupTestDB(t)

	pubA, _, _ := ed25519.GenerateKey(rand.Reader)
	pubHexA := hex.EncodeToString(pubA)

	// Challenge for A
	req := httptest.NewRequest("GET", "/auth/challenge?public_key_hex="+pubHexA, nil)
	w := httptest.NewRecorder()
	ChallengeHandler(dbConn).ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("challenge failed: %d", w.Code)
	}
	var ch struct {
		Challenge string `json:"challenge"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &ch); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Sign with a different key (invalid signer)
	_, privB, _ := ed25519.GenerateKey(rand.Reader)
	badSig := ed25519.Sign(privB, []byte(ch.Challenge))
	badSigHex := hex.EncodeToString(badSig)

	payload := map[string]string{"public_key_hex": pubHexA, "challenge": ch.Challenge, "signature": badSigHex}
	body, _ := json.Marshal(payload)
	req2 := httptest.NewRequest("POST", "/auth/verify", bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	VerifyHandler(dbConn).ServeHTTP(w2, req2)
	if w2.Code != 401 {
		t.Fatalf("expected 401 for wrong signature, got %d: %s", w2.Code, w2.Body.String())
	}
}

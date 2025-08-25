package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/lescuer97/nostr-oicd/templates/fragments"
)

func ChallengeHandler(w http.ResponseWriter, r *http.Request) {
	// generate 32-byte challenge
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, "failed to generate challenge", http.StatusInternalServerError)
		return
	}
	challenge := hex.EncodeToString(b)
	SaveChallenge(challenge)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Render the templ fragment component
	comp := fragments.ChallengeFragment(challenge)
	if err := comp.Render(r.Context(), w); err != nil {
		// If rendering fails, fall back to a simple message
		fmt.Printf("failed to render challenge fragment: %v\n", err)
		http.Error(w, "failed to render fragment", http.StatusInternalServerError)
		return
	}
}

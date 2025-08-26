package auth

import (
	"errors"
	"strings"

	"github.com/nbd-wtf/go-nostr/nip19"
)

// decodeNpubToHex converts a nip19 npub string to hex public key (lowercase, no prefix)
func decodeNpubToHex(npub string) (string, error) {
	if npub == "" {
		return "", errors.New("empty")
	}
	npub = strings.TrimSpace(npub)
	prefix, value, err := nip19.Decode(npub)
	if err != nil {
		return "", err
	}
	if prefix != "npub" {
		return "", errors.New("not an npub string")
	}
	if pk, ok := value.(string); ok {
		return pk, nil
	}
	return "", errors.New("unexpected npub payload")
}

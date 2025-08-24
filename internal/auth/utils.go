package auth

import (
	"encoding/hex"
)

func HexDecode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

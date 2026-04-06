package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateRefreshToken creates a cryptographically random 32-byte token.
// It returns the raw hex-encoded token (stored in the cookie) and its
// SHA-256 hash (stored in the database).
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating refresh token: %w", err)
	}
	raw = hex.EncodeToString(b)
	hash = HashRefreshToken(raw)
	return raw, hash, nil
}

// HashRefreshToken returns the hex-encoded SHA-256 hash of a raw refresh token.
// SHA-256 is appropriate here because the token itself is a large random secret —
// bcrypt's purpose (key-stretching against dictionary attacks) is unnecessary.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// newToken returns 32 random bytes (256 bits), base64url-encoded. Used for
// session tokens and nonces.
func newToken() string {
	b := make([]byte, 32)
	// crypto/rand.Read never returns an error; it crashes the program instead
	// if the OS can't provide randomness.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashToken returns what the sessions table stores instead of the token.
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

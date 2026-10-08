package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

const (
	// Sliding: each use pushes expiry this far into the future again.
	sessionTTL = 90 * 24 * time.Hour
	// Expiry is extended at most this often, so a busy session costs one
	// extra write per day, not one per request.
	sessionTouchInterval = 24 * time.Hour
)

// newRandomToken returns 32 random bytes (256 bits), base64url-encoded.
func newRandomToken() string {
	b := make([]byte, 32)
	// crypto/rand.Read never returns an error; it crashes the program instead
	// if the OS can't provide randomness.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// hashToken returns what the sessions table stores instead of the token.
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

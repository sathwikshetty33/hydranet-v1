package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const tokenPrefix = "hyd_agent_"

func GenerateAgentToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return tokenPrefix + base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// HashToken bcrypt-hashes a token for storage.
func HashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash token: %w", err)
	}
	return string(hash), nil
}

// VerifyToken reports whether the raw token matches the stored bcrypt hash.
// Constant-time comparison is handled inside bcrypt.
func VerifyToken(token, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) == nil
}

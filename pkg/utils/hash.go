package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashPassword hashes a password using SHA-256
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	return HashPassword(password) == hash
}
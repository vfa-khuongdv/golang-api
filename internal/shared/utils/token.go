package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken hashes a token using SHA-256 for secure storage.
// Use this for reset tokens, verification tokens, etc. that need to be stored in DB.
// SHA-256 is appropriate here because:
//   - Tokens are already high-entropy (32+ random characters)
//   - We need fast hashing for token verification
//   - We don't need salt since tokens are unique per user
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CompareTokenHash compares a plain token with a stored hash.
// Returns true if they match.
func CompareTokenHash(token, storedHash string) bool {
	return HashToken(token) == storedHash
}

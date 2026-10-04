package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashToken(t *testing.T) {
	t.Run("generates consistent hash", func(t *testing.T) {
		rawValue := "abc123def456"
		hash1 := HashToken(rawValue)
		hash2 := HashToken(rawValue)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		hash1 := HashToken("token1")
		hash2 := HashToken("token2")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("hash is 64 characters (SHA-256 hex)", func(t *testing.T) {
		hash := HashToken("test")
		assert.Len(t, hash, 64)
	})

	t.Run("empty token produces valid hash", func(t *testing.T) {
		hash := HashToken("")
		assert.Len(t, hash, 64)
	})
}

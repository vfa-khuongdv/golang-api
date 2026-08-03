package utils_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

func TestHashPassword(t *testing.T) {
	t.Run("HashPassword", func(t *testing.T) {
		password := "mySecret123"
		hashed, err := utils.HashPassword(password)

		require.NoError(t, err, "HashPassword should not return an error")
		assert.NotEmpty(t, hashed, "Hashed password should not be empty")
		assert.NotEqual(t, password, hashed, "Hashed password should not equal plain password")
	})

	t.Run("CheckPasswordHash", func(t *testing.T) {
		password := "mySecret123"
		hashed, err := utils.HashPassword(password)
		require.NoError(t, err)

		// Valid match
		isMatch := utils.CheckPasswordHash(password, hashed)
		assert.True(t, isMatch, "Password should match the hash")

		// Invalid match
		isMatch = utils.CheckPasswordHash("wrongPassword", hashed)
		assert.False(t, isMatch, "Wrong password should not match hash")
	})

	t.Run("CheckPasswordHash_InvalidHash", func(t *testing.T) {
		password := "mySecret123"
		invalidHash := "invalid-hash"

		isMatch := utils.CheckPasswordHash(password, invalidHash)
		assert.False(t, isMatch, "Invalid hash should return false")
	})

	t.Run("HashPasswordWithInvalidCost", func(t *testing.T) {
		password := "mySecret123"
		hashed, err := utils.HashPasswordWithCost(password, 1000)

		assert.Error(t, err, "Should return error for invalid cost")
		assert.Empty(t, hashed, "Should return empty string on error")
	})

	t.Run("HashPasswordWithCost", func(t *testing.T) {
		password := "mySecret123"
		hashed, err := utils.HashPasswordWithCost(password, 4)

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.True(t, utils.CheckPasswordHash(password, hashed))
	})
}

func TestBcryptLengthLimit(t *testing.T) {
	t.Run("HashPasswordTooLong", func(t *testing.T) {
		password := strings.Repeat("a", 80)

		hashed, err := utils.HashPassword(password)
		assert.Error(t, err, "HashPassword should return error when password exceeds bcrypt length limit")
		assert.Empty(t, hashed)
	})

	t.Run("HashPassword72BytesSucceeds", func(t *testing.T) {
		// 72 bytes is the bcrypt maximum; it must hash successfully.
		password := strings.Repeat("a", 72)

		hashed, err := utils.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.True(t, utils.CheckPasswordHash(password, hashed))
	})

	t.Run("HashPassword73BytesFails", func(t *testing.T) {
		// 73 bytes exceeds the bcrypt limit.
		password := strings.Repeat("a", 73)

		hashed, err := utils.HashPassword(password)
		assert.Error(t, err)
		assert.Empty(t, hashed)
	})
}
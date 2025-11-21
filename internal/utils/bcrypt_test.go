package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func TestHashPassword(t *testing.T) {
	t.Run("HashPassword", func(t *testing.T) {
		password := "mySecret123"
		hashed := utils.HashPassword(password)

		assert.NotEmpty(t, hashed, "Hashed password should not be empty")
		assert.NotEqual(t, password, hashed, "Hashed password should not equal plain password")
	})

	t.Run("CheckPasswordHash", func(t *testing.T) {
		password := "mySecret123"
		hashed := utils.HashPassword(password)

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
		hashed := utils.HashPasswordWithCost(password, 1000) // cost 1000 will invalid cost triggers error
		assert.Equal(t, "", hashed, "Should return empty string on error")
	})
}

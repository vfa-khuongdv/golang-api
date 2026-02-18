package services_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptService(t *testing.T) {
	t.Run("HashAndCheckPassword", func(t *testing.T) {
		service := services.NewBcryptService()

		password := "securepassword123"
		hashedPassword, err := service.HashPassword(password)

		require.NoError(t, err, "HashPassword should not return an error")
		assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")
		assert.True(t, service.CheckPasswordHash(password, hashedPassword), "CheckPasswordHash should return true for valid password")
		assert.False(t, service.CheckPasswordHash("wrongpassword", hashedPassword), "CheckPasswordHash should return false for invalid password")
	})

	t.Run("HashPasswordWithCost", func(t *testing.T) {
		service := services.NewBcryptService()

		password := "anotherpassword456"
		cost := bcrypt.MinCost

		hashedPassword, err := service.HashPasswordWithCost(password, cost)
		require.NoError(t, err, "HashPasswordWithCost should not return an error")
		assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")
		assert.True(t, service.CheckPasswordHash(password, hashedPassword), "CheckPasswordHash should work with hashed password created using custom cost")
	})

	t.Run("HashPasswordWithInvalidCost", func(t *testing.T) {
		service := services.NewBcryptService()

		password := "invalidcost"
		invalidCost := 1000 // invalid bcrypt cost

		_, err := service.HashPasswordWithCost(password, invalidCost)
		assert.Error(t, err, "HashPasswordWithCost should return error for invalid cost")
	})

	t.Run("HashPasswordTooLong", func(t *testing.T) {
		service := services.NewBcryptService()
		tooLongPassword := strings.Repeat("a", 80)
		assert.Greater(t, len(tooLongPassword), 72)

		_, err := service.HashPassword(tooLongPassword)
		assert.Error(t, err, "HashPassword should return error when password exceeds bcrypt length limit")
	})
}

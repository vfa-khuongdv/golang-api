package services_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

func TestJWTService(t *testing.T) {
	t.Run("GenerateAndValidateToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Generate a token for user ID 123
		result, err := svc.GenerateToken(123)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Token)
		assert.True(t, result.ExpiresAt > time.Now().Unix())

		// Validate the generated token
		claims, err := svc.ValidateToken(result.Token)
		require.NoError(t, err)
		assert.Equal(t, uint(123), claims.ID)
		assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, time.Minute)
		assert.WithinDuration(t, time.Unix(result.ExpiresAt, 0), claims.ExpiresAt.Time, time.Minute)
	})

	t.Run("ValidateToken_InvalidToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Totally invalid token string
		_, err := svc.ValidateToken("this.is.not.a.token")
		assert.Error(t, err)

		// Token signed with different key
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &services.CustomClaims{
			ID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})
		// Sign with a different secret
		signedToken, err := token.SignedString([]byte("different_secret"))
		require.NoError(t, err)

		_, err = svc.ValidateToken(signedToken)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "signature is invalid") || strings.Contains(err.Error(), "token is invalid"))
	})
}

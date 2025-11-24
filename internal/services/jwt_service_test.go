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
	t.Run("GenerateAccessToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Generate an access token for user ID 456
		result, err := svc.GenerateAccessToken(456)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Token)

		// Validate the token
		claims, err := svc.ValidateToken(result.Token)
		require.NoError(t, err)
		assert.Equal(t, uint(456), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateTokenWithScope_AccessToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Generate an access token
		result, err := svc.GenerateAccessToken(123)
		require.NoError(t, err)

		// Validate with correct scope
		claims, err := svc.ValidateTokenWithScope(result.Token, services.TokenScopeAccess)
		require.NoError(t, err)
		assert.Equal(t, uint(123), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateToken_InvalidToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Totally invalid token string
		_, err := svc.ValidateToken("this.is.not.a.token")
		assert.Error(t, err)

		// Token signed with different key
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &services.CustomClaims{
			ID:    1,
			Scope: services.TokenScopeAccess,
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

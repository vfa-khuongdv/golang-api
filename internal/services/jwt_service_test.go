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
	t.Setenv("JWT_KEY", "unit-test-secret-key")

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

	t.Run("NewJWTService_PanicWhenSecretEmpty", func(t *testing.T) {
		t.Setenv("JWT_KEY", "   ")
		assert.Panics(t, func() {
			_ = services.NewJWTService()
		})
	})

	t.Run("ValidateTokenWithScope_Mismatch", func(t *testing.T) {
		svc := services.NewJWTService()

		result, err := svc.GenerateAccessToken(789)
		require.NoError(t, err)

		claims, err := svc.ValidateTokenWithScope(result.Token, "another-scope")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenWithScope_InvalidToken", func(t *testing.T) {
		svc := services.NewJWTService()
		claims, err := svc.ValidateTokenWithScope("invalid.token.value", services.TokenScopeAccess)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenIgnoreExpiration_ExpiredTokenSuccess", func(t *testing.T) {
		svc := services.NewJWTService()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &services.CustomClaims{
			ID:    21,
			Scope: services.TokenScopeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		})
		expiredToken, err := token.SignedString([]byte("unit-test-secret-key"))
		require.NoError(t, err)

		claims, err := svc.ValidateTokenIgnoreExpiration(expiredToken)
		require.NoError(t, err)
		assert.Equal(t, uint(21), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateTokenIgnoreExpiration_InvalidToken", func(t *testing.T) {
		svc := services.NewJWTService()
		claims, err := svc.ValidateTokenIgnoreExpiration("invalid.token.value")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

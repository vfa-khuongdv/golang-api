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
	t.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-testing-purposes-32-chars")

	t.Run("GenerateAccessToken", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		result, err := svc.GenerateAccessToken(456)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Token)

		claims, err := svc.ValidateToken(result.Token)
		require.NoError(t, err)
		assert.Equal(t, uint(456), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateTokenWithScope_AccessToken", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		result, err := svc.GenerateAccessToken(123)
		require.NoError(t, err)

		claims, err := svc.ValidateTokenWithScope(result.Token, services.TokenScopeAccess)
		require.NoError(t, err)
		assert.Equal(t, uint(123), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateToken_InvalidToken", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		_, err = svc.ValidateToken("this.is.not.a.token")
		assert.Error(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &services.CustomClaims{
			ID:    1,
			Scope: services.TokenScopeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})
		signedToken, err := token.SignedString([]byte("different_secret"))
		require.NoError(t, err)

		_, err = svc.ValidateToken(signedToken)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "signature is invalid") || strings.Contains(err.Error(), "token is invalid"))
	})

	t.Run("NewJWTService_ErrorWhenSecretEmpty", func(t *testing.T) {
		t.Setenv("JWT_KEY", "   ")
		_, err := services.NewJWTService()
		assert.Error(t, err)
		assert.Equal(t, services.ErrJWTKeyMissing, err)
	})

	t.Run("NewJWTService_ErrorWhenSecretTooShort", func(t *testing.T) {
		t.Setenv("JWT_KEY", "short")
		_, err := services.NewJWTService()
		assert.Error(t, err)
		assert.Equal(t, services.ErrJWTKeyTooShort, err)
	})

	t.Run("ValidateTokenWithScope_Mismatch", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		result, err := svc.GenerateAccessToken(789)
		require.NoError(t, err)

		claims, err := svc.ValidateTokenWithScope(result.Token, "another-scope")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenWithScope_InvalidToken", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		claims, err := svc.ValidateTokenWithScope("invalid.token.value", services.TokenScopeAccess)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenIgnoreExpiration_ExpiredTokenSuccess", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &services.CustomClaims{
			ID:    21,
			Scope: services.TokenScopeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		})
		expiredToken, err := token.SignedString([]byte("this-is-a-very-long-secret-key-for-testing-purposes-32-chars"))
		require.NoError(t, err)

		claims, err := svc.ValidateTokenIgnoreExpiration(expiredToken)
		require.NoError(t, err)
		assert.Equal(t, uint(21), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateTokenIgnoreExpiration_InvalidToken", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		claims, err := svc.ValidateTokenIgnoreExpiration("invalid.token.value")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

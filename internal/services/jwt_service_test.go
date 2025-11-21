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
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
		assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, time.Minute)
		assert.WithinDuration(t, time.Unix(result.ExpiresAt, 0), claims.ExpiresAt.Time, time.Minute)
	})

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

	t.Run("GenerateMfaToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Generate an MFA token for user ID 789
		result, err := svc.GenerateMfaToken(789)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Token)

		// Validate the token
		claims, err := svc.ValidateToken(result.Token)
		require.NoError(t, err)
		assert.Equal(t, uint(789), claims.ID)
		assert.Equal(t, services.TokenScopeMfaVerification, claims.Scope)
		// MFA token expires in 10 minutes
		assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
		assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(15*time.Minute)))
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

		// Try to validate with wrong scope (should fail)
		_, err = svc.ValidateTokenWithScope(result.Token, services.TokenScopeMfaVerification)
		assert.Error(t, err)
	})

	t.Run("ValidateTokenWithScope_MfaToken", func(t *testing.T) {
		svc := services.NewJWTService()

		// Generate an MFA token
		result, err := svc.GenerateMfaToken(456)
		require.NoError(t, err)

		// Validate with correct scope
		claims, err := svc.ValidateTokenWithScope(result.Token, services.TokenScopeMfaVerification)
		require.NoError(t, err)
		assert.Equal(t, uint(456), claims.ID)
		assert.Equal(t, services.TokenScopeMfaVerification, claims.Scope)

		// Try to validate with wrong scope (should fail)
		_, err = svc.ValidateTokenWithScope(result.Token, services.TokenScopeAccess)
		assert.Error(t, err)
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

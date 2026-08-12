package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

func TestJWTService(t *testing.T) {
	const testJWTSecret = "this-is-a-very-long-secret-key-for-testing-purposes-32-chars"
	t.Setenv("JWT_KEY", testJWTSecret)

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

	t.Run("NewJWTService_SecretLengthBoundary", func(t *testing.T) {
		t.Run("31 chars rejected", func(t *testing.T) {
			t.Setenv("JWT_KEY", strings.Repeat("a", 31))
			_, err := services.NewJWTService()
			assert.Error(t, err)
			assert.Equal(t, services.ErrJWTKeyTooShort, err)
		})

		t.Run("32 chars accepted", func(t *testing.T) {
			t.Setenv("JWT_KEY", strings.Repeat("b", 32))
			svc, err := services.NewJWTService()
			assert.NoError(t, err)
			assert.NotNil(t, svc)
		})
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
		expiredToken, err := token.SignedString([]byte(testJWTSecret))
		require.NoError(t, err)

		claims, err := svc.ValidateTokenIgnoreExpiration(expiredToken)
		require.NoError(t, err)
		assert.Equal(t, uint(21), claims.ID)
		assert.Equal(t, services.TokenScopeAccess, claims.Scope)
	})

	t.Run("ValidateToken_ExpiredTokenRejected", func(t *testing.T) {
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
		expiredToken, err := token.SignedString([]byte(testJWTSecret))
		require.NoError(t, err)

		// The strict path must reject an expired token, even with a valid signature.
		claims, err := svc.ValidateToken(expiredToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenIgnoreExpiration_WrongSignature", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &services.CustomClaims{
			ID:    21,
			Scope: services.TokenScopeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})
		signedToken, err := token.SignedString([]byte("a-completely-different-secret-key-for-wrong-signature-test"))
		require.NoError(t, err)

		claims, err := svc.ValidateTokenIgnoreExpiration(signedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenIgnoreExpiration_InvalidToken", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		claims, err := svc.ValidateTokenIgnoreExpiration("invalid.token.value")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateToken_RejectNonHMACAlg", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		// An RS256 token is the classic alg-confusion attack vector:
		// the server must reject it even though it looks structurally valid.
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, &services.CustomClaims{
			ID:    1,
			Scope: services.TokenScopeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		claims, err := svc.ValidateToken(signedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})

	t.Run("ValidateTokenIgnoreExpiration_RejectNonHMACAlg", func(t *testing.T) {
		svc, err := services.NewJWTService()
		require.NoError(t, err)

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, &services.CustomClaims{
			ID:    1,
			Scope: services.TokenScopeAccess,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		})
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		// The lenient path still validates the signature algorithm,
		// so a non-HMAC token must not slip through either.
		claims, err := svc.ValidateTokenIgnoreExpiration(signedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})
}

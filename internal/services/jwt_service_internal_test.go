package services

import (
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTService_InternalBranches(t *testing.T) {
	originalSign := signJWTToken
	originalParse := parseJWTWithClaims
	t.Cleanup(func() {
		signJWTToken = originalSign
		parseJWTWithClaims = originalParse
	})

	t.Setenv("JWT_KEY", "internal-secret")
	svc := NewJWTService()

	t.Run("GenerateAccessTokenSigningError", func(t *testing.T) {
		signJWTToken = func(_ *jwt.Token, _ []byte) (string, error) {
			return "", errors.New("sign failed")
		}

		result, err := svc.GenerateAccessToken(1)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("ValidateTokenInvalidClaimsType", func(t *testing.T) {
		parseJWTWithClaims = func(_ string, _ jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
			return &jwt.Token{Claims: jwt.MapClaims{}, Valid: true}, nil
		}

		claims, err := svc.ValidateToken("any")
		assert.Nil(t, claims)
		assert.NoError(t, err)
	})

	t.Run("ValidateTokenIgnoreExpirationInvalidClaimsType", func(t *testing.T) {
		parseJWTWithClaims = func(_ string, _ jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
			return &jwt.Token{Claims: jwt.MapClaims{}, Valid: true}, nil
		}

		claims, err := svc.ValidateTokenIgnoreExpiration("any")
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, jwt.ErrInvalidType)
	})
}

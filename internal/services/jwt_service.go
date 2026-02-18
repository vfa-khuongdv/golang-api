package services

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

const (
	// TokenScopeAccess is the scope for regular access tokens
	TokenScopeAccess = "access"
)

// CustomClaims represents JWT claims with a custom user ID field and scope
type CustomClaims struct {
	ID    uint   `json:"id"`
	Scope string `json:"scope"` // Token scope: "access" or "mfa_verification"
	jwt.RegisteredClaims
}

// JWTService defines JWT-related operations
type JWTService interface {
	GenerateAccessToken(id uint) (*dto.JwtResult, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
	ValidateTokenWithScope(tokenString string, requiredScope string) (*CustomClaims, error)
	ValidateTokenIgnoreExpiration(tokenString string) (*CustomClaims, error)
}

// jwtServiceImpl implements JWTService
type jwtServiceImpl struct {
	secret []byte
}

var (
	signJWTToken = func(token *jwt.Token, secret []byte) (string, error) {
		return token.SignedString(secret)
	}
	parseJWTWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
		return jwt.ParseWithClaims(tokenString, claims, keyFunc, options...)
	}
)

// NewJWTService returns a new instance of jwtServiceImpl
func NewJWTService() JWTService {
	secret := strings.TrimSpace(utils.GetEnv("JWT_KEY", ""))
	if secret == "" {
		panic("JWT_KEY environment variable is required")
	}
	return &jwtServiceImpl{
		secret: []byte(secret),
	}
}

// GenerateAccessToken creates a new access JWT token for the given user ID
// Access tokens have 1-hour expiration and can access all authenticated endpoints
func (s *jwtServiceImpl) GenerateAccessToken(id uint) (*dto.JwtResult, error) {
	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour))
	claims := CustomClaims{
		ID:    id,
		Scope: TokenScopeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := signJWTToken(token, s.secret)
	if err != nil {
		return nil, err
	}

	return &dto.JwtResult{
		Token:     signedToken,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// ValidateToken validates a JWT token string and returns the claims if valid
func (s *jwtServiceImpl) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := parseJWTWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// ValidateTokenWithScope validates a JWT token string with a specific required scope
// Returns error if token is invalid or scope does not match
func (s *jwtServiceImpl) ValidateTokenWithScope(tokenString string, requiredScope string) (*CustomClaims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Scope != requiredScope {
		return nil, jwt.ErrInvalidType
	}

	return claims, nil
}

// ValidateTokenIgnoreExpiration validates a JWT token string but ignores expiration time
// This is useful when you want to extract user information from expired tokens
// Returns error if token signature is invalid, but ignores exp claim
func (s *jwtServiceImpl) ValidateTokenIgnoreExpiration(tokenString string) (*CustomClaims, error) {
	token, err := parseJWTWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		// We do a basic signature validation here
		// The token signature is valid if ParseWithClaims succeeded
		return claims, nil
	}

	return nil, jwt.ErrInvalidType
}

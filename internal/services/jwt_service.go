package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

const (
	// TokenScopeAccess is the scope for regular access tokens
	TokenScopeAccess = "access"
	// TokenScopeMfaVerification is the scope for MFA verification tokens (temporary)
	TokenScopeMfaVerification = "mfa_verification"
)

// CustomClaims represents JWT claims with a custom user ID field and scope
type CustomClaims struct {
	ID    uint   `json:"id"`
	Scope string `json:"scope"` // Token scope: "access" or "mfa_verification"
	jwt.RegisteredClaims
}

// JwtResult represents the result of a token generation
type JwtResult struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// IJWTService defines JWT-related operations
type IJWTService interface {
	GenerateAccessToken(id uint) (*JwtResult, error)
	GenerateMfaToken(id uint) (*JwtResult, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
	ValidateTokenWithScope(tokenString string, requiredScope string) (*CustomClaims, error)
	ValidateTokenIgnoreExpiration(tokenString string) (*CustomClaims, error)
}

// jwtService implements JWTService
type jwtService struct {
	secret []byte
}

// NewJWTService returns a new instance of jwtService
func NewJWTService() IJWTService {
	secret := []byte(utils.GetEnv("JWT_KEY", "replace_your_key"))
	return &jwtService{
		secret: secret,
	}
}

// GenerateAccessToken creates a new access JWT token for the given user ID
// Access tokens have 1-hour expiration and can access all authenticated endpoints
func (s *jwtService) GenerateAccessToken(id uint) (*JwtResult, error) {
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
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &JwtResult{
		Token:     signedToken,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// GenerateMfaToken creates a temporary JWT token for MFA verification
// MFA tokens have 10-minute expiration and can only be used for /mfa/verify-code endpoint
func (s *jwtService) GenerateMfaToken(id uint) (*JwtResult, error) {
	expiresAt := jwt.NewNumericDate(time.Now().Add(10 * time.Minute))
	claims := CustomClaims{
		ID:    id,
		Scope: TokenScopeMfaVerification,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &JwtResult{
		Token:     signedToken,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// ValidateToken validates a JWT token string and returns the claims if valid
func (s *jwtService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
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
func (s *jwtService) ValidateTokenWithScope(tokenString string, requiredScope string) (*CustomClaims, error) {
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
func (s *jwtService) ValidateTokenIgnoreExpiration(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
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

package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

// CustomClaims represents JWT claims with a custom user ID field
type CustomClaims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

// JwtResult represents the result of a token generation
type JwtResult struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// IJWTService defines JWT-related operations
type IJWTService interface {
	GenerateToken(id uint) (*JwtResult, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
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

// GenerateToken creates a new JWT token for the given user ID
func (s *jwtService) GenerateToken(id uint) (*JwtResult, error) {
	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour))
	claims := CustomClaims{
		ID: id,
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

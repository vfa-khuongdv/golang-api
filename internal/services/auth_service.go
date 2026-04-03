package services

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(email, password string, ctx *gin.Context) (*dto.LoginResponse, error)
	RefreshToken(refreshToken, accessToken string, ctx *gin.Context) (*dto.LoginResponse, error)
}

type authServiceImpl struct {
	repo                repositories.UserRepository
	refreshTokenService RefreshTokenService
	bcryptService       BcryptService
	jwtService          JWTService
}

// NewAuthService creates and returns a new instance of AuthService
// Parameters:
//   - repo: UserRepository for user data access
//   - refreshTokenService: RefreshTokenService for managing refresh tokens
//   - bcryptService: BcryptService for password hashing and verification
//   - jwtService: JWTService for JWT token generation and validation
//
// Returns:
//   - AuthService: New AuthService instance initialized with the provided dependencies
func NewAuthService(repo repositories.UserRepository, refreshTokenService RefreshTokenService, bcryptService BcryptService, jwtService JWTService) AuthService {
	return &authServiceImpl{
		repo:                repo,
		refreshTokenService: refreshTokenService,
		bcryptService:       bcryptService,
		jwtService:          jwtService,
	}
}

// Login authenticates a user with their email and password
// Parameters:
//   - email: The email of the user trying to log in
//   - password: The password provided by the user
//   - ctx: Gin context containing request information
//
// Returns:
//   - *dto.LoginResponse: Contains access token and refresh token if successful
//   - error: Returns error if login fails (user not found, invalid password, token generation fails)
func (service *authServiceImpl) Login(email, password string, ctx *gin.Context) (*dto.LoginResponse, error) {
	logger.Infof("Login attempt for email: %s", email)

	user, err := service.repo.FindByField("email", email)
	if err != nil {
		logger.Warnf("Login failed - user not found: %s", email)
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	if isValid := service.bcryptService.CheckPasswordHash(password, user.Password); !isValid {
		logger.Warnf("Login failed - invalid password for email: %s", email)
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	accessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		logger.Errorf("Failed to generate access token for user ID %d: %v", user.ID, err)
		return nil, apperror.NewInternalServerError(err.Error())
	}

	ipAddress := ctx.ClientIP()
	refreshToken, errToken := service.refreshTokenService.Create(user, ipAddress)

	if errToken != nil {
		logger.Errorf("Failed to create refresh token for user ID %d: %v", user.ID, errToken)
		return nil, errToken
	}

	logger.Infof("Login successful for user ID %d", user.ID)

	return &dto.LoginResponse{
		AccessToken: dto.JwtResult{
			Token:     accessToken.Token,
			ExpiresAt: accessToken.ExpiresAt,
		},
		RefreshToken: dto.JwtResult{
			Token:     refreshToken.Token,
			ExpiresAt: refreshToken.ExpiresAt,
		},
	}, nil
}

// RefreshToken generates new access and refresh tokens using refresh_token and validates with access_token
// Parameters:
//   - refreshToken: The existing refresh token string (used to identify user and generate new tokens)
//   - accessToken: The existing access token string (used to verify token ownership, can be expired)
//   - ctx: Gin context containing request information
//
// Returns:
//   - *dto.LoginResponse: Contains new access token and refresh token if successful
//   - error: Returns error if token refresh fails (invalid tokens, user not found, token generation fails)
func (service *authServiceImpl) RefreshToken(refreshToken, accessToken string, ctx *gin.Context) (*dto.LoginResponse, error) {
	logger.Info("Token refresh attempt")

	ipAddress := ctx.ClientIP()

	refreshResult, err := service.refreshTokenService.Update(refreshToken, ipAddress)
	if err != nil {
		logger.Warnf("Token refresh failed - invalid refresh token")
		return nil, apperror.NewUnauthorizedError("Invalid refresh token")
	}

	claims, err := service.jwtService.ValidateTokenIgnoreExpiration(accessToken)
	if err != nil {
		logger.Warnf("Token refresh failed - invalid access token")
		return nil, apperror.NewUnauthorizedError("Invalid access token")
	}

	if claims.Scope != TokenScopeAccess {
		logger.Warnf("Token refresh failed - invalid scope")
		return nil, apperror.NewUnauthorizedError("Invalid access token scope")
	}

	if claims.ID != refreshResult.UserId {
		logger.Warnf("Token refresh failed - token mismatch")
		return nil, apperror.NewUnauthorizedError("Token mismatch: refresh and access tokens belong to different users")
	}

	user, err := service.repo.GetByID(refreshResult.UserId)
	if err != nil {
		logger.Warnf("Token refresh failed - user not found: %d", refreshResult.UserId)
		return nil, apperror.NewNotFoundError("User not found")
	}

	newAccessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		logger.Errorf("Failed to generate new access token for user ID %d: %v", user.ID, err)
		return nil, apperror.NewInternalServerError(err.Error())
	}

	logger.Infof("Token refresh successful for user ID %d", user.ID)

	return &dto.LoginResponse{
		AccessToken: dto.JwtResult{
			Token:     newAccessToken.Token,
			ExpiresAt: newAccessToken.ExpiresAt,
		},
		RefreshToken: dto.JwtResult{
			Token:     refreshResult.Token.Token,
			ExpiresAt: refreshResult.Token.ExpiresAt,
		},
	}, nil
}

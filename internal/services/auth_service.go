package services

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
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
	user, err := service.repo.FindByField("email", email)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}

	// Validate password
	if isValid := service.bcryptService.CheckPasswordHash(password, user.Password); !isValid {
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	// Generate access token
	accessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}

	// Create new refresh token
	ipAddress := ctx.ClientIP()
	refreshToken, errToken := service.refreshTokenService.Create(user, ipAddress)

	if errToken != nil {
		return nil, errToken
	}

	res := &dto.LoginResponse{
		AccessToken: dto.JwtResult{
			Token:     accessToken.Token,
			ExpiresAt: accessToken.ExpiresAt,
		},
		RefreshToken: dto.JwtResult{
			Token:     refreshToken.Token,
			ExpiresAt: refreshToken.ExpiresAt,
		},
	}

	return res, nil
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
	ipAddress := ctx.ClientIP()

	// Step 1: Validate refresh token (used to identify user)
	refreshResult, err := service.refreshTokenService.Update(refreshToken, ipAddress)
	if err != nil {
		return nil, apperror.NewUnauthorizedError("Invalid refresh token")
	}

	// Step 2: Validate access token (verify token ownership, works even if expired)
	claims, err := service.jwtService.ValidateTokenIgnoreExpiration(accessToken)
	if err != nil {
		return nil, apperror.NewUnauthorizedError("Invalid access token")
	}

	// Step 3: Verify access token has correct scope
	if claims.Scope != TokenScopeAccess {
		return nil, apperror.NewUnauthorizedError("Invalid access token scope")
	}

	// Step 4: Verify that both tokens belong to the same user
	if claims.ID != refreshResult.UserId {
		return nil, apperror.NewUnauthorizedError("Token mismatch: refresh and access tokens belong to different users")
	}

	// Step 5: Get user details
	user, err := service.repo.GetByID(refreshResult.UserId)
	if err != nil {
		return nil, apperror.NewNotFoundError("User not found")
	}

	// Step 6: Generate new access token
	newAccessToken, err := service.jwtService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}

	// Step 7: Build response (refresh token already updated in Step 1)
	response := &dto.LoginResponse{
		AccessToken: dto.JwtResult{
			Token:     newAccessToken.Token,
			ExpiresAt: newAccessToken.ExpiresAt,
		},
		RefreshToken: dto.JwtResult{
			Token:     refreshResult.Token.Token,
			ExpiresAt: refreshResult.Token.ExpiresAt,
		},
	}

	return response, nil
}

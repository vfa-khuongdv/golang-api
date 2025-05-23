package services

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type IAuthService interface {
	Login(email, password string, ctx *gin.Context) (*LoginResponse, error)
	RefreshToken(token string, ctx *gin.Context) (*LoginResponse, error)
}

type AuthService struct {
	repo         repositories.IUserRepository
	tokenService IRefreshTokenService
}

type LoginResponse struct {
	AccessToken  configs.JwtResult `json:"accessToken"`
	RefreshToken configs.JwtResult `json:"refreshToken"`
}

// NewAuthService creates and returns a new instance of AuthService
// Parameters:
//   - repo: User repository for database operations
//   - tokenService: Service for handling refresh token operations
//
// Returns:
//   - *AuthService: New AuthService instance initialized with the provided dependencies
func NewAuthService(repo repositories.IUserRepository, tokenService IRefreshTokenService) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: tokenService,
	}
}

// Login authenticates a user with their username and password
// Parameters:
//   - username: The username of the user trying to log in
//   - password: The password provided by the user
//   - ctx: Gin context containing request information
//
// Returns:
//   - *LoginResponse: Contains access token and refresh token if login successful
//   - error: Returns error if login fails (user not found, invalid password, token generation fails)
func (service *AuthService) Login(email, password string, ctx *gin.Context) (*LoginResponse, error) {
	user, err := service.repo.FindByField("email", email)
	if err != nil {
		return nil, errors.New(errors.ErrResourceNotFound, err.Error())
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New(errors.ErrAuthInvalidPassword, err.Error())
	}

	// Generate refresh token
	token, err := configs.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New(errors.ErrServerInternal, err.Error())
	}

	// Create new refresh token
	ipAddress := ctx.ClientIP()
	tokenResult, err := service.tokenService.Create(*user, ipAddress)
	if err != nil {
		return nil, err // error is already wrapped by the service, so we can return it directly
	}

	res := &LoginResponse{
		AccessToken: configs.JwtResult{
			Token:     token.Token,
			ExpiresAt: token.ExpiresAt,
		},
		RefreshToken: configs.JwtResult{
			Token:     tokenResult.Token,
			ExpiresAt: tokenResult.ExpiresAt,
		},
	}

	return res, nil
}

// RefreshToken generates new access and refresh tokens using an existing refresh token
// Parameters:
//   - token: The existing refresh token string
//   - ctx: Gin context containing request information
//
// Returns:
//   - *LoginResponse: Contains new access token and refresh token if successful
//   - error: Returns error if token refresh fails (invalid token, user not found, token generation fails)
func (service *AuthService) RefreshToken(token string, ctx *gin.Context) (*LoginResponse, error) {
	// Get the client's IP address from the request context
	ipAddress := ctx.ClientIP()
	// Create new refresh token using the token service
	res, err := service.tokenService.CreateRefreshToken(token, ipAddress)
	if err != nil {
		return nil, err // error is already wrapped by the service, so we can return it directly
	}

	// Get user details from the database using the user ID from refresh token
	user, err := service.repo.GetByID(res.UserId)
	if err != nil {
		return nil, errors.New(errors.ErrDatabaseQuery, err.Error())
	}
	// Generate new access token for the user
	resultToken, err := configs.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New(errors.ErrServerInternal, err.Error())
	}

	// Return new access and refresh tokens
	return &LoginResponse{
		AccessToken: configs.JwtResult{
			Token:     resultToken.Token,
			ExpiresAt: resultToken.ExpiresAt,
		},
		RefreshToken: *res.Token,
	}, nil

}

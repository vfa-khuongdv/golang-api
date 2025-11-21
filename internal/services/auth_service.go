package services

import (
	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type IAuthService interface {
	Login(email, password string, ctx *gin.Context) (interface{}, error)
	RefreshToken(token string, ctx *gin.Context) (*LoginResponse, error)
}

type AuthService struct {
	repo                repositories.IUserRepository
	refreshTokenService IRefreshTokenService
	bcryptService       IBcryptService
	jwtService          IJWTService
	mfaRepository       repositories.IMfaRepository
}

type LoginResponse struct {
	AccessToken  JwtResult `json:"access_token"`
	RefreshToken JwtResult `json:"refresh_token"`
}

// MfaRequiredResponse is returned when user has MFA enabled but hasn't verified it yet
type MfaRequiredResponse struct {
	MfaRequired    bool   `json:"mfa_required"`
	TemporaryToken string `json:"temporary_token"`
	Message        string `json:"message"`
}

// NewAuthService creates and returns a new instance of AuthService
// Parameters:
//   - repo: IUserRepository for user data access
//   - refreshTokenService: IRefreshTokenService for managing refresh tokens
//   - bcryptService: IBcryptService for password hashing and verification
//   - jwtService: IJWTService for JWT token generation and validation
//   - mfaRepository: IMfaRepository for accessing MFA settings
//
// Returns:
//   - *AuthService: New AuthService instance initialized with the provided dependencies
func NewAuthService(repo repositories.IUserRepository, refreshTokenService IRefreshTokenService, bcryptService IBcryptService, jwtService IJWTService, mfaRepository repositories.IMfaRepository) *AuthService {
	return &AuthService{
		repo:                repo,
		refreshTokenService: refreshTokenService,
		bcryptService:       bcryptService,
		jwtService:          jwtService,
		mfaRepository:       mfaRepository,
	}
}

// Login authenticates a user with their email and password
// Parameters:
//   - email: The email of the user trying to log in
//   - password: The password provided by the user
//   - ctx: Gin context containing request information
//
// Returns:
//   - interface{}: Returns either LoginResponse (full tokens) if MFA not enabled,
//     or MfaRequiredResponse (temporary token) if MFA is enabled
//   - error: Returns error if login fails (user not found, invalid password, token generation fails)
func (service *AuthService) Login(email, password string, ctx *gin.Context) (interface{}, error) {
	user, err := service.repo.FindByField("email", email)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}

	// Validate password
	if isValid := service.bcryptService.CheckPasswordHash(password, user.Password); !isValid {
		return nil, apperror.NewInvalidPasswordError("Invalid credentials")
	}

	// Check if user has MFA enabled
	mfaSettings, err := service.mfaRepository.GetMfaSettingsByUserID(user.ID)
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}

	// If MFA is enabled, return temporary token for MFA verification
	if mfaSettings != nil && mfaSettings.MfaEnabled {
		// Generate temporary token (short-lived, only for MFA verification)
		tempToken, err := service.jwtService.GenerateToken(user.ID)
		if err != nil {
			return nil, apperror.NewInternalError(err.Error())
		}

		return &MfaRequiredResponse{
			MfaRequired:    true,
			TemporaryToken: tempToken.Token,
			Message:        "MFA code required",
		}, nil
	}

	// MFA is not enabled, proceed with normal login
	// Generate access token
	accessToken, err := service.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}

	// Create new refresh token
	ipAddress := ctx.ClientIP()
	refreshToken, errToken := service.refreshTokenService.Create(user, ipAddress)

	if errToken != nil {
		return nil, errToken
	}

	res := &LoginResponse{
		AccessToken: JwtResult{
			Token:     accessToken.Token,
			ExpiresAt: accessToken.ExpiresAt,
		},
		RefreshToken: JwtResult{
			Token:     refreshToken.Token,
			ExpiresAt: refreshToken.ExpiresAt,
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
	ipAddress := ctx.ClientIP()

	// Update the refresh token
	refreshResult, err := service.refreshTokenService.Update(token, ipAddress)
	if err != nil {
		return nil, apperror.NewDBUpdateError(err.Error())
	}

	// Get user details
	user, err := service.repo.GetByID(refreshResult.UserId)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}

	// Generate new access token
	newToken, err := service.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, apperror.NewInternalError(err.Error())
	}

	// Build response
	response := &LoginResponse{
		AccessToken: JwtResult{
			Token:     newToken.Token,
			ExpiresAt: newToken.ExpiresAt,
		},
		RefreshToken: *refreshResult.Token,
	}

	return response, nil
}

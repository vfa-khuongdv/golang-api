package services

import (
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

type IRefreshTokenService interface {
	Create(user *models.User, ipAddress string) (*dto.JwtResult, error)
	Update(token string, ipAddress string) (*RefreshTokenResult, error)
}

type RefreshTokenService struct {
	repo repositories.IRefreshTokenRepository
}

// NewRefreshTokenService creates a new instance of RefreshTokenService
// Parameters:
//   - repo: Pointer to RefreshTokenRepository that handles refresh token database operations
//
// Returns:
//   - *RefreshTokenService: New instance of RefreshTokenService initialized with the provided repository
func NewRefreshTokenService(repo repositories.IRefreshTokenRepository) *RefreshTokenService {
	return &RefreshTokenService{
		repo: repo,
	}
}

// Create creates a new refresh token for a user
// Parameters:
//   - user: User model containing user information
//   - ipAddress: IP address of the user making the request
//
// Returns:
//   - *dto.JwtResult: Contains the generated token and expiration time
//   - error: Error if token creation fails
func (service *RefreshTokenService) Create(user *models.User, ipAddress string) (*dto.JwtResult, error) {
	tokenString := utils.GenerateRandomString(60)
	expiredAt := time.Now().Add(time.Hour * 24 * 30).Unix()
	token := models.RefreshToken{
		RefreshToken: tokenString,
		IpAddress:    ipAddress, // ipaddress of user
		UsedCount:    0,         // init is zero
		ExpiredAt:    expiredAt, // 30 days
		UserID:       user.ID,   // userId
	}

	err := service.repo.Create(&token)
	if err != nil {
		return nil, apperror.NewDBInsertError(err.Error())
	}

	return &dto.JwtResult{
		Token:     tokenString,
		ExpiresAt: expiredAt,
	}, nil
}

type RefreshTokenResult struct {
	Token  *dto.JwtResult
	UserId uint
}

// Update replaces an existing refresh token with a new one
// Parameters:
//   - tokenString: The existing refresh token string to be replaced
//   - ipAddress: IP address of the user making the request
//
// Returns:
//   - *RefreshTokenResult: Contains the new token information and associated user ID
//   - *appError.AppError: Error if token creation/update fails
//
// The function:
//  1. Finds the existing token record
//  2. Generates a new random token string
//  3. Updates the token record with new token, expiry and IP
//  4. Returns the new token details and associated user ID
func (service *RefreshTokenService) Update(tokenString string, ipAddress string) (*RefreshTokenResult, error) {
	result, err := service.repo.FindByToken(tokenString)
	if err != nil {
		return nil, apperror.NewNotFoundError(err.Error())
	}
	// Update new token
	newToken := utils.GenerateRandomString(60)
	expiredAt := time.Now().Add(time.Hour * 24 * 30).Unix()

	result.RefreshToken = newToken
	result.ExpiredAt = expiredAt
	result.IpAddress = ipAddress
	result.UsedCount += 1

	if err := service.repo.Update(result); err != nil {
		return nil, apperror.NewDBUpdateError(err.Error())
	}

	return &RefreshTokenResult{
		Token: &dto.JwtResult{
			Token:     newToken,
			ExpiresAt: expiredAt,
		},
		UserId: result.UserID,
	}, nil
}

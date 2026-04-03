package services

import (
	"context"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type RefreshTokenService interface {
	Create(ctx context.Context, user *models.User, ipAddress string) (*dto.JwtResult, error)
	Update(ctx context.Context, token string, ipAddress string) (*RefreshTokenResult, error)
}

type refreshTokenServiceImpl struct {
	repo repositories.RefreshTokenRepository
}

func NewRefreshTokenService(repo repositories.RefreshTokenRepository) RefreshTokenService {
	return &refreshTokenServiceImpl{
		repo: repo,
	}
}

func (service *refreshTokenServiceImpl) Create(ctx context.Context, user *models.User, ipAddress string) (*dto.JwtResult, error) {
	tokenString := utils.GenerateRandomString(60)
	expiredAt := time.Now().Add(time.Hour * 24 * 30).Unix()
	token := models.RefreshToken{
		RefreshToken: tokenString,
		IpAddress:    ipAddress,
		UsedCount:    0,
		ExpiredAt:    expiredAt,
		UserID:       user.ID,
	}

	err := service.repo.Create(ctx, &token)
	if err != nil {
		logger.Errorf("Failed to create refresh token: %v", err)
		return nil, apperror.NewDBInsertError("Failed to create refresh token")
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

func (service *refreshTokenServiceImpl) Update(ctx context.Context, tokenString string, ipAddress string) (*RefreshTokenResult, error) {
	result, err := service.repo.FindByToken(ctx, tokenString)
	if err != nil {
		return nil, apperror.NewNotFoundError("Refresh token not found or expired")
	}

	newToken := utils.GenerateRandomString(60)
	expiredAt := time.Now().Add(time.Hour * 24 * 30).Unix()

	result.RefreshToken = newToken
	result.ExpiredAt = expiredAt
	result.IpAddress = ipAddress
	result.UsedCount += 1

	if err := service.repo.Update(ctx, result); err != nil {
		logger.Errorf("Failed to update refresh token: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update refresh token")
	}

	return &RefreshTokenResult{
		Token: &dto.JwtResult{
			Token:     newToken,
			ExpiresAt: expiredAt,
		},
		UserId: result.UserID,
	}, nil
}

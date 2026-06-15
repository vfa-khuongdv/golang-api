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
		logger.WithContext(ctx).Errorf("Failed to create refresh token: %v", err)
		return nil, apperror.NewDBInsertError("Failed to create refresh token")
	}

	logger.WithContext(ctx).Infof("Created refresh token for user ID %d", user.ID)

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
	tx, err := service.repo.BeginTx(ctx)
	if err != nil {
		logger.WithContext(ctx).Errorf("Failed to begin transaction: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update refresh token")
	}

	result, err := service.repo.FindByTokenWithTx(ctx, tx, tokenString)
	if err != nil {
		if rerr := tx.Rollback().Error; rerr != nil {
			logger.WithContext(ctx).Errorf("Rollback failed: %v", rerr)
		}
		return nil, err
	}

	newToken := utils.GenerateRandomString(60)
	expiredAt := time.Now().Add(time.Hour * 24 * 30).Unix()

	result.RefreshToken = newToken
	result.ExpiredAt = expiredAt
	result.IpAddress = ipAddress
	result.UsedCount += 1

	if err := service.repo.UpdateWithTx(ctx, tx, result); err != nil {
		if rerr := tx.Rollback().Error; rerr != nil {
			logger.WithContext(ctx).Errorf("Rollback failed: %v", rerr)
		}
		logger.WithContext(ctx).Errorf("Failed to update refresh token: %v", err)
		return nil, apperror.NewDBUpdateError("Failed to update refresh token")
	}

	if err := tx.Commit().Error; err != nil {
		logger.WithContext(ctx).Errorf("Failed to commit transaction: %v", err)
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

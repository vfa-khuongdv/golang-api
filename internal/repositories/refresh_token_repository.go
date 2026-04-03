package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	Update(ctx context.Context, token *models.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*models.RefreshToken, error)
	UpdateWithTx(ctx context.Context, token *models.RefreshToken, tx *gorm.DB) error
}

type refreshTokenRepositoryImpl struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{db: db}
}

func (repo *refreshTokenRepositoryImpl) Create(ctx context.Context, token *models.RefreshToken) error {
	if err := repo.db.WithContext(ctx).Create(token).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to create refresh token: %v", err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to create refresh token", err)
	}
	return nil
}

func (repo *refreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := repo.db.WithContext(ctx).Where("refresh_token = ? and expired_at > ?", token, time.Now().Unix()).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.ErrNotFound, 1001, "Refresh token not found or expired")
		}
		logger.WithContext(ctx).Errorf("DB error: failed to fetch refresh token: %v", err)
		return nil, apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to fetch refresh token", err)
	}
	return &refreshToken, nil
}

func (repo *refreshTokenRepositoryImpl) Update(ctx context.Context, token *models.RefreshToken) error {
	if err := repo.db.WithContext(ctx).Save(token).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to update refresh token: %v", err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to update refresh token", err)
	}
	return nil
}

func (repo *refreshTokenRepositoryImpl) UpdateWithTx(ctx context.Context, token *models.RefreshToken, tx *gorm.DB) error {
	if err := tx.WithContext(ctx).Save(token).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to update refresh token with tx: %v", err)
		return apperror.Wrap(apperror.ErrInternalServer, 500, "Failed to update refresh token", err)
	}
	return nil
}

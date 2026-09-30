package repositories

import (
	"context"
	"net/http"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type SettingRepository interface {
	// GetValues returns the values of the given keys. Keys that do not exist
	// in the settings table are absent from the returned map.
	GetValues(ctx context.Context, keys ...string) (map[string]string, error)
}

type settingRepositoryImpl struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepositoryImpl{db: db}
}

func (repo *settingRepositoryImpl) GetValues(ctx context.Context, keys ...string) (map[string]string, error) {
	var settings []models.Setting
	if err := repo.db.WithContext(ctx).Where("`key` IN ?", keys).Find(&settings).Error; err != nil {
		logger.WithContext(ctx).Errorf("DB error: failed to fetch settings: %v", err)
		return nil, apperror.Wrap(http.StatusInternalServerError, apperror.ErrInternalServer, "Failed to fetch settings", err)
	}

	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	return values, nil
}

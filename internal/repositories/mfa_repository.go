package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/gorm"
)

// IMfaRepository defines operations for MFA settings database access
type IMfaRepository interface {
	// GetMfaSettingsByUserID retrieves MFA settings for a user by their ID
	GetMfaSettingsByUserID(userID uint) (*models.MfaSettings, error)
	// CreateMfaSettings creates a new MFA settings record for a user
	CreateMfaSettings(settings *models.MfaSettings) error
	// UpdateMfaSettings updates existing MFA settings
	UpdateMfaSettings(settings *models.MfaSettings) error
	// DeleteMfaSettings deletes MFA settings for a user (soft delete)
	DeleteMfaSettings(userID uint) error
}

type MfaRepository struct {
	db *gorm.DB
}

// NewMfaRepository creates a new instance of MfaRepository
func NewMfaRepository(db *gorm.DB) IMfaRepository {
	return &MfaRepository{
		db: db,
	}
}

// GetMfaSettingsByUserID retrieves MFA settings for a user by their ID
func (r *MfaRepository) GetMfaSettingsByUserID(userID uint) (*models.MfaSettings, error) {
	var settings models.MfaSettings
	if err := r.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

// CreateMfaSettings creates a new MFA settings record for a user
func (r *MfaRepository) CreateMfaSettings(settings *models.MfaSettings) error {
	return r.db.Create(settings).Error
}

// UpdateMfaSettings updates existing MFA settings
func (r *MfaRepository) UpdateMfaSettings(settings *models.MfaSettings) error {
	return r.db.Save(settings).Error
}

// DeleteMfaSettings deletes MFA settings for a user (soft delete)
func (r *MfaRepository) DeleteMfaSettings(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.MfaSettings{}).Error
}

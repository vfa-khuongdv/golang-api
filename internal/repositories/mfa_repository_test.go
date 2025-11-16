package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.MfaSettings{})
	return db
}

func TestMfaRepository_CreateMfaSettings(t *testing.T) {
	db := setupTestDB()
	repo := NewMfaRepository(db)

	secret := "test-secret"
	settings := &models.MfaSettings{
		UserID:     1,
		MfaEnabled: false,
		TotpSecret: &secret,
	}

	err := repo.CreateMfaSettings(settings)
	assert.NoError(t, err)
	assert.Greater(t, settings.ID, uint(0))
}

func TestMfaRepository_GetMfaSettingsByUserID(t *testing.T) {
	db := setupTestDB()
	repo := NewMfaRepository(db)

	secret := "test-secret"
	settings := &models.MfaSettings{
		UserID:     1,
		MfaEnabled: true,
		TotpSecret: &secret,
	}
	db.Create(settings)

	retrieved, err := repo.GetMfaSettingsByUserID(1)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, uint(1), retrieved.UserID)
	assert.True(t, retrieved.MfaEnabled)
}

func TestMfaRepository_UpdateMfaSettings(t *testing.T) {
	db := setupTestDB()
	repo := NewMfaRepository(db)

	settings := &models.MfaSettings{
		UserID:     1,
		MfaEnabled: false,
	}
	db.Create(settings)

	settings.MfaEnabled = true
	err := repo.UpdateMfaSettings(settings)
	assert.NoError(t, err)

	retrieved, _ := repo.GetMfaSettingsByUserID(1)
	assert.True(t, retrieved.MfaEnabled)
}

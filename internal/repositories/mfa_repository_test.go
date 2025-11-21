package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err := db.AutoMigrate(&models.MfaSettings{}); err != nil {
		panic(err)
	}
	return db
}

func TestMfaRepository(t *testing.T) {
	t.Run("CreateMfaSettings", func(t *testing.T) {
		db := setupTestDB()
		repo := NewMfaRepository(db)

		secret := "test-secret"
		settings := &models.MfaSettings{
			UserID:     1,
			MfaEnabled: false,
			TotpSecret: &secret,
		}

		err := repo.CreateMfaSettings(settings)
		require.NoError(t, err)
		assert.Greater(t, settings.ID, uint(0))
	})

	t.Run("GetMfaSettingsByUserID", func(t *testing.T) {
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
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, uint(1), retrieved.UserID)
		assert.True(t, retrieved.MfaEnabled)
	})

	t.Run("UpdateMfaSettings", func(t *testing.T) {
		db := setupTestDB()
		repo := NewMfaRepository(db)

		settings := &models.MfaSettings{
			UserID:     1,
			MfaEnabled: false,
		}
		db.Create(settings)

		settings.MfaEnabled = true
		err := repo.UpdateMfaSettings(settings)
		require.NoError(t, err)

		retrieved, _ := repo.GetMfaSettingsByUserID(1)
		assert.True(t, retrieved.MfaEnabled)
	})
}

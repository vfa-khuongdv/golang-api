package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSettingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Setting{}))
	return db
}

func TestSettingRepository(t *testing.T) {
	t.Run("GetValues - Returns Requested Keys", func(t *testing.T) {
		// Arrange
		db := setupSettingTestDB(t)
		require.NoError(t, db.Create(&[]models.Setting{
			{Key: "mail.host", Value: "smtp.example.com"},
			{Key: "mail.port", Value: "587"},
			{Key: "app.frontend_url", Value: "https://example.com"},
		}).Error)
		repo := repositories.NewSettingRepository(db)

		// Act
		values, err := repo.GetValues(context.Background(), "mail.host", "mail.port")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"mail.host": "smtp.example.com", "mail.port": "587"}, values)
	})

	t.Run("GetValues - Missing Keys Are Absent", func(t *testing.T) {
		// Arrange
		db := setupSettingTestDB(t)
		require.NoError(t, db.Create(&models.Setting{Key: "mail.host", Value: "smtp.example.com"}).Error)
		repo := repositories.NewSettingRepository(db)

		// Act
		values, err := repo.GetValues(context.Background(), "mail.host", "unknown_key")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"mail.host": "smtp.example.com"}, values)
	})

	t.Run("GetValues - Empty Value Is Preserved", func(t *testing.T) {
		// Arrange
		db := setupSettingTestDB(t)
		require.NoError(t, db.Create(&models.Setting{Key: "mail.from", Value: ""}).Error)
		repo := repositories.NewSettingRepository(db)

		// Act
		values, err := repo.GetValues(context.Background(), "mail.from")

		// Assert
		require.NoError(t, err)
		value, ok := values["mail.from"]
		assert.True(t, ok)
		assert.Empty(t, value)
	})

	t.Run("GetValues - DB Error", func(t *testing.T) {
		// Arrange
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		repo := repositories.NewSettingRepository(db) // settings table not migrated

		// Act
		values, err := repo.GetValues(context.Background(), "mail.host")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, values)
		assert.Contains(t, err.Error(), "Failed to fetch settings")
	})
}

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

	t.Run("SetValues - Inserts New And Updates Existing Keys", func(t *testing.T) {
		// Arrange
		db := setupSettingTestDB(t)
		require.NoError(t, db.Create(&models.Setting{Key: "mail.host", Value: "old.example.com"}).Error)
		repo := repositories.NewSettingRepository(db)

		// Act
		err := repo.SetValues(context.Background(), map[string]string{
			"mail.host": "smtp.example.com",
			"mail.port": "587",
		})

		// Assert
		require.NoError(t, err)
		values, err := repo.GetValues(context.Background(), "mail.host", "mail.port")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"mail.host": "smtp.example.com", "mail.port": "587"}, values)

		var count int64
		db.Model(&models.Setting{}).Where("`key` = ?", "mail.host").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("SetValues - Empty Value Is Stored", func(t *testing.T) {
		// Arrange
		db := setupSettingTestDB(t)
		require.NoError(t, db.Create(&models.Setting{Key: "mail.username", Value: "mailer"}).Error)
		repo := repositories.NewSettingRepository(db)

		// Act
		err := repo.SetValues(context.Background(), map[string]string{"mail.username": ""})

		// Assert
		require.NoError(t, err)
		values, err := repo.GetValues(context.Background(), "mail.username")
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"mail.username": ""}, values)
	})

	t.Run("SetValues - No Values Is A No-Op", func(t *testing.T) {
		// Arrange
		repo := repositories.NewSettingRepository(setupSettingTestDB(t))

		// Act
		err := repo.SetValues(context.Background(), map[string]string{})

		// Assert
		assert.NoError(t, err)
	})

	t.Run("SetValues - DB Error", func(t *testing.T) {
		// Arrange
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		repo := repositories.NewSettingRepository(db) // settings table not migrated

		// Act
		err = repo.SetValues(context.Background(), map[string]string{"mail.host": "x"})

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to update settings")
	})
}

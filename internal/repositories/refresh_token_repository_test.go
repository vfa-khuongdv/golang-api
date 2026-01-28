package repositories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NotNil(t, db)

	// Auto-migrate the models
	err = db.AutoMigrate(&models.RefreshToken{})
	require.NoError(t, err)

	return db
}

func TestRefreshTokenRepository(t *testing.T) {
	t.Run("Create - Success", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		item := &models.RefreshToken{
			RefreshToken: "test_refresh_token",
			IpAddress:    "127.0.0.1",
			UsedCount:    0,
			ExpiredAt:    1710000000,
			UserID:       1,
		}

		// Act
		err := repo.Create(item)

		// Assert
		require.NoError(t, err)
		assert.NotEqual(t, uint(0), item.ID)
	})

	t.Run("Create - Duplicate Token Error", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		token1 := &models.RefreshToken{
			RefreshToken: "duplicate_token",
			UserID:       1,
		}
		token2 := &models.RefreshToken{
			RefreshToken: "duplicate_token",
			UserID:       2,
		}

		// Act
		err1 := repo.Create(token1)
		require.NoError(t, err1)

		err2 := repo.Create(token2)

		// Assert
		assert.Error(t, err2, "Expected error due to duplicate token")
	})

	t.Run("First - Success", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		item := &models.RefreshToken{
			RefreshToken: "test_refresh_token_1",
			IpAddress:    "127.0.0.1",
			UsedCount:    0,
			ExpiredAt:    1710000000,
			UserID:       1,
		}
		err := repo.Create(item)
		require.NoError(t, err)

		// Act
		foundItem, err := repo.First(item.RefreshToken)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundItem)
		assert.Equal(t, item.RefreshToken, foundItem.RefreshToken)
		assert.Equal(t, item.IpAddress, foundItem.IpAddress)
		assert.Equal(t, item.UserID, foundItem.UserID)
		assert.Equal(t, item.UsedCount, foundItem.UsedCount)
	})

	t.Run("First - Not Found", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		// Act
		foundItem, err := repo.First("non_existent_token")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, foundItem)
	})

	t.Run("FindByToken - Not Expired Token Success", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		now := time.Now().Unix() + int64(time.Minute)
		item := &models.RefreshToken{
			RefreshToken: "test_refresh_token",
			IpAddress:    "127.0.0.1",
			UsedCount:    0,
			ExpiredAt:    now,
			UserID:       1,
		}
		err := repo.Create(item)
		require.NoError(t, err)

		// Act
		foundItem, err := repo.FindByToken(item.RefreshToken)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, foundItem)
		assert.Equal(t, item.RefreshToken, foundItem.RefreshToken)
		assert.Equal(t, item.IpAddress, foundItem.IpAddress)
		assert.Equal(t, item.UserID, foundItem.UserID)
		assert.Equal(t, item.UsedCount, foundItem.UsedCount)
		assert.Equal(t, item.ExpiredAt, foundItem.ExpiredAt)
	})

	t.Run("FindByToken - Expired Token Error", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		now := time.Now().Unix() - int64(time.Minute)
		item := &models.RefreshToken{
			RefreshToken: "test_refresh_token_expired",
			IpAddress:    "127.0.0.1",
			UsedCount:    0,
			ExpiredAt:    now,
			UserID:       1,
		}
		err := repo.Create(item)
		require.NoError(t, err)

		// Act
		foundItem, err := repo.FindByToken(item.RefreshToken)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, foundItem)
	})

	t.Run("Update - Success", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		item := &models.RefreshToken{
			RefreshToken: "test_original_refresh_token",
			IpAddress:    "",
			UsedCount:    0,
			ExpiredAt:    1710000000,
			UserID:       1,
		}
		err := repo.Create(item)
		require.NoError(t, err)

		// Update fields
		item.IpAddress = "127.0.0.1"
		item.RefreshToken = "test_updated_refresh_token"
		item.UsedCount = 1
		item.ExpiredAt = time.Now().Unix() + int64(time.Hour)

		// Act
		err = repo.Update(item)

		// Assert
		require.NoError(t, err)

		// Verify the update
		foundItem, err := repo.FindByToken(item.RefreshToken)
		require.NoError(t, err)
		require.NotNil(t, foundItem)
		assert.Equal(t, item.RefreshToken, foundItem.RefreshToken)
		assert.Equal(t, item.IpAddress, foundItem.IpAddress)
		assert.Equal(t, item.UserID, foundItem.UserID)
		assert.Equal(t, item.UsedCount, foundItem.UsedCount)
		assert.Equal(t, item.ExpiredAt, foundItem.ExpiredAt)
	})
}

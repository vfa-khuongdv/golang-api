package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"gorm.io/driver/mysql"
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
			ExpiredAt:    1710000000,
			UserID:       1,
		}

		// Act
		err := repo.Create(context.Background(), item)

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
		err1 := repo.Create(context.Background(), token1)
		require.NoError(t, err1)

		err2 := repo.Create(context.Background(), token2)

		// Assert
		assert.Error(t, err2, "Expected error due to duplicate token")
	})

	t.Run("UpdateWithTx - Success", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		now := time.Now().Unix() + int64(time.Hour)
		item := &models.RefreshToken{
			RefreshToken: "test_tx_token",
			IpAddress:    "127.0.0.1",
			ExpiredAt:    now,
			UserID:       1,
		}
		err := repo.Create(context.Background(), item)
		require.NoError(t, err)

		// Act - use transaction
		tx := db.Begin()
		require.NotNil(t, tx)

		err = repo.UpdateWithTx(context.Background(), tx, item)

		// Assert
		require.NoError(t, err)
		tx.Commit()

		var foundItem models.RefreshToken
		require.NoError(t, db.Where("refresh_token = ?", "test_tx_token").First(&foundItem).Error)
	})

	t.Run("FindByTokenWithTx - Success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		item := &models.RefreshToken{
			RefreshToken: "token_tx_success",
			IpAddress:    "127.0.0.1",
			ExpiredAt:    time.Now().Unix() + int64(time.Hour),
			UserID:       1,
		}
		err := repo.Create(context.Background(), item)
		require.NoError(t, err)

		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		found, err := repo.FindByTokenWithTx(context.Background(), tx, item.RefreshToken)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, item.RefreshToken, found.RefreshToken)
	})

	t.Run("FindByTokenWithTx - Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		found, err := repo.FindByTokenWithTx(context.Background(), tx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("FindByTokenWithTx - Expired", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		item := &models.RefreshToken{
			RefreshToken: "token_tx_expired",
			IpAddress:    "127.0.0.1",
			ExpiredAt:    time.Now().Unix() - int64(time.Minute),
			UserID:       1,
		}
		err := repo.Create(context.Background(), item)
		require.NoError(t, err)

		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		found, err := repo.FindByTokenWithTx(context.Background(), tx, item.RefreshToken)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("FindByTokenWithTx - Database Error", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		_ = db.Callback().Query().Before("gorm:query").Register("force_token_tx_db_error", func(tx2 *gorm.DB) {
			_ = tx2.AddError(assert.AnError)
		})
		defer func() { _ = db.Callback().Query().Remove("force_token_tx_db_error") }()

		found, err := repo.FindByTokenWithTx(context.Background(), tx, "any_token")
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("UpdateWithTx - Database Error", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		item := &models.RefreshToken{
			RefreshToken: "updatetx_error_token",
			IpAddress:    "127.0.0.1",
			ExpiredAt:    time.Now().Unix() + int64(time.Hour),
			UserID:       1,
		}
		err := repo.Create(context.Background(), item)
		require.NoError(t, err)

		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		_ = db.Callback().Update().Before("gorm:update").Register("force_updatetx_db_error", func(tx2 *gorm.DB) {
			_ = tx2.AddError(assert.AnError)
		})
		defer func() { _ = db.Callback().Update().Remove("force_updatetx_db_error") }()

		err = repo.UpdateWithTx(context.Background(), tx, item)
		assert.Error(t, err)
	})

	t.Run("BeginTx - Success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		tx, err := repo.BeginTx(context.Background())
		require.NoError(t, err)
		require.NotNil(t, tx)
		tx.Rollback()
	})

	t.Run("BeginTx - Database Error", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Close()
		require.NoError(t, err)

		tx, err := repo.BeginTx(context.Background())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("DeleteByUserID - Success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		item := &models.RefreshToken{
			RefreshToken: "delete_me",
			IpAddress:    "127.0.0.1",
			ExpiredAt:    time.Now().Unix() + int64(time.Hour),
			UserID:       1,
		}
		require.NoError(t, repo.Create(context.Background(), item))

		err := repo.DeleteByUserID(context.Background(), 1)
		assert.NoError(t, err)
	})

	t.Run("DeleteByUserID - No Tokens", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		err := repo.DeleteByUserID(context.Background(), 999)
		assert.NoError(t, err)
	})

	t.Run("DeleteByUserID - Database Error", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)

		sqlDB, err := db.DB()
		require.NoError(t, err)
		_ = sqlDB.Close()

		err = repo.DeleteByUserID(context.Background(), 1)
		assert.Error(t, err)
	})
}

func TestRefreshTokenRepository_HardDelete(t *testing.T) {
	countAll := func(t *testing.T, db *gorm.DB) int64 {
		var n int64
		require.NoError(t, db.Unscoped().Model(&models.RefreshToken{}).Count(&n).Error)
		return n
	}

	t.Run("DeleteByUserID hard-deletes so logout does not leave rows behind", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repositories.NewRefreshTokenRepository(db)
		require.NoError(t, repo.Create(context.Background(), &models.RefreshToken{RefreshToken: "a", ExpiredAt: time.Now().Unix() + 3600, UserID: 1}))
		require.NoError(t, repo.Create(context.Background(), &models.RefreshToken{RefreshToken: "b", ExpiredAt: time.Now().Unix() + 3600, UserID: 2}))

		require.NoError(t, repo.DeleteByUserID(context.Background(), 1))

		assert.Equal(t, int64(1), countAll(t, db))
	})
}

// A plain SELECT lets two concurrent refreshes with the same token both succeed.
// The row must be locked (SELECT ... FOR UPDATE) inside the rotation transaction.
// SQLite ignores FOR UPDATE, so the generated MySQL SQL is inspected instead.
func TestRefreshTokenRepository_FindByTokenWithTx_LocksRow(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "u:p@tcp(127.0.0.1:1)/d",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	var sql string
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("capture_sql", func(d *gorm.DB) {
		sql = d.Statement.SQL.String()
	}))

	repo := repositories.NewRefreshTokenRepository(db)
	_, _ = repo.FindByTokenWithTx(context.Background(), db, "some-token")

	assert.Contains(t, sql, "FOR UPDATE")
}

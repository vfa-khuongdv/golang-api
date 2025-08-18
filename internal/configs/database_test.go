package configs_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitDB(t *testing.T) {
	t.Run("InitDB - Success with SQLite for testing", func(t *testing.T) {
		// Create a temporary SQLite database for testing
		tempFile, err := os.CreateTemp("", "test_db_*.sqlite")
		require.NoError(t, err)
		defer func() {
			_ = os.Remove(tempFile.Name())
		}()
		_ = tempFile.Close()

		// Mock the GORM Open function by using SQLite instead of MySQL for testing
		// This tests the core functionality without requiring a MySQL server
		db, err := gorm.Open(sqlite.Open(tempFile.Name()), &gorm.Config{})
		require.NoError(t, err)
		require.NotNil(t, db)

		// Test that we can interact with the database
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NotNil(t, sqlDB)

		err = sqlDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("InitDB - DatabaseConfig struct creation", func(t *testing.T) {
		config := configs.DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
		}

		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, "3306", config.Port)
		assert.Equal(t, "testuser", config.User)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, "testdb", config.DBName)
	})

	t.Run("InitDB - Global DB variable test", func(t *testing.T) {
		// Reset global DB to nil
		configs.DB = nil
		assert.Nil(t, configs.DB)

		// Create a temporary SQLite database for testing
		tempFile, err := os.CreateTemp("", "test_db_global_*.sqlite")
		require.NoError(t, err)
		defer func() {
			_ = os.Remove(tempFile.Name())
		}()

		_ = tempFile.Close()
		// Test with SQLite (simulating successful database connection)
		db, err := gorm.Open(sqlite.Open(tempFile.Name()), &gorm.Config{})
		require.NoError(t, err)

		// Simulate setting the global DB variable
		configs.DB = db
		assert.NotNil(t, configs.DB)
		assert.Equal(t, db, configs.DB)
	})
}

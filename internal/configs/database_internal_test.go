package configs

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetDefaults(t *testing.T) {
	t.Run("ApplyDefaultsWhenZeroValues", func(t *testing.T) {
		cfg := DatabaseConfig{}
		setDefaults(&cfg)

		assert.Equal(t, DEFAULT_MAX_OPEN_CONNS, cfg.MaxOpenConns)
		assert.Equal(t, DEFAULT_MAX_IDLE_CONNS, cfg.MaxIdleConns)
		assert.Equal(t, DEFAULT_CONN_MAX_LIFETIME, cfg.ConnMaxLifetime)
		assert.Equal(t, DEFAULT_CONN_MAX_IDLE_TIME, cfg.ConnMaxIdleTime)
	})

	t.Run("KeepCustomValuesWhenSet", func(t *testing.T) {
		cfg := DatabaseConfig{
			MaxOpenConns:    120,
			MaxIdleConns:    40,
			ConnMaxLifetime: 45 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		}
		setDefaults(&cfg)

		assert.Equal(t, 120, cfg.MaxOpenConns)
		assert.Equal(t, 40, cfg.MaxIdleConns)
		assert.Equal(t, 45*time.Minute, cfg.ConnMaxLifetime)
		assert.Equal(t, 10*time.Minute, cfg.ConnMaxIdleTime)
	})
}

func TestPingDB(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := gdb.DB()
	require.NoError(t, err)

	t.Run("PingSuccess", func(t *testing.T) {
		err := pingDB(sqlDB)
		assert.NoError(t, err)
	})

	t.Run("PingFailureWhenClosed", func(t *testing.T) {
		require.NoError(t, sqlDB.Close())
		err := pingDB(sqlDB)
		assert.Error(t, err)
	})
}

func TestInitDB_InternalBranches(t *testing.T) {
	originalOpen := openGormConnection
	originalFatalf := logFatalf
	originalInfof := logInfof
	originalPing := pingDBFn
	t.Cleanup(func() {
		openGormConnection = originalOpen
		logFatalf = originalFatalf
		logInfof = originalInfof
		pingDBFn = originalPing
	})

	config := DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "user",
		Password: "pass",
		DBName:   "db",
	}

	t.Run("OpenConnectionFailure", func(t *testing.T) {
		openGormConnection = func(_ string) (*gorm.DB, error) {
			return nil, errors.New("open failed")
		}
		logFatalf = func(_ string, _ ...interface{}) {
			panic("fatal-open")
		}

		assert.PanicsWithValue(t, "fatal-open", func() {
			_ = InitDB(config)
		})
	})

	t.Run("PingFailure", func(t *testing.T) {
		gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		openGormConnection = func(_ string) (*gorm.DB, error) {
			return gdb, nil
		}
		pingDBFn = func(_ *sql.DB) error {
			return errors.New("ping failed")
		}
		logFatalf = func(_ string, _ ...interface{}) {
			panic("fatal-ping")
		}

		assert.PanicsWithValue(t, "fatal-ping", func() {
			_ = InitDB(config)
		})
	})

	t.Run("Success", func(t *testing.T) {
		gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		require.NoError(t, err)

		infoCalled := false
		openGormConnection = func(_ string) (*gorm.DB, error) {
			return gdb, nil
		}
		pingDBFn = func(_ *sql.DB) error {
			return nil
		}
		logFatalf = func(_ string, _ ...interface{}) {
			panic("should-not-fatal")
		}
		logInfof = func(_ string, _ ...interface{}) {
			infoCalled = true
		}

		result := InitDB(config)
		assert.NotNil(t, result)
		assert.Equal(t, gdb, result)
		assert.Equal(t, gdb, DB)
		assert.True(t, infoCalled)
	})
}

package configs_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
)

func TestLoad(t *testing.T) {
	// Save and restore the environment variables Load() depends on so the
	// subtests don't leak mutations into the rest of the test process.
	envKeys := []string{"PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE", "JWT_KEY"}
	original := make(map[string]string, len(envKeys))
	for _, k := range envKeys {
		if v, ok := os.LookupEnv(k); ok {
			original[k] = v
		}
	}
	t.Cleanup(func() {
		for _, k := range envKeys {
			if v, ok := original[k]; ok {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})

	t.Run("Load - With .env file", func(t *testing.T) {
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(originalDir) })

		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		require.NoError(t, err)

		envContent := `PORT=4000
DB_USERNAME=testuser
DB_PASSWORD=testpass
DB_DATABASE=testdb
JWT_KEY=this-is-a-long-enough-secret-key-32-chars!!`
		err = os.WriteFile(".env", []byte(envContent), 0644)
		require.NoError(t, err)

		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_USERNAME")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_DATABASE")
		_ = os.Unsetenv("JWT_KEY")

		cfg, err := configs.Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "4000", cfg.Server.Port)
		assert.Equal(t, "testuser", cfg.Database.User)
		assert.Equal(t, "testpass", cfg.Database.Password)
		assert.Equal(t, "testdb", cfg.Database.DBName)
		assert.Equal(t, "this-is-a-long-enough-secret-key-32-chars!!", cfg.JWT.Secret)
	})

	t.Run("Load - Missing required vars returns error", func(t *testing.T) {
		_ = os.Unsetenv("DB_USERNAME")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_DATABASE")
		_ = os.Unsetenv("JWT_KEY")
		_ = os.Setenv("PORT", "3000")

		cfg, err := configs.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Load - Missing multiple required vars", func(t *testing.T) {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_USERNAME")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_DATABASE")
		_ = os.Unsetenv("JWT_KEY")

		cfg, err := configs.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "DB_USERNAME")
		assert.Contains(t, err.Error(), "DB_PASSWORD")
		assert.Contains(t, err.Error(), "DB_DATABASE")
		assert.Contains(t, err.Error(), "JWT_KEY")
	})

	t.Run("Load - Empty PORT env var", func(t *testing.T) {
		_ = os.Setenv("PORT", "")
		_ = os.Setenv("DB_USERNAME", "u")
		_ = os.Setenv("DB_PASSWORD", "p")
		_ = os.Setenv("DB_DATABASE", "d")
		_ = os.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-testing-32chars")

		cfg, err := configs.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "PORT")
	})

	t.Run("Load - Uses system env as fallback", func(t *testing.T) {
		_ = os.Setenv("PORT", "5000")
		_ = os.Setenv("DB_USERNAME", "sysuser")
		_ = os.Setenv("DB_PASSWORD", "syspass")
		_ = os.Setenv("DB_DATABASE", "sysdb")
		_ = os.Setenv("JWT_KEY", "system-wide-secret-key-that-is-long-enough")

		cfg, err := configs.Load()
		require.NoError(t, err)
		assert.Equal(t, "5000", cfg.Server.Port)
		assert.Equal(t, "sysuser", cfg.Database.User)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("GetEnv", func(t *testing.T) {
		key := "TEST_ENV_VAR"
		defaultVal := "default"

		_ = os.Unsetenv(key)
		val := configs.GetEnv(key, defaultVal)
		assert.Equal(t, defaultVal, val)

		expectedVal := "value123"
		_ = os.Setenv(key, expectedVal)
		val = configs.GetEnv(key, defaultVal)
		assert.Equal(t, expectedVal, val)

		_ = os.Unsetenv(key)
	})

	t.Run("GetEnvAsInt", func(t *testing.T) {
		key := "TEST_ENV_INT"
		defaultVal := 42

		_ = os.Unsetenv(key)
		val := configs.GetEnvAsInt(key, defaultVal)
		assert.Equal(t, defaultVal, val)

		_ = os.Setenv(key, "100")
		val = configs.GetEnvAsInt(key, defaultVal)
		assert.Equal(t, 100, val)

		_ = os.Setenv(key, "not_an_int")
		val = configs.GetEnvAsInt(key, defaultVal)
		assert.Equal(t, defaultVal, val)

		_ = os.Unsetenv(key)
	})
}

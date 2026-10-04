package configs_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
)

func TestLoad(t *testing.T) {
	// Save and restore the environment variables Load() depends on so the
	// subtests don't leak mutations into the rest of the test process.
	envKeys := []string{"PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE", "JWT_KEY", "SETTINGS_ENCRYPTION_KEY"}
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
JWT_KEY=this-is-a-long-enough-secret-key-32-chars!!
SETTINGS_ENCRYPTION_KEY=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`
		err = os.WriteFile(".env", []byte(envContent), 0644)
		require.NoError(t, err)

		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DB_USERNAME")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_DATABASE")
		_ = os.Unsetenv("JWT_KEY")
		_ = os.Unsetenv("SETTINGS_ENCRYPTION_KEY")

		cfg, err := configs.Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "4000", cfg.Server.Port)
		assert.Equal(t, "testuser", cfg.Database.User)
		assert.Equal(t, "testpass", cfg.Database.Password)
		assert.Equal(t, "testdb", cfg.Database.DBName)
		assert.Equal(t, "this-is-a-long-enough-secret-key-32-chars!!", cfg.JWT.Secret)
		assert.Equal(t, strings.Repeat("a", 40), cfg.Settings.EncryptionKey)
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
		_ = os.Unsetenv("SETTINGS_ENCRYPTION_KEY")

		cfg, err := configs.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "DB_USERNAME")
		assert.Contains(t, err.Error(), "DB_PASSWORD")
		assert.Contains(t, err.Error(), "DB_DATABASE")
		assert.Contains(t, err.Error(), "JWT_KEY")
		assert.Contains(t, err.Error(), "SETTINGS_ENCRYPTION_KEY")
	})

	t.Run("Load - Missing SETTINGS_ENCRYPTION_KEY", func(t *testing.T) {
		_ = os.Setenv("PORT", "3000")
		_ = os.Setenv("DB_USERNAME", "u")
		_ = os.Setenv("DB_PASSWORD", "p")
		_ = os.Setenv("DB_DATABASE", "d")
		_ = os.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-testing-32chars")
		_ = os.Setenv("SETTINGS_ENCRYPTION_KEY", "   ")

		cfg, err := configs.Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "SETTINGS_ENCRYPTION_KEY")
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
		_ = os.Setenv("SETTINGS_ENCRYPTION_KEY", strings.Repeat("b", 40))

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

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("PORT", "3000")
	t.Setenv("DB_USERNAME", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_DATABASE", "d")
	t.Setenv("JWT_KEY", "this-is-a-very-long-secret-key-for-testing-32chars")
	t.Setenv("SETTINGS_ENCRYPTION_KEY", strings.Repeat("a", 40))
}

func TestLoadDatabasePool(t *testing.T) {
	t.Run("defaults are sized for several tasks sharing one small RDS", func(t *testing.T) {
		setRequiredEnv(t)
		for _, k := range []string{"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME"} {
			t.Setenv(k, "")
			_ = os.Unsetenv(k)
		}

		cfg, err := configs.Load()
		require.NoError(t, err)
		assert.Equal(t, 20, cfg.Database.MaxOpenConns)
		assert.Equal(t, 5, cfg.Database.MaxIdleConns)
		assert.Equal(t, 30*time.Minute, cfg.Database.ConnMaxLifetime)
		assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxIdleTime)
	})

	t.Run("pool settings can be overridden from env", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("DB_MAX_OPEN_CONNS", "12")
		t.Setenv("DB_MAX_IDLE_CONNS", "3")
		t.Setenv("DB_CONN_MAX_LIFETIME", "10m")
		t.Setenv("DB_CONN_MAX_IDLE_TIME", "90s")

		cfg, err := configs.Load()
		require.NoError(t, err)
		assert.Equal(t, 12, cfg.Database.MaxOpenConns)
		assert.Equal(t, 3, cfg.Database.MaxIdleConns)
		assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxLifetime)
		assert.Equal(t, 90*time.Second, cfg.Database.ConnMaxIdleTime)
	})

	t.Run("invalid duration falls back to the default", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("DB_CONN_MAX_LIFETIME", "soon")

		cfg, err := configs.Load()
		require.NoError(t, err)
		assert.Equal(t, 30*time.Minute, cfg.Database.ConnMaxLifetime)
	})
}

func TestLoadTrustedProxies(t *testing.T) {
	t.Run("Unset Trusts Every Peer And Is Marked As The Default", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("TRUSTED_PROXIES", "")
		require.NoError(t, os.Unsetenv("TRUSTED_PROXIES"))

		cfg, err := configs.Load()

		require.NoError(t, err)
		assert.Equal(t, []string{"0.0.0.0/0"}, cfg.Server.TrustedProxies)
		assert.True(t, cfg.Server.TrustedProxiesDefault)
	})

	t.Run("Empty Trusts Nothing", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("TRUSTED_PROXIES", "")

		cfg, err := configs.Load()

		require.NoError(t, err)
		assert.Empty(t, cfg.Server.TrustedProxies)
		assert.False(t, cfg.Server.TrustedProxiesDefault)
	})

	t.Run("A List Is Split And Trimmed", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 192.168.0.0/16")

		cfg, err := configs.Load()

		require.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.0/8", "192.168.0.0/16"}, cfg.Server.TrustedProxies)
		assert.False(t, cfg.Server.TrustedProxiesDefault)
	})
}

func TestLoadCORSAllowedOrigins(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example.com, https://b.example.com")

	cfg, err := configs.Load()

	require.NoError(t, err)
	assert.Equal(t, []string{"https://a.example.com", "https://b.example.com"}, cfg.CORS.AllowedOrigins)
}

func TestLoadRejectsShortSecrets(t *testing.T) {
	for _, key := range []string{"JWT_KEY", "SETTINGS_ENCRYPTION_KEY"} {
		t.Run(key, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(key, strings.Repeat("x", 31))

			_, err := configs.Load()

			assert.ErrorContains(t, err, key)
		})
	}
}

func TestLoadCORSAllowedOriginsDefault(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	require.NoError(t, os.Unsetenv("CORS_ALLOWED_ORIGINS"))

	cfg, err := configs.Load()

	require.NoError(t, err)
	assert.Equal(t, []string{"http://localhost:5173"}, cfg.CORS.AllowedOrigins)
}

func TestLoadRejectsPlaceholderSecretsInProd(t *testing.T) {
	for _, key := range []string{"JWT_KEY", "SETTINGS_ENCRYPTION_KEY"} {
		t.Run(key, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("STAGE", "prod")
			t.Setenv(key, "change-me-to-a-secret-at-least-32-chars-long!!")

			_, err := configs.Load()

			assert.ErrorContains(t, err, key)
		})
	}
}

func TestLoadAllowsPlaceholderSecretsOutsideProd(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("STAGE", "dev")
	t.Setenv("JWT_KEY", "change-me-to-a-secret-at-least-32-chars-long!!")

	_, err := configs.Load()

	assert.NoError(t, err)
}

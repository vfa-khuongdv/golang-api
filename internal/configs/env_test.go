package configs_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
)

func TestLoadEnv(t *testing.T) {
	t.Run("LoadEnv - With .env file", func(t *testing.T) {
		// Save current working directory
		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "env_test")
		assert.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		// Change to the temporary directory
		err = os.Chdir(tempDir)
		assert.NoError(t, err)

		// Create a .env file in the temporary directory
		envContent := `TEST_VAR=test_value
ANOTHER_VAR=another_value`

		err = os.WriteFile(".env", []byte(envContent), 0644)
		assert.NoError(t, err)

		// Clean up any existing environment variables
		_ = os.Unsetenv("TEST_VAR")
		_ = os.Unsetenv("ANOTHER_VAR")

		// Call LoadEnv
		configs.LoadEnv()

		// Verify that environment variables are loaded
		testVar := os.Getenv("TEST_VAR")
		anotherVar := os.Getenv("ANOTHER_VAR")

		assert.Equal(t, "test_value", testVar)
		assert.Equal(t, "another_value", anotherVar)

		// Clean up environment variables
		_ = os.Unsetenv("TEST_VAR")
		_ = os.Unsetenv("ANOTHER_VAR")
	})

	t.Run("LoadEnv - Without .env file", func(t *testing.T) {
		// Create a temporary directory without .env file
		tempDir, err := os.MkdirTemp("", "no_env_test")
		assert.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		originalDir, err := os.Getwd()
		assert.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		err = os.Chdir(tempDir)
		assert.NoError(t, err)

		// Set a system environment variable
		_ = os.Setenv("SYSTEM_TEST_VAR", "system_value")
		defer func() {
			_ = os.Unsetenv("SYSTEM_TEST_VAR")
		}()

		// Call LoadEnv - should not panic and should use system env vars
		configs.LoadEnv()

		// Verify that system environment variable is still available
		systemVar := os.Getenv("SYSTEM_TEST_VAR")
		assert.Equal(t, "system_value", systemVar)
	})

	t.Run("LoadEnv - Function doesn't panic", func(t *testing.T) {
		// Test that LoadEnv doesn't panic regardless of file presence
		assert.NotPanics(t, func() {
			configs.LoadEnv()
		})
	})

	t.Run("LoadEnv - Multiple calls", func(t *testing.T) {
		// Test that multiple calls to LoadEnv don't cause issues
		assert.NotPanics(t, func() {
			configs.LoadEnv()
			configs.LoadEnv()
			configs.LoadEnv()
		})
	})
}

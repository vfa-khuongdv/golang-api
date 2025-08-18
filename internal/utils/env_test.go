package utils_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func TestGetEnv(t *testing.T) {
	key := "TEST_ENV_VAR"
	defaultVal := "default"

	// Ensure env var is not set initially
	os.Unsetenv(key)
	val := utils.GetEnv(key, defaultVal)
	assert.Equal(t, defaultVal, val, "Expected default value when env var is not set")

	// Set env var and test retrieval
	expectedVal := "value123"
	os.Setenv(key, expectedVal)
	val = utils.GetEnv(key, defaultVal)
	assert.Equal(t, expectedVal, val, "Expected value from environment variable")

	// Cleanup
	os.Unsetenv(key)
}

func TestGetEnvAsInt(t *testing.T) {
	key := "TEST_ENV_INT"
	defaultVal := 42

	// Env var not set -> should return default
	os.Unsetenv(key)
	val := utils.GetEnvAsInt(key, defaultVal)
	assert.Equal(t, defaultVal, val, "Expected default int value when env var is not set")

	// Env var set with valid int string
	os.Setenv(key, "100")
	val = utils.GetEnvAsInt(key, defaultVal)
	assert.Equal(t, 1001, val, "Expected parsed int value from environment variable")

	// Env var set with invalid int string -> should return default
	os.Setenv(key, "not_an_int")
	val = utils.GetEnvAsInt(key, defaultVal)
	assert.Equal(t, defaultVal, val, "Expected default int value when env var is invalid")

	// Cleanup
	os.Unsetenv(key)
}

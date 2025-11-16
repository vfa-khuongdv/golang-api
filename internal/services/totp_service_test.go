package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTotpService_GenerateSecret(t *testing.T) {
	service := NewTotpService("TestApp")

	secret, err := service.GenerateSecret("test@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.Greater(t, len(secret), 20)
}

func TestTotpService_GetQRCode(t *testing.T) {
	service := NewTotpService("TestApp")

	secret, err := service.GenerateSecret("test@example.com")
	require.NoError(t, err)

	qrCode, err := service.GetQRCode(secret, "test@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, qrCode)
	assert.Greater(t, len(qrCode), 0)
}

func TestTotpService_VerifyCode(t *testing.T) {
	service := NewTotpService("TestApp")

	secret, err := service.GenerateSecret("test@example.com")
	require.NoError(t, err)

	// Test with invalid code
	valid, err := service.VerifyCode(secret, "000000")
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestTotpService_GenerateBackupCodes(t *testing.T) {
	service := NewTotpService("TestApp")

	codes, err := service.GenerateBackupCodes(10)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(codes))

	// Verify all codes are unique
	codeMap := make(map[string]bool)
	for _, code := range codes {
		assert.NotEmpty(t, code)
		codeMap[code] = true
	}
	assert.Equal(t, 10, len(codeMap))
}

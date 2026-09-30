package utils_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

// Low-entropy dummy key so secret scanners do not flag it.
var testSecretKey = strings.Repeat("a", 40)

func TestEncryptDecryptSecret(t *testing.T) {
	t.Run("Roundtrip", func(t *testing.T) {
		for _, plaintext := range []string{"smtp-password", "", "パスワード 🔑", strings.Repeat("x", 4096)} {
			encrypted, err := utils.EncryptSecret(testSecretKey, plaintext)
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(encrypted, utils.EncryptedPrefix))
			if plaintext != "" {
				assert.NotContains(t, encrypted, plaintext)
			}

			decrypted, err := utils.DecryptSecret(testSecretKey, encrypted)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		}
	})

	t.Run("Random Nonce Produces Different Ciphertexts", func(t *testing.T) {
		first, err := utils.EncryptSecret(testSecretKey, "same")
		require.NoError(t, err)
		second, err := utils.EncryptSecret(testSecretKey, "same")
		require.NoError(t, err)
		assert.NotEqual(t, first, second)
	})

	t.Run("Key Exactly Min Length", func(t *testing.T) {
		key := strings.Repeat("k", utils.MinSecretKeyLength)
		encrypted, err := utils.EncryptSecret(key, "secret")
		require.NoError(t, err)
		decrypted, err := utils.DecryptSecret(key, encrypted)
		require.NoError(t, err)
		assert.Equal(t, "secret", decrypted)
	})

	t.Run("Key Too Short", func(t *testing.T) {
		key := strings.Repeat("k", utils.MinSecretKeyLength-1)
		_, err := utils.EncryptSecret(key, "secret")
		assert.ErrorIs(t, err, utils.ErrSecretKeyTooShort)
		_, err = utils.DecryptSecret(key, utils.EncryptedPrefix+"AAAA")
		assert.ErrorIs(t, err, utils.ErrSecretKeyTooShort)
	})

	t.Run("Wrong Key", func(t *testing.T) {
		encrypted, err := utils.EncryptSecret(testSecretKey, "secret")
		require.NoError(t, err)
		_, err = utils.DecryptSecret(testSecretKey+"-other", encrypted)
		assert.ErrorIs(t, err, utils.ErrDecryptFailed)
	})

	t.Run("Plaintext Value Rejected", func(t *testing.T) {
		_, err := utils.DecryptSecret(testSecretKey, "plain-password")
		assert.ErrorIs(t, err, utils.ErrNotEncrypted)
	})

	t.Run("Tampered Ciphertext", func(t *testing.T) {
		encrypted, err := utils.EncryptSecret(testSecretKey, "secret")
		require.NoError(t, err)
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encrypted, utils.EncryptedPrefix))
		require.NoError(t, err)
		raw[len(raw)-1] ^= 0xFF
		tampered := utils.EncryptedPrefix + base64.StdEncoding.EncodeToString(raw)

		_, err = utils.DecryptSecret(testSecretKey, tampered)
		assert.ErrorIs(t, err, utils.ErrDecryptFailed)
	})

	t.Run("Invalid Base64", func(t *testing.T) {
		_, err := utils.DecryptSecret(testSecretKey, utils.EncryptedPrefix+"!!not-base64!!")
		assert.ErrorIs(t, err, utils.ErrDecryptFailed)
	})

	t.Run("Payload Shorter Than Nonce", func(t *testing.T) {
		short := utils.EncryptedPrefix + base64.StdEncoding.EncodeToString([]byte("short"))
		_, err := utils.DecryptSecret(testSecretKey, short)
		assert.ErrorIs(t, err, utils.ErrDecryptFailed)
	})
}

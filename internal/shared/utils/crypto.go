package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

// EncryptedPrefix marks values produced by EncryptSecret, so encrypted and
// plaintext values can never be confused.
const EncryptedPrefix = "enc:v1:"

// MinSecretKeyLength is the minimum length of the key passed to EncryptSecret/DecryptSecret.
const MinSecretKeyLength = 32

var (
	ErrSecretKeyTooShort = errors.New("encryption key must be at least 32 characters")
	ErrNotEncrypted      = errors.New("value is not encrypted")
	ErrDecryptFailed     = errors.New("failed to decrypt value")
)

// EncryptSecret encrypts plaintext with AES-256-GCM using a key derived from
// secretKey. The result is "enc:v1:" + base64(nonce || ciphertext).
func EncryptSecret(secretKey, plaintext string) (string, error) {
	gcm, err := newGCM(secretKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return EncryptedPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptSecret reverses EncryptSecret. It fails if the value is not
// encrypted, was encrypted with another key, or has been tampered with.
func DecryptSecret(secretKey, value string) (string, error) {
	gcm, err := newGCM(secretKey)
	if err != nil {
		return "", err
	}

	encoded, ok := strings.CutPrefix(value, EncryptedPrefix)
	if !ok {
		return "", ErrNotEncrypted
	}

	sealed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(sealed) < gcm.NonceSize() {
		return "", ErrDecryptFailed
	}

	nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptFailed
	}
	return string(plaintext), nil
}

func newGCM(secretKey string) (cipher.AEAD, error) {
	if len(secretKey) < MinSecretKeyLength {
		return nil, ErrSecretKeyTooShort
	}

	key := sha256.Sum256([]byte(secretKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

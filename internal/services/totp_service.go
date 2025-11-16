package services

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image/png"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// ITotpService defines operations for TOTP (Time-based One-Time Password) functionality
type ITotpService interface {
	// GenerateSecret generates a new TOTP secret for a user
	GenerateSecret(email string) (secret string, err error)
	// GetQRCode generates a QR code image as PNG bytes for the given secret and email
	GetQRCode(secret string, email string) (qrCode []byte, err error)
	// VerifyCode verifies if the provided code is valid for the given secret
	VerifyCode(secret string, code string) (valid bool, err error)
	// GenerateBackupCodes generates a list of backup codes for account recovery
	GenerateBackupCodes(count int) (codes []string, err error)
}

type TotpService struct {
	issuer string
}

// NewTotpService creates a new instance of TotpService
func NewTotpService(issuer string) ITotpService {
	return &TotpService{
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (ts *TotpService) GenerateSecret(email string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      ts.issuer,
		AccountName: email,
		SecretSize:  32,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return key.Secret(), nil
}

// GetQRCode generates a QR code image as PNG bytes for the given secret and email
func (ts *TotpService) GetQRCode(secret string, email string) ([]byte, error) {
	// Create a key from the secret
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		ts.issuer, email, secret, ts.issuer))
	if err != nil {
		return nil, fmt.Errorf("failed to create key from secret: %w", err)
	}

	// Generate QR code image
	image, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Encode image to PNG
	buf := new(bytes.Buffer)
	err = png.Encode(buf, image)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// VerifyCode verifies if the provided code is valid for the given secret
func (ts *TotpService) VerifyCode(secret string, code string) (bool, error) {
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return valid, err
}

// GenerateBackupCodes generates a list of backup codes for account recovery
func (ts *TotpService) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate 8 random bytes
		randomBytes := make([]byte, 8)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}
		// Encode as base32 and take first 12 characters for readable backup code
		code := base64.StdEncoding.EncodeToString(randomBytes)
		codes[i] = code[:12]
	}
	return codes, nil
}

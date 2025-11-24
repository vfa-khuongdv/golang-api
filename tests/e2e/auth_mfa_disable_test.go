package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAuthMfaDisable(t *testing.T) {
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword := utils.HashPassword(password)

	// Generate a real TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GolangCMS",
		AccountName: "test_mfa_disable@example.com",
	})
	require.NoError(t, err)
	secret := key.Secret()

	user := models.User{
		Name:     "Test User MFA Disable",
		Email:    "test_mfa_disable@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	// Enable MFA for the user
	mfaSettings := models.MfaSettings{
		UserID:     user.ID,
		MfaEnabled: true,
		TotpSecret: &secret,
	}
	db.Create(&mfaSettings)

	// Login to get access token (simulating login before MFA was enforced or after MFA verification)
	// In a real scenario, if MFA is enabled, login returns a temp token.
	// But the disable endpoint requires an access token.
	// So we assume the user has fully authenticated (e.g. via recovery code or just standard flow if we bypass for test)
	// Actually, the login flow returns a temp token if MFA is enabled.
	// To get an access token, we need to verify MFA.

	// Login first
	loginPayload := map[string]string{
		"email":    "test_mfa_disable@example.com",
		"password": password,
	}
	payloadBytes, _ := json.Marshal(loginPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var mfaResponse services.MfaRequiredResponse
	err = json.Unmarshal(w.Body.Bytes(), &mfaResponse)
	require.NoError(t, err)
	tempToken := mfaResponse.TemporaryToken

	// Verify MFA to get access token
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	verifyPayload := map[string]string{
		"code": code,
	}
	payloadBytes, _ = json.Marshal(verifyPayload)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/mfa/verify-code", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tempToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResponse services.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	accessToken := loginResponse.AccessToken.Token

	t.Run("MFA Disable - Success", func(t *testing.T) {
		// We need to provide the current password to disable MFA
		disablePayload := map[string]string{
			"password": password,
		}
		payloadBytes, _ := json.Marshal(disablePayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/disable", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message": "MFA disabled successfully"}`, w.Body.String())

		// Verify in DB
		var settings models.MfaSettings
		db.Where("user_id = ?", user.ID).First(&settings)
		assert.False(t, settings.MfaEnabled)
		assert.Nil(t, settings.TotpSecret)
	})

	t.Run("MFA Disable - Invalid Password", func(t *testing.T) {
		// Re-enable MFA for this test
		secret := "secret"
		mfaSettings.MfaEnabled = true
		mfaSettings.TotpSecret = &secret
		db.Save(&mfaSettings)

		disablePayload := map[string]string{
			"password": "wrongpassword",
		}
		payloadBytes, _ := json.Marshal(disablePayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/disable", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrInvalidPassword, errResp.Code)
	})
}

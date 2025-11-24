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

func TestAuthMfaVerifySetup(t *testing.T) {
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword := utils.HashPassword(password)

	user := models.User{
		Name:     "Test User MFA Verify Setup",
		Email:    "test_mfa_verify_setup@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	// Login to get access token
	loginPayload := map[string]string{
		"email":    "test_mfa_verify_setup@example.com",
		"password": password,
	}
	payloadBytes, _ := json.Marshal(loginPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResponse services.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	accessToken := loginResponse.AccessToken.Token

	// Setup MFA first to get the secret
	w = httptest.NewRecorder()
	setupPayload := map[string]string{
		"email": "test_mfa_verify_setup@example.com",
	}
	payloadBytes, _ = json.Marshal(setupPayload)
	req, _ = http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var setupResponse struct {
		Secret    string `json:"secret"`
		QrCodeURL string `json:"qr_code_url"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &setupResponse)
	require.NoError(t, err)
	secret := setupResponse.Secret

	t.Run("Verify MFA Setup - Success", func(t *testing.T) {
		code, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)

		verifyPayload := map[string]string{
			"code": code,
		}
		payloadBytes, _ := json.Marshal(verifyPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/verify-setup", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "MFA setup verified successfully", response["message"])
		assert.NotNil(t, response["backup_codes"])
		assert.Len(t, response["backup_codes"], 10)

		// Verify in DB
		var mfaSettings models.MfaSettings
		db.Where("user_id = ?", user.ID).First(&mfaSettings)
		assert.True(t, mfaSettings.MfaEnabled)
	})

	t.Run("Verify MFA Setup - Invalid Code", func(t *testing.T) {
		verifyPayload := map[string]string{
			"code": "000000",
		}
		payloadBytes, _ := json.Marshal(verifyPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/mfa/verify-setup", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrMfaInvalidCode, errResp.Code)
	})
}

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAuthMfaSetup(t *testing.T) {
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword := utils.HashPassword(password)

	user := models.User{
		Name:     "Test User MFA Setup",
		Email:    "test_mfa_setup@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	// Login to get access token
	loginPayload := map[string]string{
		"email":    "test_mfa_setup@example.com",
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

	t.Run("MFA Setup - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		setupPayload := map[string]string{
			"email": "test_mfa_setup@example.com",
		}
		payloadBytes, _ := json.Marshal(setupPayload)
		req, _ := http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Secret    string `json:"secret"`
			QrCodeURL string `json:"qr_code"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Secret)
		assert.NotEmpty(t, response.QrCodeURL)
	})

	t.Run("MFA Setup - Already Enabled", func(t *testing.T) {
		// Enable MFA for the user manually
		// Update existing MFA settings to enabled
		db.Model(&models.MfaSettings{}).Where("user_id = ?", user.ID).Updates(map[string]interface{}{
			"mfa_enabled": true,
			"totp_secret": "secret",
		})

		w := httptest.NewRecorder()
		setupPayload := map[string]string{
			"email": "test_mfa_setup@example.com",
		}
		payloadBytes, _ := json.Marshal(setupPayload)
		req, _ := http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrMfaAlreadyEnabled, errResp.Code)

		// Clean up
		db.Where("user_id = ?", user.ID).Delete(&models.MfaSettings{})
	})

	t.Run("MFA Setup - Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		setupPayload := map[string]string{
			"email": "test_mfa_setup@example.com",
		}
		payloadBytes, _ := json.Marshal(setupPayload)
		req, _ := http.NewRequest("POST", "/api/v1/mfa/setup", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

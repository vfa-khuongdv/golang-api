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
)

func TestAuthMfaStatus(t *testing.T) {
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword := utils.HashPassword(password)

	user := models.User{
		Name:     "Test User MFA Status",
		Email:    "test_mfa_status@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	// Login to get access token
	loginPayload := map[string]string{
		"email":    "test_mfa_status@example.com",
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

	t.Run("MFA Status - Disabled", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/mfa/status", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			MfaEnabled bool `json:"mfa_enabled"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.MfaEnabled)
	})

	t.Run("MFA Status - Enabled", func(t *testing.T) {
		// Enable MFA
		mfaSettings := models.MfaSettings{
			UserID:     user.ID,
			MfaEnabled: true,
		}
		db.Create(&mfaSettings)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/mfa/status", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			MfaEnabled bool `json:"mfa_enabled"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.MfaEnabled)
	})
}

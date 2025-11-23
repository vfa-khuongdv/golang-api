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

func TestAuthRefreshToken(t *testing.T) {
	router, db := setupTestRouter()

	// Helper to create a user directly in DB
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User Refresh",
		Email:    "test_refresh@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	// Create MFA settings for the user
	mfaSettings := models.MfaSettings{
		UserID:     user.ID,
		MfaEnabled: false,
	}
	db.Create(&mfaSettings)

	// Login to get tokens
	loginPayload := map[string]string{
		"email":    "test_refresh@example.com",
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
	refreshToken := loginResponse.RefreshToken.Token

	t.Run("Refresh Token - Success", func(t *testing.T) {
		refreshPayload := map[string]string{
			"refresh_token": refreshToken,
			"access_token":  accessToken,
		}
		payloadBytes, _ := json.Marshal(refreshPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/refresh-token", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response services.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken.Token)
	})

	t.Run("Refresh Token - Invalid Token", func(t *testing.T) {
		refreshPayload := map[string]string{
			"refresh_token": "invalid_token",
			"access_token":  accessToken,
		}
		payloadBytes, _ := json.Marshal(refreshPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/refresh-token", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 3000, errResp.Code) // ErrUnauthorized
	})

	t.Run("Refresh Token - Missing Token", func(t *testing.T) {
		refreshPayload := map[string]string{
			"access_token": accessToken,
		}
		payloadBytes, _ := json.Marshal(refreshPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/refresh-token", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 4001, errResp.Code) // ErrValidationFailed
	})
}

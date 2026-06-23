package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAuthRefreshToken(t *testing.T) {
	router, db := setupTestRouter()

	// Helper to create a user directly in DB
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User Refresh",
		Email:    "test_refresh@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

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

	var loginResponse dto.LoginResponse
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

		var response dto.LoginResponse
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
		assert.Equal(t, apperror.ErrUnauthorized, errResp.Code)
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
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Refresh Token - Missing AccessToken", func(t *testing.T) {
		refreshPayload := map[string]string{
			"refresh_token": refreshToken,
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
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Refresh Token - Both Tokens Missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/refresh-token", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Refresh Token - Expired Refresh Token", func(t *testing.T) {
		expiredRefresh := models.RefreshToken{
			RefreshToken: "expired-refresh-token",
			IpAddress:    "127.0.0.1",
			ExpiredAt:    time.Now().Add(-1 * time.Hour).Unix(),
			UserID:       user.ID,
		}
		db.Create(&expiredRefresh)

		refreshPayload := map[string]string{
			"refresh_token": "expired-refresh-token",
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
		assert.Equal(t, apperror.ErrUnauthorized, errResp.Code)
	})

	t.Run("Refresh Token - Token Mismatch", func(t *testing.T) {
		otherUser := models.User{
			Name:     "Other User",
			Email:    "other@example.com",
			Password: hashedPassword,
			Gender:   1,
		}
		db.Create(&otherUser)

		otherLoginPayload := map[string]string{
			"email":    "other@example.com",
			"password": password,
		}
		otherPayloadBytes, _ := json.Marshal(otherLoginPayload)
		otherW := httptest.NewRecorder()
		otherReq, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(otherPayloadBytes))
		otherReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(otherW, otherReq)
		require.Equal(t, http.StatusOK, otherW.Code)

		var otherLogin dto.LoginResponse
		err := json.Unmarshal(otherW.Body.Bytes(), &otherLogin)
		require.NoError(t, err)

		refreshPayload := map[string]string{
			"refresh_token": otherLogin.RefreshToken.Token,
			"access_token":  accessToken,
		}
		payloadBytes, _ := json.Marshal(refreshPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/refresh-token", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errResp ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrUnauthorized, errResp.Code)
	})
}

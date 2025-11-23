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

func TestAuthLogin(t *testing.T) {
	router, db := setupTestRouter()

	// Helper to create a user directly in DB
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User",
		Email:    "test_login@example.com",
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

	t.Run("Login - Success", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test_login@example.com",
			"password": password,
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response services.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken.Token)
		assert.NotEmpty(t, response.RefreshToken.Token)
	})

	t.Run("Login - Invalid Credentials", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test_login@example.com",
			"password": "wrongpassword",
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 3003, errResp.Code) // ErrInvalidPassword
	})

	t.Run("Login - Missing Fields", func(t *testing.T) {
		loginPayload := map[string]string{
			"email": "test_login@example.com",
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 4001, errResp.Code) // ErrValidationFailed
	})
}

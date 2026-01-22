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

func TestUsersChangePassword(t *testing.T) {
	router, db := setupTestRouter()

	// Create test user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	testUser := models.User{
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	db.Create(&testUser)

	// Generate access token for test user
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(testUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Change Password - Success", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     password,
			"new_password":     "newpassword123",
			"confirm_password": "newpassword123",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Change password successfully", response["message"])

		// Verify password was changed
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		bcryptService := services.NewBcryptService()
		assert.True(t, bcryptService.CheckPasswordHash("newpassword123", updatedUser.Password))
	})

	t.Run("Change Password - Incorrect Old Password", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     "wrongpassword",
			"new_password":     "newpassword456",
			"confirm_password": "newpassword456",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrInvalidPassword, errResp.Code)
	})

	t.Run("Change Password - New Password Same as Old", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     "newpassword123",
			"new_password":     "newpassword123",
			"confirm_password": "newpassword123",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrPasswordUnchanged, errResp.Code)
	})

	t.Run("Change Password - Password Mismatch", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     "newpassword123",
			"new_password":     "newpassword456",
			"confirm_password": "differentpassword",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrPasswordMismatch, errResp.Code)
	})

	t.Run("Change Password - Missing Fields", func(t *testing.T) {
		payload := map[string]string{
			"old_password": "newpassword123",
			"new_password": "newpassword789",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Change Password - Password Too Short", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     "newpassword123",
			"new_password":     "12345",
			"confirm_password": "12345",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Change Password - Unauthorized without Token", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     "newpassword123",
			"new_password":     "anotherpassword",
			"confirm_password": "anotherpassword",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

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
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestUsersChangePassword(t *testing.T) {
	router, db := setupTestRouter()

	jwtService, err := services.NewJWTService()
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Each subtest gets its own user seeded with the same password, so no
	// subtest depends on another mutating the shared password first.
	createUser := func(t *testing.T, email string) (models.User, string) {
		t.Helper()
		hashedPassword, _ := utils.HashPassword("Password@123")
		user := models.User{
			Name:     "Test User",
			Email:    email,
			Password: hashedPassword,
			Gender:   1,
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		tokenResult, err := jwtService.GenerateAccessToken(user.ID)
		require.NoError(t, err)
		return user, tokenResult.Token
	}

	t.Run("Change Password - Success", func(t *testing.T) {
		testUser, accessToken := createUser(t, "success@example.com")

		payload := map[string]string{
			"old_password":     "Password@123",
			"new_password":     "Newpassword@123",
			"confirm_password": "Newpassword@123",
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
		assert.True(t, utils.CheckPasswordHash("Newpassword@123", updatedUser.Password))
	})

	t.Run("Change Password - Incorrect Old Password", func(t *testing.T) {
		_, accessToken := createUser(t, "wrongold@example.com")

		payload := map[string]string{
			"old_password":     "wrongpassword",
			"new_password":     "Newpassword@456",
			"confirm_password": "Newpassword@456",
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
		_, accessToken := createUser(t, "sameold@example.com")

		payload := map[string]string{
			"old_password":     "Password@123",
			"new_password":     "Password@123",
			"confirm_password": "Password@123",
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
		_, accessToken := createUser(t, "mismatch@example.com")

		payload := map[string]string{
			"old_password":     "Password@123",
			"new_password":     "Newpassword@456",
			"confirm_password": "Different@456",
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
		_, accessToken := createUser(t, "missingfields@example.com")

		payload := map[string]string{
			"old_password": "Password@123",
			"new_password": "Newpassword@789",
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
		_, accessToken := createUser(t, "tooshort@example.com")

		payload := map[string]string{
			"old_password":     "Password@123",
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
			"old_password":     "password123",
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

	t.Run("Change Password - Expired Token", func(t *testing.T) {
		payload := map[string]string{
			"old_password":     "password123",
			"new_password":     "anotherpassword",
			"confirm_password": "anotherpassword",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/change-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateExpiredToken(1))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

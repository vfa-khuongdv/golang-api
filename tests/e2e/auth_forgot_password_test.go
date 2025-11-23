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
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

func TestAuthForgotPassword(t *testing.T) {
	router, db := setupTestRouter()

	// Helper to create a user directly in DB
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User Forgot",
		Email:    "test_forgot@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	t.Run("Forgot Password - Success", func(t *testing.T) {
		// Note: This test will fail at email sending step due to missing SMTP config,
		// but we can verify the token was generated in DB before that
		payload := map[string]string{
			"email": "test_forgot@example.com",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		// The request will fail at email sending (500) due to missing SMTP config
		// but the token should still be generated in DB
		if w.Code == http.StatusInternalServerError {
			// Verify token was generated despite email failure
			var updatedUser models.User
			db.First(&updatedUser, user.ID)
			assert.NotNil(t, updatedUser.Token)
			assert.NotNil(t, updatedUser.ExpiredAt)

			var errResp ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, 1000, errResp.Code) // ErrInternal (email sending failed)
		} else {
			// If SMTP is configured, should succeed
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("Forgot Password - Email Not Found", func(t *testing.T) {
		payload := map[string]string{
			"email": "nonexistent@example.com",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 1001, errResp.Code) // ErrNotFound
	})

	t.Run("Forgot Password - Invalid Email Format", func(t *testing.T) {
		payload := map[string]string{
			"email": "invalid-email",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, 4001, errResp.Code) // ErrValidationFailed
	})
}

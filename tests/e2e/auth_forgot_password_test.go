package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAuthForgotPassword(t *testing.T) {
	router, db := setupTestRouter()

	// Helper to create a user directly in DB
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	user := models.User{
		Name:     "Test User Forgot",
		Email:    "test_forgot@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	t.Run("Forgot Password - Token Persisted On Email Failure", func(t *testing.T) {
		// Force deterministic behavior: with MAIL_FROM unset the sender fails
		// fast on address parsing (no network), so we always hit the 500 path
		// and can verify the token was persisted before the send.
		for _, key := range []string{"MAIL_HOST", "MAIL_PORT", "MAIL_USERNAME", "MAIL_PASSWORD", "MAIL_FROM"} {
			prev, had := os.LookupEnv(key)
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("failed to unset %s: %v", key, err)
			}
			t.Cleanup(func() {
				if had {
					_ = os.Setenv(key, prev)
				}
			})
		}

		payload := map[string]string{
			"email": "test_forgot@example.com",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		// Email sending fails (500) but the token should still be generated in DB
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.NotNil(t, updatedUser.ResetToken)
		assert.NotNil(t, updatedUser.ResetExpiredAt)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrInternalServer, errResp.Code)
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

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "If your email is in our system, you will receive instructions to reset your password", resp["message"])
	})

	t.Run("Forgot Password - Empty Email", func(t *testing.T) {
		payload := map[string]string{
			"email": "",
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
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
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
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})
}

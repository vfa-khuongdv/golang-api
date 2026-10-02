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

	t.Run("Forgot Password - Email Failure Answers Like An Unknown Email", func(t *testing.T) {
		// With mail.from empty the sender fails fast on address parsing (no
		// network). The answer must still be the same 200 as for an unknown
		// email, otherwise the endpoint reveals which emails are registered.
		require.NoError(t, db.Create(&[]models.Setting{
			{Key: "mail.host", Value: "smtp.example.com"},
			{Key: "mail.port", Value: "587"},
			{Key: "mail.username", Value: ""},
			{Key: "mail.password", Value: ""},
			{Key: "mail.from", Value: ""},
			{Key: "app.frontend_url", Value: "http://localhost:5173"},
		}).Error)

		payload := map[string]string{
			"email": "test_forgot@example.com",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		unknown := httptest.NewRecorder()
		unknownReq, _ := http.NewRequest("POST", "/api/v1/forgot-password", bytes.NewBufferString(`{"email":"nobody@example.com"}`))
		unknownReq.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(unknown, unknownReq)
		assert.Equal(t, unknown.Body.String(), w.Body.String())

		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.NotNil(t, updatedUser.ResetToken)
		assert.NotNil(t, updatedUser.ResetExpiredAt)
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

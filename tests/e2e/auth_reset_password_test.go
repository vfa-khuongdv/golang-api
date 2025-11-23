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
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestAuthResetPassword(t *testing.T) {
	router, db := setupTestRouter()

	password := "password123"
	hashedPassword := utils.HashPassword(password)
	token := "valid_reset_token"
	expiredAt := time.Now().Add(time.Hour).Unix()

	user := models.User{
		Name:      "Test User Reset",
		Email:     "test_reset@example.com",
		Password:  hashedPassword,
		Gender:    1,
		Token:     &token,
		ExpiredAt: &expiredAt,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)

	t.Run("Reset Password - Success", func(t *testing.T) {
		newPassword := "newpassword123"
		payload := map[string]string{
			"token":        token,
			"password":     password, // Current implementation requires old password
			"new_password": newPassword,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Reset password successfully")

		// Verify password was changed
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.True(t, utils.CheckPasswordHash(newPassword, updatedUser.Password))
		assert.Nil(t, updatedUser.Token)
		assert.Nil(t, updatedUser.ExpiredAt)
	})

	t.Run("Reset Password - Invalid Token", func(t *testing.T) {
		payload := map[string]string{
			"token":        "invalid_token",
			"password":     password,
			"new_password": "newpassword123",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusOK, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrNotFound, errResp.Code)
	})

	t.Run("Reset Password - Expired Token", func(t *testing.T) {
		expiredToken := "expired_token"
		expiredTime := time.Now().Add(-time.Hour).Unix()

		expiredUser := models.User{
			Name:      "Expired User",
			Email:     "expired@example.com",
			Password:  hashedPassword,
			Gender:    1,
			Token:     &expiredToken,
			ExpiredAt: &expiredTime,
		}
		db.Create(&expiredUser)

		payload := map[string]string{
			"token":        expiredToken,
			"password":     password,
			"new_password": "newpassword123",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/reset-password", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrTokenExpired, errResp.Code)
	})
}

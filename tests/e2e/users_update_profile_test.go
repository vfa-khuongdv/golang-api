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
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestUsersUpdateProfile(t *testing.T) {
	router, db := setupTestRouter()

	// Create test user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	birthday := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	address := "123 Original Street"
	testUser := models.User{
		Name:     "Original Name",
		Email:    "testuser@example.com",
		Password: hashedPassword,
		Birthday: &birthday,
		Address:  &address,
		Gender:   1,
	}
	db.Create(&testUser)

	// Generate access token for test user
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(testUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Update Profile - Name Only", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Updated Profile Name",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Update profile successfully", response["message"])

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, "Updated Profile Name", updatedUser.Name)
	})

	t.Run("Update Profile - Birthday Only", func(t *testing.T) {
		newBirthday := "1995-08-20"
		payload := map[string]interface{}{
			"birthday": newBirthday,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, newBirthday, updatedUser.Birthday.Format("2006-01-02"))
	})

	t.Run("Update Profile - Address Only", func(t *testing.T) {
		newAddress := "456 New Profile Street"
		payload := map[string]interface{}{
			"address": newAddress,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, newAddress, *updatedUser.Address)
	})

	t.Run("Update Profile - Gender Only", func(t *testing.T) {
		payload := map[string]interface{}{
			"gender": 2,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, int16(2), updatedUser.Gender)
	})

	t.Run("Update Profile - Multiple Fields", func(t *testing.T) {
		newBirthday := "2000-01-01"
		newAddress := "789 Multi Update Avenue"
		payload := map[string]interface{}{
			"name":     "Multi Update Name",
			"birthday": newBirthday,
			"address":  newAddress,
			"gender":   3,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify all updates in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, "Multi Update Name", updatedUser.Name)
		assert.Equal(t, newAddress, *updatedUser.Address)
		assert.Equal(t, int16(3), updatedUser.Gender)
	})

	t.Run("Update Profile - Invalid Gender", func(t *testing.T) {
		payload := map[string]interface{}{
			"gender": 5, // Invalid: must be 1, 2, or 3
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Update Profile - Invalid Birthday Format", func(t *testing.T) {
		payload := map[string]interface{}{
			"birthday": "invalid-date",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Update Profile - Blank Name", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "   ", // Blank name
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Update Profile - Unauthorized without Token", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Unauthorized Update",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/profile", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

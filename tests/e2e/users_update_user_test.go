package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

func TestUsersUpdateUser(t *testing.T) {
	router, db := setupTestRouter()

	// Create test user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	birthday := "1990-01-15"
	address := "Original Address"
	testUser := models.User{
		Name:     "Original Name",
		Email:    "testuser@example.com",
		Password: hashedPassword,
		Birthday: &birthday,
		Address:  &address,
		Gender:   1,
	}
	db.Create(&testUser)

	// Create an authenticated user and generate token
	authUser := models.User{
		Name:     "Auth User",
		Email:    "auth@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	db.Create(&authUser)

	// Generate access token
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(authUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Update User - Name Only", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Updated Name",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Update user successfully", response["message"])

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, "Updated Name", updatedUser.Name)
	})

	t.Run("Update User - Birthday Only", func(t *testing.T) {
		newBirthday := "1995-05-20"
		payload := map[string]interface{}{
			"birthday": newBirthday,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		// Birthday may be returned as ISO timestamp, so check if it starts with the expected date
		assert.True(t, strings.HasPrefix(*updatedUser.Birthday, newBirthday), "Expected birthday to start with %s, got %s", newBirthday, *updatedUser.Birthday)
	})

	t.Run("Update User - Address Only", func(t *testing.T) {
		newAddress := "New Address 456"
		payload := map[string]interface{}{
			"address": newAddress,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, newAddress, *updatedUser.Address)
	})

	t.Run("Update User - Gender Only", func(t *testing.T) {
		payload := map[string]interface{}{
			"gender": 2,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, int16(2), updatedUser.Gender)
	})

	t.Run("Update User - Multiple Fields", func(t *testing.T) {
		newBirthday := "2000-12-25"
		newAddress := "Multi Update Address"
		payload := map[string]interface{}{
			"name":     "Multi Update Name",
			"birthday": newBirthday,
			"address":  newAddress,
			"gender":   3,
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify all updates in database
		var updatedUser models.User
		db.First(&updatedUser, testUser.ID)
		assert.Equal(t, "Multi Update Name", updatedUser.Name)
		// Birthday may be returned as ISO timestamp, so check if it starts with the expected date
		assert.True(t, strings.HasPrefix(*updatedUser.Birthday, newBirthday), "Expected birthday to start with %s, got %s", newBirthday, *updatedUser.Birthday)
		assert.Equal(t, newAddress, *updatedUser.Address)
		assert.Equal(t, int16(3), updatedUser.Gender)
	})

	t.Run("Update User - Not Found", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Non Existent",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/users/99999", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrNotFound, errResp.Code)
	})

	t.Run("Update User - Invalid ID Format", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Invalid ID",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/users/invalid", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrParseError, errResp.Code)
	})

	t.Run("Update User - Invalid Gender", func(t *testing.T) {
		payload := map[string]interface{}{
			"gender": 5, // Invalid: must be 1, 2, or 3
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Update User - Blank Name", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "   ", // Blank name
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Update User - Unauthorized without Token", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Unauthorized Update",
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/%d", testUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

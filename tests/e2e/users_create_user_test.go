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

func TestUsersCreateUser(t *testing.T) {
	router, db := setupTestRouter()

	// Create an authenticated user and generate token
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	authUser := models.User{
		Name:     "Auth User",
		Email:    "auth@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	db.Create(&authUser)

	// Create MFA settings
	mfaSettings := models.MfaSettings{
		UserID:     authUser.ID,
		MfaEnabled: false,
	}
	db.Create(&mfaSettings)

	// Generate access token
	jwtService := services.NewJWTService()
	tokenResult, err := jwtService.GenerateAccessToken(authUser.ID)
	require.NoError(t, err)
	accessToken := tokenResult.Token

	t.Run("Create User - Success", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "newuser@example.com",
			"password": "password123",
			"name":     "New User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Create user successfully", response["message"])

		// Verify user was created in database
		var createdUser models.User
		result := db.Where("email = ?", "newuser@example.com").First(&createdUser)
		assert.NoError(t, result.Error)
		assert.Equal(t, "New User", createdUser.Name)
	})

	t.Run("Create User - Missing Email", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"password": "password123",
			"name":     "New User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Create User - Invalid Email Format", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "invalid-email",
			"password": "password123",
			"name":     "New User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Create User - Password Too Short", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "short@example.com",
			"password": "12345",
			"name":     "New User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Create User - Invalid Gender", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "gender@example.com",
			"password": "password123",
			"name":     "New User",
			"birthday": birthday,
			"address":  address,
			"gender":   5, // Invalid: must be 1, 2, or 3
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Create User - Duplicate Email", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "auth@example.com", // Already exists
			"password": "password123",
			"name":     "Duplicate User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		// Database constraint violations return 500, not 400
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Create User - Empty Role IDs", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "norole@example.com",
			"password": "password123",
			"name":     "No Role User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrValidationFailed, errResp.Code)
	})

	t.Run("Create User - Unauthorized without Token", func(t *testing.T) {
		birthday := "1990-01-15"
		address := "123 Main Street"
		payload := map[string]interface{}{
			"email":    "unauth@example.com",
			"password": "password123",
			"name":     "Unauth User",
			"birthday": birthday,
			"address":  address,
			"gender":   1,
			"role_ids": []uint{1},
		}
		payloadBytes, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

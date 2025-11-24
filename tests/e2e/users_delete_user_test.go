package e2e

import (
	"encoding/json"
	"fmt"
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

func TestUsersDeleteUser(t *testing.T) {
	router, db := setupTestRouter()

	// Create test user
	password := "password123"
	hashedPassword := utils.HashPassword(password)
	testUser := models.User{
		Name:     "Test User To Delete",
		Email:    "delete@example.com",
		Password: hashedPassword,
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

	t.Run("Delete User - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", testUser.ID), nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify user was soft deleted
		var deletedUser models.User
		result := db.Unscoped().First(&deletedUser, testUser.ID)
		assert.NoError(t, result.Error)
		assert.NotNil(t, deletedUser.DeletedAt)
	})

	t.Run("Delete User - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/users/99999", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrNotFound, errResp.Code)
	})

	t.Run("Delete User - Invalid ID Format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/users/invalid", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, apperror.ErrParseError, errResp.Code)
	})

	t.Run("Delete User - Unauthorized without Token", func(t *testing.T) {
		// Create another user to delete
		anotherUser := models.User{
			Name:     "Another User",
			Email:    "another@example.com",
			Password: hashedPassword,
			Gender:   1,
		}
		db.Create(&anotherUser)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", anotherUser.ID), nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

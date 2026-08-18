package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

// Regression: /logout is a body-less POST. Empty-body rejection is handled
// centrally in TranslateValidationErrors, so a plain POST /logout (which
// binds no body) must succeed instead of being rejected with 400.
func TestLogout(t *testing.T) {
	router, db := setupTestRouter()

	jwtService, err := services.NewJWTService()
	require.NoError(t, err)

	hashedPassword, _ := utils.HashPassword("Password@123")
	user := models.User{
		Name:     "Logout User",
		Email:    "logout@example.com",
		Password: hashedPassword,
		Gender:   1,
	}
	require.NoError(t, db.Create(&user).Error)

	tokenResult, err := jwtService.GenerateAccessToken(user.ID)
	require.NoError(t, err)

	t.Run("Logout with valid token and no body", func(t *testing.T) {
		w := httptest.NewRecorder()
		// No body on purpose — this must not be rejected as an empty request body.
		req, _ := http.NewRequest("POST", "/api/v1/logout", nil)
		req.Header.Set("Authorization", "Bearer "+tokenResult.Token)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Logged out successfully")
	})

	t.Run("Logout without token is unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/logout", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

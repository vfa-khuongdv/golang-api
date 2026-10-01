package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"gorm.io/gorm"
)

func createAdminWithToken(t *testing.T, db *gorm.DB, email string) string {
	t.Helper()
	_, token := createUserWithPermissions(t, db, email, constants.AllPermissions()...)
	return token
}

func settingsRequest(router http.Handler, method, token, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, "/api/v1/settings", bytes.NewBufferString(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	return w
}

func TestSettings(t *testing.T) {
	router, db := setupTestRouter()
	token := createAdminWithToken(t, db, "settings-user@example.com")

	// The in-memory database is shared across tests, so start from a clean table.
	require.NoError(t, db.Where("1 = 1").Delete(&models.Setting{}).Error)
	require.NoError(t, db.Create(&[]models.Setting{
		{Key: "mail.host", Value: "127.0.0.1"},
		{Key: "mail.port", Value: "1026"},
		{Key: "mail.username", Value: ""},
		{Key: "mail.password", Value: ""},
		{Key: "mail.from", Value: "noreply@example.com"},
		{Key: "app.frontend_url", Value: "http://localhost:5173"},
	}).Error)

	t.Run("Unauthorized Without Token", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, settingsRequest(router, "GET", "", "").Code)
		assert.Equal(t, http.StatusUnauthorized, settingsRequest(router, "PUT", "", `{"mail_host":"x"}`).Code)
	})

	t.Run("Forbidden Without Permission", func(t *testing.T) {
		_, noPermToken := createUserWithPermissions(t, db, "settings-noperm@example.com")
		_, readOnlyToken := createUserWithPermissions(t, db, "settings-readonly@example.com", constants.PermissionSettingsRead)

		assert.Equal(t, http.StatusForbidden, settingsRequest(router, "GET", noPermToken, "").Code)
		assert.Equal(t, http.StatusOK, settingsRequest(router, "GET", readOnlyToken, "").Code)
		assert.Equal(t, http.StatusForbidden, settingsRequest(router, "PUT", readOnlyToken, `{"mail_host":"x"}`).Code)
	})

	t.Run("Gets Settings", func(t *testing.T) {
		w := settingsRequest(router, "GET", token, "")

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"mail_host":"127.0.0.1","mail_port":1026,"mail_username":"","mail_password_set":false,"mail_from":"noreply@example.com","frontend_url":"http://localhost:5173"}`, w.Body.String())
	})

	t.Run("Updates Settings And Password Is Encrypted", func(t *testing.T) {
		w := settingsRequest(router, "PUT", token, `{"mail_host":"smtp.example.com","mail_port":587,"mail_password":"s3cret","frontend_url":"https://app.example.com/"}`)
		require.Equal(t, http.StatusOK, w.Code)

		w = settingsRequest(router, "GET", token, "")
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "s3cret")
		var got map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "smtp.example.com", got["mail_host"])
		assert.EqualValues(t, 587, got["mail_port"])
		assert.Equal(t, true, got["mail_password_set"])
		assert.Equal(t, "https://app.example.com", got["frontend_url"])
		assert.Equal(t, "noreply@example.com", got["mail_from"], "fields not sent stay unchanged")

		var stored models.Setting
		require.NoError(t, db.Where("`key` = ?", "mail.password").First(&stored).Error)
		assert.True(t, strings.HasPrefix(stored.Value, utils.EncryptedPrefix))
	})

	t.Run("Invalid Input Is Rejected", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, settingsRequest(router, "PUT", token, `{"mail_port":0}`).Code)
		assert.Equal(t, http.StatusBadRequest, settingsRequest(router, "PUT", token, `{"frontend_url":"javascript:alert(1)"}`).Code)
		assert.Equal(t, http.StatusBadRequest, settingsRequest(router, "PUT", token, "").Code)
	})
}

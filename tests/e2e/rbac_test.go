package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"gorm.io/gorm"
)

// createUserWithPermissions creates a user holding a dedicated role that grants
// exactly the given permissions, and returns the user with an access token.
func createUserWithPermissions(t *testing.T, db *gorm.DB, email string, permissions ...string) (models.User, string) {
	t.Helper()
	hashed, err := utils.HashPassword("password123")
	require.NoError(t, err)
	user := models.User{Name: "Test", Email: email, Password: hashed, Gender: 1}
	require.NoError(t, db.Create(&user).Error)

	role := models.Role{Name: "test-role-" + email}
	for _, name := range permissions {
		perm := models.Permission{Name: name}
		require.NoError(t, db.Where("name = ?", name).FirstOrCreate(&perm).Error)
		role.Permissions = append(role.Permissions, perm)
	}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&models.UserRole{UserID: user.ID, RoleID: role.ID}).Error)

	jwtService, err := services.NewJWTService()
	require.NoError(t, err)
	token, err := jwtService.GenerateAccessToken(user.ID)
	require.NoError(t, err)
	return user, token.Token
}

func rbacRequest(router http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(w, req)
	return w
}

func TestRBAC(t *testing.T) {
	router, db := setupTestRouter()
	_, adminToken := createUserWithPermissions(t, db, "rbac-admin@example.com", constants.AllPermissions()...)
	_, plainToken := createUserWithPermissions(t, db, "rbac-plain@example.com")

	var readPerm, updatePerm models.Permission
	require.NoError(t, db.Where("name = ?", constants.PermissionRolesRead).First(&readPerm).Error)
	require.NoError(t, db.Where("name = ?", constants.PermissionRolesUpdate).First(&updatePerm).Error)

	t.Run("Every Endpoint Requires A Token", func(t *testing.T) {
		for _, route := range [][2]string{
			{"GET", "/api/v1/roles"}, {"POST", "/api/v1/roles"}, {"GET", "/api/v1/roles/1"},
			{"PUT", "/api/v1/roles/1"}, {"DELETE", "/api/v1/roles/1"},
			{"GET", "/api/v1/permissions"}, {"PUT", "/api/v1/users/1/roles"},
		} {
			assert.Equal(t, http.StatusUnauthorized, rbacRequest(router, route[0], route[1], "", "{}").Code, route)
		}
	})

	t.Run("Every Endpoint Is Forbidden Without Permission", func(t *testing.T) {
		for _, route := range [][2]string{
			{"GET", "/api/v1/roles"}, {"POST", "/api/v1/roles"}, {"GET", "/api/v1/roles/1"},
			{"PUT", "/api/v1/roles/1"}, {"DELETE", "/api/v1/roles/1"},
			{"GET", "/api/v1/permissions"}, {"PUT", "/api/v1/users/1/roles"},
		} {
			assert.Equal(t, http.StatusForbidden, rbacRequest(router, route[0], route[1], plainToken, `{"name":"x","role_ids":[]}`).Code, route)
		}
	})

	t.Run("Role Lifecycle", func(t *testing.T) {
		body := fmt.Sprintf(`{"name":"editor","description":"edits","permission_ids":[%d]}`, readPerm.ID)
		w := rbacRequest(router, "POST", "/api/v1/roles", adminToken, body)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var created models.Role
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
		require.Len(t, created.Permissions, 1)

		w = rbacRequest(router, "POST", "/api/v1/roles", adminToken, body)
		assert.Equal(t, http.StatusConflict, w.Code, "duplicate name")

		w = rbacRequest(router, "PUT", fmt.Sprintf("/api/v1/roles/%d", created.ID), adminToken,
			fmt.Sprintf(`{"name":"writer","permission_ids":[%d]}`, updatePerm.ID))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		w = rbacRequest(router, "GET", fmt.Sprintf("/api/v1/roles/%d", created.ID), adminToken, "")
		require.Equal(t, http.StatusOK, w.Code)
		var got models.Role
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "writer", got.Name)
		require.Len(t, got.Permissions, 1)
		assert.Equal(t, constants.PermissionRolesUpdate, got.Permissions[0].Name)

		w = rbacRequest(router, "DELETE", fmt.Sprintf("/api/v1/roles/%d", created.ID), adminToken, "")
		assert.Equal(t, http.StatusOK, w.Code)
		w = rbacRequest(router, "GET", fmt.Sprintf("/api/v1/roles/%d", created.ID), adminToken, "")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Unknown Permission Is Rejected", func(t *testing.T) {
		w := rbacRequest(router, "POST", "/api/v1/roles", adminToken, `{"name":"ghost","permission_ids":[99999]}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Admin Role Is Protected", func(t *testing.T) {
		admin := models.Role{Name: models.RoleAdmin}
		require.NoError(t, db.Where("name = ?", models.RoleAdmin).FirstOrCreate(&admin).Error)

		w := rbacRequest(router, "PUT", fmt.Sprintf("/api/v1/roles/%d", admin.ID), adminToken, `{"name":"renamed"}`)
		assert.Equal(t, http.StatusForbidden, w.Code)
		w = rbacRequest(router, "DELETE", fmt.Sprintf("/api/v1/roles/%d", admin.ID), adminToken, "")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Lists Permissions", func(t *testing.T) {
		w := rbacRequest(router, "GET", "/api/v1/permissions", adminToken, "")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), constants.PermissionRolesRead)
	})

	t.Run("Assigning Roles Takes Effect Immediately", func(t *testing.T) {
		target, targetToken := createUserWithPermissions(t, db, "rbac-target@example.com")
		require.Equal(t, http.StatusForbidden, rbacRequest(router, "GET", "/api/v1/roles", targetToken, "").Code)

		var role models.Role
		w := rbacRequest(router, "POST", "/api/v1/roles", adminToken,
			fmt.Sprintf(`{"name":"reader","permission_ids":[%d]}`, readPerm.ID))
		require.Equal(t, http.StatusCreated, w.Code)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &role))

		w = rbacRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/roles", target.ID), adminToken,
			fmt.Sprintf(`{"role_ids":[%d]}`, role.ID))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, http.StatusOK, rbacRequest(router, "GET", "/api/v1/roles", targetToken, "").Code)

		w = rbacRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/roles", target.ID), adminToken, `{"role_ids":[]}`)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, http.StatusForbidden, rbacRequest(router, "GET", "/api/v1/roles", targetToken, "").Code)
	})

	t.Run("Assigning To Unknown User Or Role", func(t *testing.T) {
		w := rbacRequest(router, "PUT", "/api/v1/users/99999/roles", adminToken, `{"role_ids":[]}`)
		assert.Equal(t, http.StatusNotFound, w.Code)
		w = rbacRequest(router, "PUT", "/api/v1/users/1/roles", adminToken, `{"role_ids":[99999]}`)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("The Last Admin Cannot Be Demoted", func(t *testing.T) {
		// Other tests share this database, so make the admin role's holders known.
		var adminRole models.Role
		require.NoError(t, db.Where("name = ?", models.RoleAdmin).FirstOrCreate(&adminRole, models.Role{Name: models.RoleAdmin}).Error)
		require.NoError(t, db.Where("role_id = ?", adminRole.ID).Delete(&models.UserRole{}).Error)
		onlyAdmin, _ := createUserWithPermissions(t, db, "rbac-only-admin@example.com")
		require.NoError(t, db.Create(&models.UserRole{UserID: onlyAdmin.ID, RoleID: adminRole.ID}).Error)
		actor, actorToken := createUserWithPermissions(t, db, "rbac-actor@example.com", constants.PermissionUsersAssign)

		w := rbacRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/roles", onlyAdmin.ID), actorToken, `{"role_ids":[]}`)
		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())

		// With a second admin, the first can be demoted.
		require.NoError(t, db.Create(&models.UserRole{UserID: actor.ID, RoleID: adminRole.ID}).Error)
		w = rbacRequest(router, "PUT", fmt.Sprintf("/api/v1/users/%d/roles", onlyAdmin.ID), actorToken, `{"role_ids":[]}`)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})
}

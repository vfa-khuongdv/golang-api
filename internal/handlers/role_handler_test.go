package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/handlers"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func newRoleContext(method, body, id string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(method, "/api/v1/roles", bytes.NewBufferString(body))
	if id != "" {
		c.Params = gin.Params{{Key: "id", Value: id}}
	}
	c.Set("UserID", testActorID)
	return c, w
}

// testActorID is the authenticated user making the requests.
const testActorID = uint(99)

func TestRoleHandler_ListRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("ListRoles", mock.Anything).Return([]models.Role{{ID: 1, Name: "admin", Permissions: []models.Permission{{ID: 1, Name: "roles:read"}}}}, nil)
		c, w := newRoleContext("GET", "", "")

		handlers.NewRoleHandler(svc).ListRoles(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"admin"`)
		assert.Contains(t, w.Body.String(), `"roles:read"`)
	})

	t.Run("Empty List Is An Array", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("ListRoles", mock.Anything).Return(nil, nil)
		c, w := newRoleContext("GET", "", "")

		handlers.NewRoleHandler(svc).ListRoles(c)

		assert.JSONEq(t, `[]`, w.Body.String())
	})

	t.Run("Service Error", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("ListRoles", mock.Anything).Return(nil, apperror.NewInternalServerError("boom"))
		c, w := newRoleContext("GET", "", "")

		handlers.NewRoleHandler(svc).ListRoles(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestRoleHandler_GetRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("GetRole", mock.Anything, uint(2)).Return(&models.Role{ID: 2, Name: "editor"}, nil)
		c, w := newRoleContext("GET", "", "2")

		handlers.NewRoleHandler(svc).GetRole(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"editor"`)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		for _, id := range []string{"abc", "0", "-1"} {
			svc := new(mocks.MockRoleService)
			c, w := newRoleContext("GET", "", id)

			handlers.NewRoleHandler(svc).GetRole(c)

			assert.Equal(t, http.StatusBadRequest, w.Code, id)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("GetRole", mock.Anything, uint(2)).Return(nil, apperror.NewNotFoundError("Role not found"))
		c, w := newRoleContext("GET", "", "2")

		handlers.NewRoleHandler(svc).GetRole(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRoleHandler_CreateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitValidator()

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("CreateRole", mock.Anything, testActorID, &dto.RoleInput{Name: "editor", Description: "d", PermissionIDs: []uint{1, 2}}).
			Return(&models.Role{ID: 3, Name: "editor"}, nil)
		c, w := newRoleContext("POST", `{"name":"editor","description":"d","permission_ids":[1,2]}`, "")

		handlers.NewRoleHandler(svc).CreateRole(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"id":3`)
		svc.AssertExpectations(t)
	})

	invalid := map[string]string{
		"missing name":     `{"description":"d"}`,
		"blank name":       `{"name":"   "}`,
		"name too long":    `{"name":"` + string(bytes.Repeat([]byte("a"), 101)) + `"}`,
		"zero permission":  `{"name":"x","permission_ids":[0]}`,
		"negative id":      `{"name":"x","permission_ids":[-1]}`,
		"malformed json":   `{"name":`,
		"description long": `{"name":"x","description":"` + string(bytes.Repeat([]byte("a"), 256)) + `"}`,
	}
	for name, body := range invalid {
		t.Run("Invalid - "+name, func(t *testing.T) {
			svc := new(mocks.MockRoleService)
			c, w := newRoleContext("POST", body, "")

			handlers.NewRoleHandler(svc).CreateRole(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			svc.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything)
		})
	}

	t.Run("Service Conflict", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("CreateRole", mock.Anything, testActorID, mock.Anything).Return(nil, apperror.NewConflictError("Role name already exists"))
		c, w := newRoleContext("POST", `{"name":"editor"}`, "")

		handlers.NewRoleHandler(svc).CreateRole(c)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestRoleHandler_RequiresAnAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MockRoleService)
	h := handlers.NewRoleHandler(svc)

	for name, call := range map[string]func(*gin.Context){
		"CreateRole":   h.CreateRole,
		"UpdateRole":   h.UpdateRole,
		"DeleteRole":   h.DeleteRole,
		"SetUserRoles": h.SetUserRoles,
	} {
		t.Run(name, func(t *testing.T) {
			c, w := newRoleContext("POST", `{"name":"editor","role_ids":[1]}`, "3")
			delete(c.Keys, "UserID")

			call(c)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
	svc.AssertNotCalled(t, "CreateRole", mock.Anything, mock.Anything, mock.Anything)
	svc.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	svc.AssertNotCalled(t, "DeleteRole", mock.Anything, mock.Anything, mock.Anything)
	svc.AssertNotCalled(t, "SetUserRoles", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitValidator()

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("UpdateRole", mock.Anything, testActorID, uint(3), &dto.RoleInput{Name: "writer"}).Return(&models.Role{ID: 3, Name: "writer"}, nil)
		c, w := newRoleContext("PUT", `{"name":"writer"}`, "3")

		handlers.NewRoleHandler(svc).UpdateRole(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"writer"`)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		c, w := newRoleContext("PUT", `{"name":"writer"}`, "abc")

		handlers.NewRoleHandler(svc).UpdateRole(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		c, w := newRoleContext("PUT", `{"name":""}`, "3")

		handlers.NewRoleHandler(svc).UpdateRole(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Forbidden For Admin", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("UpdateRole", mock.Anything, testActorID, uint(1), mock.Anything).Return(nil, apperror.NewForbiddenError("no"))
		c, w := newRoleContext("PUT", `{"name":"writer"}`, "1")

		handlers.NewRoleHandler(svc).UpdateRole(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRoleHandler_DeleteRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("DeleteRole", mock.Anything, testActorID, uint(3)).Return(nil)
		c, w := newRoleContext("DELETE", "", "3")

		handlers.NewRoleHandler(svc).DeleteRole(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Delete role successfully"}`, w.Body.String())
	})

	t.Run("Invalid ID", func(t *testing.T) {
		c, w := newRoleContext("DELETE", "", "abc")

		handlers.NewRoleHandler(new(mocks.MockRoleService)).DeleteRole(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("DeleteRole", mock.Anything, testActorID, uint(3)).Return(apperror.NewNotFoundError("Role not found"))
		c, w := newRoleContext("DELETE", "", "3")

		handlers.NewRoleHandler(svc).DeleteRole(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRoleHandler_ListPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("ListPermissions", mock.Anything).Return([]models.Permission{{ID: 1, Name: "roles:read"}}, nil)
		c, w := newRoleContext("GET", "", "")

		handlers.NewRoleHandler(svc).ListPermissions(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"roles:read"`)
	})

	t.Run("Service Error", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("ListPermissions", mock.Anything).Return(nil, apperror.NewInternalServerError("boom"))
		c, w := newRoleContext("GET", "", "")

		handlers.NewRoleHandler(svc).ListPermissions(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestRoleHandler_SetUserRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitValidator()

	t.Run("Success", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("SetUserRoles", mock.Anything, testActorID, uint(4), []uint{1, 2}).Return([]models.Role{{ID: 1, Name: "admin"}}, nil)
		c, w := newRoleContext("PUT", `{"role_ids":[1,2]}`, "4")

		handlers.NewRoleHandler(svc).SetUserRoles(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"admin"`)
	})

	t.Run("Empty List Clears Roles", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("SetUserRoles", mock.Anything, testActorID, uint(4), []uint{}).Return(nil, nil)
		c, w := newRoleContext("PUT", `{"role_ids":[]}`, "4")

		handlers.NewRoleHandler(svc).SetUserRoles(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `[]`, w.Body.String())
	})

	t.Run("Invalid ID", func(t *testing.T) {
		c, w := newRoleContext("PUT", `{"role_ids":[1]}`, "abc")

		handlers.NewRoleHandler(new(mocks.MockRoleService)).SetUserRoles(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		for name, body := range map[string]string{"missing role_ids": `{}`, "zero id": `{"role_ids":[0]}`, "not an array": `{"role_ids":1}`} {
			svc := new(mocks.MockRoleService)
			c, w := newRoleContext("PUT", body, "4")

			handlers.NewRoleHandler(svc).SetUserRoles(c)

			assert.Equal(t, http.StatusBadRequest, w.Code, name)
		}
	})

	t.Run("Service Error", func(t *testing.T) {
		svc := new(mocks.MockRoleService)
		svc.On("SetUserRoles", mock.Anything, testActorID, uint(4), mock.Anything).Return(nil, apperror.NewNotFoundError("User not found"))
		c, w := newRoleContext("PUT", `{"role_ids":[1]}`, "4")

		handlers.NewRoleHandler(svc).SetUserRoles(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

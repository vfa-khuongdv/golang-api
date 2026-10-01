package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

var errRoleNotFound = apperror.New(http.StatusNotFound, apperror.ErrNotFound, "Role not found")

func newRoleService() (services.RoleService, *mocks.MockRoleRepository, *mocks.MockUserRepository) {
	roleRepo := new(mocks.MockRoleRepository)
	userRepo := new(mocks.MockUserRepository)
	return services.NewRoleService(roleRepo, userRepo), roleRepo, userRepo
}

func assertAppError(t *testing.T, err error, status int) {
	t.Helper()
	appErr, ok := err.(*apperror.AppError)
	require.True(t, ok, "expected AppError, got %v", err)
	assert.Equal(t, status, appErr.HttpStatusCode)
}

func TestRoleService_HasPermission(t *testing.T) {
	svc, roleRepo, _ := newRoleService()
	roleRepo.On("HasPermission", mock.Anything, uint(1), "roles:read").Return(true, nil)

	ok, err := svc.HasPermission(context.Background(), 1, "roles:read")

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRoleService_CreateRole(t *testing.T) {
	ctx := context.Background()
	input := &dto.RoleInput{Name: "editor", Description: "d", PermissionIDs: []uint{1, 2}}

	t.Run("Success", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		perms := []models.Permission{{ID: 1}, {ID: 2}}
		roleRepo.On("FindByName", ctx, "editor").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1, 2}).Return(perms, nil)
		roleRepo.On("Create", ctx, &models.Role{Name: "editor", Description: "d", Permissions: perms}).Return(nil)

		role, err := svc.CreateRole(ctx, input)

		require.NoError(t, err)
		assert.Equal(t, "editor", role.Name)
		roleRepo.AssertExpectations(t)
	})

	t.Run("Duplicate Name Is Conflict", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(&models.Role{ID: 9, Name: "editor"}, nil)

		_, err := svc.CreateRole(ctx, input)

		assertAppError(t, err, http.StatusConflict)
		roleRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("Unknown Permission ID Is Bad Request", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1, 2}).Return([]models.Permission{{ID: 1}}, nil)

		_, err := svc.CreateRole(ctx, input)

		assertAppError(t, err, http.StatusBadRequest)
	})

	t.Run("Duplicate Permission IDs Are Collapsed", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		perms := []models.Permission{{ID: 1}}
		roleRepo.On("FindByName", ctx, "editor").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1}).Return(perms, nil)
		roleRepo.On("Create", ctx, mock.Anything).Return(nil)

		_, err := svc.CreateRole(ctx, &dto.RoleInput{Name: "editor", PermissionIDs: []uint{1, 1}})

		require.NoError(t, err)
	})

	t.Run("Repository Error Is Returned", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(nil, apperror.NewInternalServerError("boom"))

		_, err := svc.CreateRole(ctx, input)

		assertAppError(t, err, http.StatusInternalServerError)
	})
}

func TestRoleService_UpdateRole(t *testing.T) {
	ctx := context.Background()
	input := &dto.RoleInput{Name: "writer", Description: "new", PermissionIDs: []uint{2}}

	t.Run("Success", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		perms := []models.Permission{{ID: 2}}
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "editor"}, nil)
		roleRepo.On("FindByName", ctx, "writer").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{2}).Return(perms, nil)
		roleRepo.On("Update", ctx, &models.Role{ID: 5, Name: "writer", Description: "new", Permissions: perms}).Return(nil)

		role, err := svc.UpdateRole(ctx, 5, input)

		require.NoError(t, err)
		assert.Equal(t, "writer", role.Name)
	})

	t.Run("Keeping Own Name Is Allowed", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "writer"}, nil)
		roleRepo.On("FindByName", ctx, "writer").Return(&models.Role{ID: 5, Name: "writer"}, nil)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{2}).Return([]models.Permission{{ID: 2}}, nil)
		roleRepo.On("Update", ctx, mock.Anything).Return(nil)

		_, err := svc.UpdateRole(ctx, 5, input)

		require.NoError(t, err)
	})

	t.Run("Name Taken By Another Role Is Conflict", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "editor"}, nil)
		roleRepo.On("FindByName", ctx, "writer").Return(&models.Role{ID: 6, Name: "writer"}, nil)

		_, err := svc.UpdateRole(ctx, 5, input)

		assertAppError(t, err, http.StatusConflict)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(nil, errRoleNotFound)

		_, err := svc.UpdateRole(ctx, 5, input)

		assertAppError(t, err, http.StatusNotFound)
	})

	t.Run("Admin Role Is Forbidden", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(1)).Return(&models.Role{ID: 1, Name: models.RoleAdmin}, nil)

		_, err := svc.UpdateRole(ctx, 1, input)

		assertAppError(t, err, http.StatusForbidden)
		roleRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})
}

func TestRoleService_DeleteRole(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "editor"}, nil)
		roleRepo.On("Delete", ctx, uint(5)).Return(nil)

		require.NoError(t, svc.DeleteRole(ctx, 5))
		roleRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(nil, errRoleNotFound)

		assertAppError(t, svc.DeleteRole(ctx, 5), http.StatusNotFound)
	})

	t.Run("Admin Role Is Forbidden", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(1)).Return(&models.Role{ID: 1, Name: models.RoleAdmin}, nil)

		assertAppError(t, svc.DeleteRole(ctx, 1), http.StatusForbidden)
		roleRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	})
}

func TestRoleService_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("ListRoles", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindAll", ctx).Return([]models.Role{{ID: 1}}, nil)

		roles, err := svc.ListRoles(ctx)

		require.NoError(t, err)
		assert.Len(t, roles, 1)
	})

	t.Run("GetRole", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(1)).Return(&models.Role{ID: 1}, nil)

		role, err := svc.GetRole(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, uint(1), role.ID)
	})

	t.Run("ListPermissions", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindPermissions", ctx).Return([]models.Permission{{ID: 1}}, nil)

		perms, err := svc.ListPermissions(ctx)

		require.NoError(t, err)
		assert.Len(t, perms, 1)
	})
}

func TestRoleService_SetUserRoles(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1, 2}).Return(int64(2), nil)
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{1, 2}).Return(nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{{ID: 1}, {ID: 2}}, nil)

		roles, err := svc.SetUserRoles(ctx, 3, []uint{1, 2, 1})

		require.NoError(t, err)
		assert.Len(t, roles, 2)
	})

	t.Run("Empty Clears Roles Without Counting", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{}).Return(nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{}, nil)

		_, err := svc.SetUserRoles(ctx, 3, []uint{})

		require.NoError(t, err)
		roleRepo.AssertNotCalled(t, "CountRolesByIDs", mock.Anything, mock.Anything)
	})

	t.Run("User Not Found", func(t *testing.T) {
		svc, _, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(nil, apperror.New(http.StatusNotFound, apperror.ErrNotFound, "User not found"))

		_, err := svc.SetUserRoles(ctx, 3, []uint{1})

		assertAppError(t, err, http.StatusNotFound)
	})

	t.Run("Unknown Role ID Is Bad Request", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1, 2}).Return(int64(1), nil)

		_, err := svc.SetUserRoles(ctx, 3, []uint{1, 2})

		assertAppError(t, err, http.StatusBadRequest)
		roleRepo.AssertNotCalled(t, "SetUserRoles", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Repository Error", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1}).Return(int64(0), errors.New("db down"))

		_, err := svc.SetUserRoles(ctx, 3, []uint{1})

		assert.Error(t, err)
	})
}

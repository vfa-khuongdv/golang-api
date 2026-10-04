package services_test

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

var errRoleNotFound = apperror.New(http.StatusNotFound, apperror.ErrNotFound, "Role not found")

// actorID performs the changes in these tests and holds every permission, like
// an admin. limitedActorID holds only permission 2. Role 2 grants only
// permission 2; any other set of roles includes permission 1.
const (
	actorID        = uint(99)
	limitedActorID = uint(50)
)

func isOnlyRole2(ids []uint) bool { return slices.Equal(ids, []uint{2}) }

func newRoleService() (services.RoleService, *mocks.MockRoleRepository, *mocks.MockUserRepository) {
	roleRepo := new(mocks.MockRoleRepository)
	userRepo := new(mocks.MockUserRepository)
	roleRepo.On("FindPermissionIDsByUserID", mock.Anything, actorID).Return([]uint{1, 2, 3, 4, 5, 6, 7, 8}, nil).Maybe()
	roleRepo.On("FindPermissionIDsByUserID", mock.Anything, limitedActorID).Return([]uint{2}, nil).Maybe()
	roleRepo.On("FindPermissionIDsByRoleIDs", mock.Anything, mock.MatchedBy(isOnlyRole2)).Return([]uint{2}, nil).Maybe()
	roleRepo.On("FindPermissionIDsByRoleIDs", mock.Anything, mock.MatchedBy(func(ids []uint) bool { return !isOnlyRole2(ids) })).Return([]uint{1}, nil).Maybe()
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

		role, err := svc.CreateRole(ctx, actorID, input)

		require.NoError(t, err)
		assert.Equal(t, "editor", role.Name)
		roleRepo.AssertExpectations(t)
	})

	t.Run("Without Permissions Returns An Empty List Not Nil", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "bare").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{}).Return(nil, nil)
		roleRepo.On("Create", ctx, mock.Anything).Return(nil)

		role, err := svc.CreateRole(ctx, actorID, &dto.RoleInput{Name: "bare"})

		require.NoError(t, err)
		assert.NotNil(t, role.Permissions)
		assert.Empty(t, role.Permissions)
	})

	t.Run("Duplicate Name Is Conflict", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(&models.Role{ID: 9, Name: "editor"}, nil)

		_, err := svc.CreateRole(ctx, actorID, input)

		assertAppError(t, err, http.StatusConflict)
		roleRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("Unknown Permission ID Is Bad Request", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1, 2}).Return([]models.Permission{{ID: 1}}, nil)

		_, err := svc.CreateRole(ctx, actorID, input)

		assertAppError(t, err, http.StatusBadRequest)
	})

	t.Run("Duplicate Permission IDs Are Collapsed", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		perms := []models.Permission{{ID: 1}}
		roleRepo.On("FindByName", ctx, "editor").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1}).Return(perms, nil)
		roleRepo.On("Create", ctx, mock.Anything).Return(nil)

		_, err := svc.CreateRole(ctx, actorID, &dto.RoleInput{Name: "editor", PermissionIDs: []uint{1, 1}})

		require.NoError(t, err)
	})

	t.Run("Repository Error Is Returned", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(nil, apperror.NewInternalServerError("boom"))

		_, err := svc.CreateRole(ctx, actorID, input)

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

		role, err := svc.UpdateRole(ctx, actorID, 5, input)

		require.NoError(t, err)
		assert.Equal(t, "writer", role.Name)
	})

	t.Run("Keeping Own Name Is Allowed", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "writer"}, nil)
		roleRepo.On("FindByName", ctx, "writer").Return(&models.Role{ID: 5, Name: "writer"}, nil)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{2}).Return([]models.Permission{{ID: 2}}, nil)
		roleRepo.On("Update", ctx, mock.Anything).Return(nil)

		_, err := svc.UpdateRole(ctx, actorID, 5, input)

		require.NoError(t, err)
	})

	t.Run("Name Taken By Another Role Is Conflict", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "editor"}, nil)
		roleRepo.On("FindByName", ctx, "writer").Return(&models.Role{ID: 6, Name: "writer"}, nil)

		_, err := svc.UpdateRole(ctx, actorID, 5, input)

		assertAppError(t, err, http.StatusConflict)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(nil, errRoleNotFound)

		_, err := svc.UpdateRole(ctx, actorID, 5, input)

		assertAppError(t, err, http.StatusNotFound)
	})

	t.Run("Admin Role Is Forbidden", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(1)).Return(&models.Role{ID: 1, Name: models.RoleAdmin}, nil)

		_, err := svc.UpdateRole(ctx, actorID, 1, input)

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

		require.NoError(t, svc.DeleteRole(ctx, actorID, 5))
		roleRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(nil, errRoleNotFound)

		assertAppError(t, svc.DeleteRole(ctx, actorID, 5), http.StatusNotFound)
	})

	t.Run("Admin Role Is Forbidden", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(1)).Return(&models.Role{ID: 1, Name: models.RoleAdmin}, nil)

		assertAppError(t, svc.DeleteRole(ctx, actorID, 1), http.StatusForbidden)
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
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{1, 2}, uint(0)).Return(nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{{ID: 1}, {ID: 2}}, nil)

		roles, err := svc.SetUserRoles(ctx, actorID, 3, []uint{1, 2, 1})

		require.NoError(t, err)
		assert.Len(t, roles, 2)
	})

	t.Run("Empty Clears Roles Without Counting", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{}, uint(0)).Return(nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{}, nil)

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{})

		require.NoError(t, err)
		roleRepo.AssertNotCalled(t, "CountRolesByIDs", mock.Anything, mock.Anything)
	})

	t.Run("User Not Found", func(t *testing.T) {
		svc, _, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(nil, apperror.New(http.StatusNotFound, apperror.ErrNotFound, "User not found"))

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{1})

		assertAppError(t, err, http.StatusNotFound)
	})

	t.Run("Unknown Role ID Is Bad Request", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1, 2}).Return(int64(1), nil)

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{1, 2})

		assertAppError(t, err, http.StatusBadRequest)
		roleRepo.AssertNotCalled(t, "SetUserRoles", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Repository Error", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1}).Return(int64(0), errors.New("db down"))

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{1})

		assert.Error(t, err)
	})
}

func TestRoleService_SetUserRoles_KeepsAnAdmin(t *testing.T) {
	ctx := context.Background()
	admin := models.Role{ID: 1, Name: models.RoleAdmin}

	setup := func() (services.RoleService, *mocks.MockRoleRepository) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{admin}, nil)
		return svc, roleRepo
	}

	t.Run("Removing The Admin Role Asks The Repository To Keep An Admin", func(t *testing.T) {
		svc, roleRepo := setup()
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{}, admin.ID).Return(nil)

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{})

		require.NoError(t, err)
	})

	t.Run("Removing The Last Admin Is Forbidden", func(t *testing.T) {
		svc, roleRepo := setup()
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{}, admin.ID).Return(repositories.ErrLastRoleHolder)

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{})

		assertAppError(t, err, http.StatusForbidden)
	})

	t.Run("Swapping The Last Admin For Another Role Is Forbidden", func(t *testing.T) {
		svc, roleRepo := setup()
		roleRepo.On("CountRolesByIDs", ctx, []uint{2}).Return(int64(1), nil)
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{2}, admin.ID).Return(repositories.ErrLastRoleHolder)

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{2})

		assertAppError(t, err, http.StatusForbidden)
	})

	t.Run("Keeping The Admin Role Needs No Guard", func(t *testing.T) {
		svc, roleRepo := setup()
		roleRepo.On("CountRolesByIDs", ctx, []uint{1, 2}).Return(int64(2), nil)
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{1, 2}, uint(0)).Return(nil)

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{1, 2})

		require.NoError(t, err)
	})

	t.Run("Repository Errors Are Returned", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return(nil, errors.New("db down"))

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{})

		assert.Error(t, err)
	})

	t.Run("Write Error Is Returned", func(t *testing.T) {
		svc, roleRepo := setup()
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{}, admin.ID).Return(errors.New("db down"))

		_, err := svc.SetUserRoles(ctx, actorID, 3, []uint{})

		assert.EqualError(t, err, "db down")
	})
}

func TestRoleService_PreventsPrivilegeEscalation(t *testing.T) {
	ctx := context.Background()
	forbidden := func(t *testing.T, err error) {
		t.Helper()
		assertAppError(t, err, http.StatusForbidden)
	}

	t.Run("Creating A Role With A Permission The Actor Lacks", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "editor").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1, 2}).Return([]models.Permission{{ID: 1}, {ID: 2}}, nil)

		_, err := svc.CreateRole(ctx, limitedActorID, &dto.RoleInput{Name: "editor", PermissionIDs: []uint{1, 2}})

		forbidden(t, err)
		roleRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("Creating A Role With Only Permissions The Actor Holds Is Allowed", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByName", ctx, "viewer").Return(nil, errRoleNotFound)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{2}).Return([]models.Permission{{ID: 2}}, nil)
		roleRepo.On("Create", ctx, mock.Anything).Return(nil)

		_, err := svc.CreateRole(ctx, limitedActorID, &dto.RoleInput{Name: "viewer", PermissionIDs: []uint{2}})

		require.NoError(t, err)
	})

	t.Run("Adding A Permission The Actor Lacks To A Role", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "viewer", Permissions: []models.Permission{{ID: 2}}}, nil)
		roleRepo.On("FindByName", ctx, "viewer").Return(&models.Role{ID: 5, Name: "viewer"}, nil)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{1, 2}).Return([]models.Permission{{ID: 1}, {ID: 2}}, nil)

		_, err := svc.UpdateRole(ctx, limitedActorID, 5, &dto.RoleInput{Name: "viewer", PermissionIDs: []uint{1, 2}})

		forbidden(t, err)
		roleRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("Changing A Role That Holds A Permission The Actor Lacks", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "editor", Permissions: []models.Permission{{ID: 1}, {ID: 2}}}, nil)
		roleRepo.On("FindByName", ctx, "editor").Return(&models.Role{ID: 5, Name: "editor"}, nil)
		roleRepo.On("FindPermissionsByIDs", ctx, []uint{2}).Return([]models.Permission{{ID: 2}}, nil)

		_, err := svc.UpdateRole(ctx, limitedActorID, 5, &dto.RoleInput{Name: "editor", PermissionIDs: []uint{2}})

		forbidden(t, err)
		roleRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("Deleting A Role That Holds A Permission The Actor Lacks", func(t *testing.T) {
		svc, roleRepo, _ := newRoleService()
		roleRepo.On("FindByID", ctx, uint(5)).Return(&models.Role{ID: 5, Name: "editor", Permissions: []models.Permission{{ID: 1}}}, nil)

		forbidden(t, svc.DeleteRole(ctx, limitedActorID, 5))
		roleRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	})

	t.Run("Assigning A Role With A Permission The Actor Lacks, e.g. Admin To Themselves", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, limitedActorID).Return(&models.User{ID: limitedActorID}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1}).Return(int64(1), nil)
		roleRepo.On("FindByUserID", ctx, limitedActorID).Return([]models.Role{}, nil)

		_, err := svc.SetUserRoles(ctx, limitedActorID, limitedActorID, []uint{1})

		forbidden(t, err)
		roleRepo.AssertNotCalled(t, "SetUserRoles", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Removing A Role With A Permission The Actor Lacks", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{{ID: 1, Name: "editor"}}, nil)

		_, err := svc.SetUserRoles(ctx, limitedActorID, 3, []uint{})

		forbidden(t, err)
		roleRepo.AssertNotCalled(t, "SetUserRoles", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Only Changed Roles Are Checked", func(t *testing.T) {
		svc, roleRepo, userRepo := newRoleService()
		userRepo.On("GetByID", ctx, uint(3)).Return(&models.User{ID: 3}, nil)
		roleRepo.On("CountRolesByIDs", ctx, []uint{1, 2}).Return(int64(2), nil)
		roleRepo.On("FindByUserID", ctx, uint(3)).Return([]models.Role{{ID: 1, Name: "editor"}}, nil)
		roleRepo.On("SetUserRoles", ctx, uint(3), []uint{1, 2}, uint(0)).Return(nil)

		// Role 1 is kept as it is, so only role 2 is checked.
		_, err := svc.SetUserRoles(ctx, limitedActorID, 3, []uint{1, 2})

		require.NoError(t, err)
		roleRepo.AssertCalled(t, "FindPermissionIDsByRoleIDs", mock.Anything, []uint{2})
	})
}

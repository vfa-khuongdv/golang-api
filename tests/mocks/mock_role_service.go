package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
)

type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleService) ListRoles(ctx context.Context) ([]models.Role, error) {
	args := m.Called(ctx)
	roles, _ := args.Get(0).([]models.Role)
	return roles, args.Error(1)
}

func (m *MockRoleService) GetRole(ctx context.Context, id uint) (*models.Role, error) {
	args := m.Called(ctx, id)
	role, _ := args.Get(0).(*models.Role)
	return role, args.Error(1)
}

func (m *MockRoleService) CreateRole(ctx context.Context, actorID uint, input *dto.RoleInput) (*models.Role, error) {
	args := m.Called(ctx, actorID, input)
	role, _ := args.Get(0).(*models.Role)
	return role, args.Error(1)
}

func (m *MockRoleService) UpdateRole(ctx context.Context, actorID uint, id uint, input *dto.RoleInput) (*models.Role, error) {
	args := m.Called(ctx, actorID, id, input)
	role, _ := args.Get(0).(*models.Role)
	return role, args.Error(1)
}

func (m *MockRoleService) DeleteRole(ctx context.Context, actorID uint, id uint) error {
	return m.Called(ctx, actorID, id).Error(0)
}

func (m *MockRoleService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	args := m.Called(ctx)
	perms, _ := args.Get(0).([]models.Permission)
	return perms, args.Error(1)
}

func (m *MockRoleService) SetUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint) ([]models.Role, error) {
	args := m.Called(ctx, actorID, userID, roleIDs)
	roles, _ := args.Get(0).([]models.Role)
	return roles, args.Error(1)
}

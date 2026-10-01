package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
)

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) HasPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepository) FindAll(ctx context.Context) ([]models.Role, error) {
	args := m.Called(ctx)
	roles, _ := args.Get(0).([]models.Role)
	return roles, args.Error(1)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id uint) (*models.Role, error) {
	args := m.Called(ctx, id)
	role, _ := args.Get(0).(*models.Role)
	return role, args.Error(1)
}

func (m *MockRoleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	args := m.Called(ctx, name)
	role, _ := args.Get(0).(*models.Role)
	return role, args.Error(1)
}

func (m *MockRoleRepository) FindByUserID(ctx context.Context, userID uint) ([]models.Role, error) {
	args := m.Called(ctx, userID)
	roles, _ := args.Get(0).([]models.Role)
	return roles, args.Error(1)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *models.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRoleRepository) FindPermissions(ctx context.Context) ([]models.Permission, error) {
	args := m.Called(ctx)
	perms, _ := args.Get(0).([]models.Permission)
	return perms, args.Error(1)
}

func (m *MockRoleRepository) FindPermissionsByIDs(ctx context.Context, ids []uint) ([]models.Permission, error) {
	args := m.Called(ctx, ids)
	perms, _ := args.Get(0).([]models.Permission)
	return perms, args.Error(1)
}

func (m *MockRoleRepository) CountRolesByIDs(ctx context.Context, ids []uint) (int64, error) {
	args := m.Called(ctx, ids)
	n, _ := args.Get(0).(int64)
	return n, args.Error(1)
}

func (m *MockRoleRepository) SetUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return m.Called(ctx, userID, roleIDs).Error(0)
}

package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUsers(ctx context.Context, page int, limit int) (*dto.Pagination[*models.User], error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Pagination[*models.User]), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, userId uint) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func (m *MockUserRepository) FindByField(ctx context.Context, field string, value string) (*models.User, error) {
	args := m.Called(ctx, field, value)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, user *models.User) (*models.User, error) {
	args := m.Called(ctx, tx, user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) BeginTx(ctx context.Context) (*gorm.DB, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gorm.DB), args.Error(1)
}

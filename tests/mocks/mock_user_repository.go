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

func (m *MockUserRepository) UpdateColumns(ctx context.Context, user *models.User, columns ...string) error {
	args := m.Called(ctx, user, columns)
	return args.Error(0)
}

func (m *MockUserRepository) RecordFailedLogin(ctx context.Context, userID uint, maxAttempts int, lockUntil int64) error {
	args := m.Called(ctx, userID, maxAttempts, lockUntil)
	return args.Error(0)
}

func (m *MockUserRepository) ResetFailedLogins(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, userId uint) error {
	args := m.Called(ctx, userId)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByResetToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
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

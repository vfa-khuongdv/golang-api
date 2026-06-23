package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"gorm.io/gorm"
)

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) Update(ctx context.Context, token *models.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) UpdateWithTx(ctx context.Context, tx *gorm.DB, token *models.RefreshToken) error {
	args := m.Called(ctx, tx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByTokenWithTx(ctx context.Context, tx *gorm.DB, token string) (*models.RefreshToken, error) {
	args := m.Called(ctx, tx, token)
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) BeginTx(ctx context.Context) (*gorm.DB, error) {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

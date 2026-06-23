package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
)

type MockRefreshTokenService struct {
	mock.Mock
}

func (m *MockRefreshTokenService) Create(ctx context.Context, user *models.User, ipAddress string) (*dto.JwtResult, error) {
	args := m.Called(ctx, user, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*dto.JwtResult)
	return result, args.Error(1)
}

func (m *MockRefreshTokenService) Update(ctx context.Context, token string, ipAddress string) (*dto.RefreshTokenResult, error) {
	args := m.Called(ctx, token, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*dto.RefreshTokenResult)
	return result, args.Error(1)
}

func (m *MockRefreshTokenService) DeleteByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

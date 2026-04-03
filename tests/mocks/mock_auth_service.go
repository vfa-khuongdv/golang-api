package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, email string, password string, ipAddress string) (*dto.LoginResponse, error) {
	args := m.Called(ctx, email, password, ipAddress)
	if res, ok := args.Get(0).(*dto.LoginResponse); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken, accessToken string, ipAddress string) (*dto.LoginResponse, error) {
	args := m.Called(ctx, refreshToken, accessToken, ipAddress)
	if res, ok := args.Get(0).(*dto.LoginResponse); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

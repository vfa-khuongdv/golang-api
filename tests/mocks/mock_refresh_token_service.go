package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

type MockRefreshTokenService struct {
	mock.Mock
}

func (m *MockRefreshTokenService) Create(user *models.User, ipAddress string) (*dto.JwtResult, error) {
	args := m.Called(user, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*dto.JwtResult)
	return result, args.Error(1)
}

func (m *MockRefreshTokenService) Update(token string, ipAddress string) (*services.RefreshTokenResult, error) {
	args := m.Called(token, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*services.RefreshTokenResult)
	return result, args.Error(1)
}

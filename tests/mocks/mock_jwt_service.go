package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

// MockJWTService is a mock implementation of services.IJWTService
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(id uint) (*services.JwtResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.JwtResult), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*services.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CustomClaims), args.Error(1)
}

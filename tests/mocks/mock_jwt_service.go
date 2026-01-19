package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

// MockJWTService is a mock implementation of services.JWTService
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateAccessToken(id uint) (*dto.JwtResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.JwtResult), args.Error(1)
}

func (m *MockJWTService) GenerateMfaToken(id uint) (*dto.JwtResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.JwtResult), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*services.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CustomClaims), args.Error(1)
}

func (m *MockJWTService) ValidateTokenWithScope(tokenString string, requiredScope string) (*services.CustomClaims, error) {
	args := m.Called(tokenString, requiredScope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CustomClaims), args.Error(1)
}

func (m *MockJWTService) ValidateTokenIgnoreExpiration(tokenString string) (*services.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CustomClaims), args.Error(1)
}

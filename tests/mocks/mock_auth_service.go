package mocks

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
)

type MockAuthService struct {
	mock.Mock
}

// Login provides a mock function with given fields: email, password, ctx
func (m *MockAuthService) Login(email string, password string, ctx *gin.Context) (*dto.LoginResponse, error) {
	args := m.Called(email, password, ctx)
	if res, ok := args.Get(0).(*dto.LoginResponse); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken, accessToken string, ctx *gin.Context) (*dto.LoginResponse, error) {
	args := m.Called(refreshToken, accessToken, ctx)
	if res, ok := args.Get(0).(*dto.LoginResponse); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

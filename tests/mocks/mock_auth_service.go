package mocks

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(email, password string, ctx *gin.Context) (interface{}, error) {
	args := m.Called(email, password, ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockAuthService) RefreshToken(token string, ctx *gin.Context) (*services.LoginResponse, error) {
	args := m.Called(token, ctx)
	if res, ok := args.Get(0).(*services.LoginResponse); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

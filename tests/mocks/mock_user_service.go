package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(userID uint) (*models.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(userID uint, input *dto.UpdateProfileInput) error {
	args := m.Called(userID, input)
	return args.Error(0)
}

func (m *MockUserService) ForgotPassword(input *dto.ForgotPasswordInput) (*models.User, error) {
	args := m.Called(input)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ResetPassword(input *dto.ResetPasswordInput) (*models.User, error) {
	args := m.Called(input)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(userId uint, input *dto.ChangePasswordInput) (*models.User, error) {
	args := m.Called(userId, input)
	return args.Get(0).(*models.User), args.Error(1)
}

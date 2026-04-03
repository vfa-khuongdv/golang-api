package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint, input *dto.UpdateProfileInput) error {
	args := m.Called(ctx, userID, input)
	return args.Error(0)
}

func (m *MockUserService) ForgotPassword(ctx context.Context, input *dto.ForgotPasswordInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(ctx context.Context, input *dto.ResetPasswordInput) (*models.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ChangePassword(ctx context.Context, userId uint, input *dto.ChangePasswordInput) (*models.User, error) {
	args := m.Called(ctx, userId, input)
	return args.Get(0).(*models.User), args.Error(1)
}

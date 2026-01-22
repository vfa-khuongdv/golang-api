package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
)

type MockMailerService struct {
	mock.Mock
}

func (m *MockMailerService) SendMailForgotPassword(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

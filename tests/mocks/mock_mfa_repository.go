package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
)

type MockMfaRepository struct {
	mock.Mock
}

func NewMockMfaRepository() *MockMfaRepository {
	return &MockMfaRepository{}
}

func (m *MockMfaRepository) GetMfaSettingsByUserID(userID uint) (*models.MfaSettings, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MfaSettings), args.Error(1)
}

func (m *MockMfaRepository) CreateMfaSettings(settings *models.MfaSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockMfaRepository) UpdateMfaSettings(settings *models.MfaSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockMfaRepository) DeleteMfaSettings(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

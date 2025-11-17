package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockMfaService is a mock implementation of services.IMfaService
type MockMfaService struct {
	mock.Mock
}

func (m *MockMfaService) SetupMfa(userID uint, email string) (string, []byte, []string, error) {
	args := m.Called(userID, email)
	secret, _ := args.Get(0).(string)
	qrCodeBytes, _ := args.Get(1).([]byte)
	backupCodes, _ := args.Get(2).([]string)
	return secret, qrCodeBytes, backupCodes, args.Error(3)
}

func (m *MockMfaService) VerifyMfaSetup(userID uint, totpCode string) ([]string, error) {
	args := m.Called(userID, totpCode)
	backupCodes, _ := args.Get(0).([]string)
	return backupCodes, args.Error(1)
}

func (m *MockMfaService) VerifyMfaCode(userID uint, totpCode string) (bool, error) {
	args := m.Called(userID, totpCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockMfaService) DisableMfa(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockMfaService) GetMfaStatus(userID uint) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockMfaService) ValidateBackupCode(userID uint, code string) (bool, error) {
	args := m.Called(userID, code)
	return args.Bool(0), args.Error(1)
}

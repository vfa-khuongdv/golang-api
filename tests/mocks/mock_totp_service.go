package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockTotpService struct {
	mock.Mock
}

func NewMockTotpService() *MockTotpService {
	return &MockTotpService{}
}

func (m *MockTotpService) GenerateSecret(email string) (string, error) {
	return "test-secret-key", nil
}

func (m *MockTotpService) GetQRCode(secret string, email string) ([]byte, error) {
	return []byte("mock-qr-code"), nil
}

func (m *MockTotpService) VerifyCode(secret string, code string) (bool, error) {
	return true, nil
}

func (m *MockTotpService) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = "BACKUP" + string(rune(i))
	}
	return codes, nil
}

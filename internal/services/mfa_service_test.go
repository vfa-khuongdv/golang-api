package services_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

func TestMfaService(t *testing.T) {
	t.Run("NewMfaService", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)

		// Act
		service := services.NewMfaService(mockRepo, mockTotp)

		// Assert
		assert.NotNil(t, service)
	})

	t.Run("SetupMfa - Already Enabled", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)
		email := "test@example.com"
		secret := "existing-secret"

		existingSettings := &models.MfaSettings{
			UserID:     userID,
			TotpSecret: &secret,
			MfaEnabled: true,
		}

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(existingSettings, nil).Once()

		// Act
		resultSecret, resultQR, resultCodes, err := service.SetupMfa(userID, email)

		// Assert
		require.Error(t, err)
		assert.Empty(t, resultSecret)
		assert.Nil(t, resultQR)
		assert.Nil(t, resultCodes)

		var appErr *apperror.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ErrMfaAlreadyEnabled, appErr.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("SetupMfa - Repository Error", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)
		email := "test@example.com"

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(nil, errors.New("database error")).Once()

		// Act
		resultSecret, resultQR, resultCodes, err := service.SetupMfa(userID, email)

		// Assert
		require.Error(t, err)
		assert.Empty(t, resultSecret)
		assert.Nil(t, resultQR)
		assert.Nil(t, resultCodes)
		assert.ErrorContains(t, err, "failed to retrieve MFA settings")

		mockRepo.AssertExpectations(t)
	})

	t.Run("VerifyMfaSetup - Setup Not Initiated", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)
		totpCode := "123456"

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(nil, nil).Once()

		// Act
		resultCodes, err := service.VerifyMfaSetup(userID, totpCode)

		// Assert
		require.Error(t, err)
		assert.Nil(t, resultCodes)

		var appErr *apperror.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ErrMfaSetupNotInitiated, appErr.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("VerifyMfaCode - Not Enabled", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)
		totpCode := "123456"

		settings := &models.MfaSettings{
			UserID:     userID,
			MfaEnabled: false,
		}

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(settings, nil).Once()

		// Act
		valid, err := service.VerifyMfaCode(userID, totpCode)

		// Assert
		require.Error(t, err)
		assert.False(t, valid)

		var appErr *apperror.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, apperror.ErrMfaNotEnabled, appErr.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidateBackupCode - Success", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)
		backupCode := "valid-code"
		backupCodes := []string{backupCode, "code2", "code3"}

		settings := &models.MfaSettings{
			UserID:      userID,
			BackupCodes: mustMarshal(backupCodes),
			MfaEnabled:  true,
		}

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(settings, nil).Once()
		mockRepo.On("UpdateMfaSettings", mock.MatchedBy(func(s *models.MfaSettings) bool {
			return s.UserID == userID
		})).Return(nil).Once()

		// Act
		valid, err := service.ValidateBackupCode(userID, backupCode)

		// Assert
		require.NoError(t, err)
		assert.True(t, valid)

		// Verify backup code was removed
		var updatedCodes []string
		json.Unmarshal(settings.BackupCodes, &updatedCodes)
		assert.NotContains(t, updatedCodes, backupCode)
		assert.Len(t, updatedCodes, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ValidateBackupCode - Invalid Code", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)
		invalidCode := "invalid-code"
		backupCodes := []string{"code1", "code2", "code3"}

		settings := &models.MfaSettings{
			UserID:      userID,
			BackupCodes: mustMarshal(backupCodes),
			MfaEnabled:  true,
		}

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(settings, nil).Once()

		// Act
		valid, err := service.ValidateBackupCode(userID, invalidCode)

		// Assert
		require.NoError(t, err)
		assert.False(t, valid)

		mockRepo.AssertExpectations(t)
	})

	t.Run("DisableMfa", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)

		mockRepo.On("DeleteMfaSettings", userID).Return(nil).Once()

		// Act
		err := service.DisableMfa(userID)

		// Assert
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetMfaStatus - Enabled", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)

		settings := &models.MfaSettings{
			UserID:     userID,
			MfaEnabled: true,
		}

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(settings, nil).Once()

		// Act
		enabled, err := service.GetMfaStatus(userID)

		// Assert
		require.NoError(t, err)
		assert.True(t, enabled)

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetMfaStatus - Disabled", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockMfaRepository)
		mockTotp := new(mocks.MockTotpService)
		service := services.NewMfaService(mockRepo, mockTotp)

		userID := uint(1)

		mockRepo.On("GetMfaSettingsByUserID", userID).Return(nil, nil).Once()

		// Act
		enabled, err := service.GetMfaStatus(userID)

		// Assert
		require.NoError(t, err)
		assert.False(t, enabled)

		mockRepo.AssertExpectations(t)
	})
}

// Helper function to marshal backup codes to JSON
func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

package services

import (
	"encoding/json"
	"fmt"

	"github.com/vfa-khuongdv/golang-cms/internal/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
)

// IMfaService defines operations for Multi-Factor Authentication management
type IMfaService interface {
	// SetupMfa initiates MFA setup for a user, generating a TOTP secret and backup codes
	SetupMfa(userID uint, email string) (secret string, qrCodeBytes []byte, backupCodes []string, err error)
	// VerifyMfaSetup verifies the TOTP code provided during MFA setup
	VerifyMfaSetup(userID uint, totpCode string) (backupCodes []string, err error)
	// VerifyMfaCode verifies a TOTP code during login or sensitive operations
	VerifyMfaCode(userID uint, totpCode string) (bool, error)
	// DisableMfa disables MFA for a user
	DisableMfa(userID uint) error
	// GetMfaStatus retrieves the MFA status for a user
	GetMfaStatus(userID uint) (bool, error)
	// ValidateBackupCode validates and uses a backup code
	ValidateBackupCode(userID uint, code string) (bool, error)
}

type MfaService struct {
	mfaRepository IMfaRepository
	totpService   ITotpService
}

type IMfaRepository interface {
	GetMfaSettingsByUserID(userID uint) (*models.MfaSettings, error)
	CreateMfaSettings(settings *models.MfaSettings) error
	UpdateMfaSettings(settings *models.MfaSettings) error
	DeleteMfaSettings(userID uint) error
}

// NewMfaService creates a new instance of MfaService
func NewMfaService(mfaRepository IMfaRepository, totpService ITotpService) IMfaService {
	return &MfaService{
		mfaRepository: mfaRepository,
		totpService:   totpService,
	}
}

// SetupMfa initiates MFA setup for a user, generating a TOTP secret and backup codes
func (s *MfaService) SetupMfa(userID uint, email string) (string, []byte, []string, error) {
	// Check if MFA is already enabled
	settings, err := s.mfaRepository.GetMfaSettingsByUserID(userID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to retrieve MFA settings: %w", err)
	}

	if settings != nil && settings.MfaEnabled {
		return "", nil, nil, apperror.NewMfaAlreadyEnabledError("MFA is already enabled for this user")
	}

	// Generate TOTP secret
	secret, err := s.totpService.GenerateSecret(email)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate backup codes
	backupCodes, err := s.totpService.GenerateBackupCodes(constants.MFA_BACKUP_CODE_COUNT)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Generate QR code
	qrCode, err := s.totpService.GetQRCode(secret, email)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert backup codes to JSON
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to marshal backup codes: %w", err)
	}

	if settings == nil {
		// Create new MFA settings with temporary secret (MFA not yet enabled)
		settings = &models.MfaSettings{
			UserID:      userID,
			TotpSecret:  &secret,
			BackupCodes: backupCodesJSON,
			MfaEnabled:  false,
		}
		if err := s.mfaRepository.CreateMfaSettings(settings); err != nil {
			return "", nil, nil, fmt.Errorf("failed to create MFA settings: %w", err)
		}
	} else {
		// Update existing MFA settings with new temporary secret
		settings.TotpSecret = &secret
		settings.BackupCodes = backupCodesJSON
		settings.MfaEnabled = false
		if err := s.mfaRepository.UpdateMfaSettings(settings); err != nil {
			return "", nil, nil, fmt.Errorf("failed to update MFA settings: %w", err)
		}
	}

	return secret, qrCode, backupCodes, nil
}

// VerifyMfaSetup verifies the TOTP code provided during MFA setup
func (s *MfaService) VerifyMfaSetup(userID uint, totpCode string) ([]string, error) {
	// Get or create MFA settings from database
	settings, err := s.mfaRepository.GetMfaSettingsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve MFA settings: %w", err)
	}

	if settings == nil || settings.TotpSecret == nil {
		return nil, apperror.NewMfaSetupNotInitiatedError("MFA setup session expired or not initiated")
	}

	// Verify TOTP code with the stored secret
	valid, err := s.totpService.VerifyCode(*settings.TotpSecret, totpCode)
	if err != nil || !valid {
		return nil, apperror.NewMfaInvalidCodeError("Invalid TOTP code")
	}

	// Get backup codes
	var backupCodes []string
	if settings.BackupCodes != nil {
		if err := json.Unmarshal(settings.BackupCodes, &backupCodes); err != nil {
			return nil, apperror.NewParseError("Invalid backup codes in settings")
		}
	}

	// Enable MFA in settings
	settings.MfaEnabled = true
	if err := s.mfaRepository.UpdateMfaSettings(settings); err != nil {
		return nil, fmt.Errorf("failed to update MFA settings: %w", err)
	}

	return backupCodes, nil
}

// VerifyMfaCode verifies a TOTP code during login or sensitive operations
func (s *MfaService) VerifyMfaCode(userID uint, totpCode string) (bool, error) {
	settings, err := s.mfaRepository.GetMfaSettingsByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve MFA settings: %w", err)
	}

	if settings == nil || !settings.MfaEnabled {
		return false, apperror.NewMfaNotEnabledError("MFA is not enabled for this user")
	}

	if settings.TotpSecret == nil {
		return false, apperror.NewMfaNotEnabledError("MFA is not properly configured")
	}

	// Try TOTP code first
	valid, err := s.totpService.VerifyCode(*settings.TotpSecret, totpCode)
	if err != nil {
		return false, fmt.Errorf("failed to verify TOTP code: %w", err)
	}

	if valid {
		return true, nil
	}

	// If TOTP fails, check backup codes
	isBackupCode, err := s.ValidateBackupCode(userID, totpCode)
	if err != nil {
		return false, fmt.Errorf("failed to validate backup code: %w", err)
	}

	return isBackupCode, nil
}

// ValidateBackupCode validates and uses a backup code
func (s *MfaService) ValidateBackupCode(userID uint, code string) (bool, error) {
	settings, err := s.mfaRepository.GetMfaSettingsByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve MFA settings: %w", err)
	}

	if settings == nil {
		return false, apperror.NewMfaNotEnabledError("MFA settings not found")
	}

	var backupCodes []string
	if settings.BackupCodes != nil {
		if err := json.Unmarshal(settings.BackupCodes, &backupCodes); err != nil {
			return false, fmt.Errorf("failed to parse backup codes: %w", err)
		}
	}

	// Find and remove the backup code
	for i, backupCode := range backupCodes {
		if backupCode == code {
			// Remove used code
			backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
			updatedCodes, _ := json.Marshal(backupCodes)
			settings.BackupCodes = updatedCodes
			if err := s.mfaRepository.UpdateMfaSettings(settings); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

// DisableMfa disables MFA for a user
func (s *MfaService) DisableMfa(userID uint) error {
	return s.mfaRepository.DeleteMfaSettings(userID)
}

// GetMfaStatus retrieves the MFA status for a user
func (s *MfaService) GetMfaStatus(userID uint) (bool, error) {
	settings, err := s.mfaRepository.GetMfaSettingsByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve MFA settings: %w", err)
	}

	if settings == nil {
		return false, nil
	}

	return settings.MfaEnabled, nil
}

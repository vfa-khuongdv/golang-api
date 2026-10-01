package services

import (
	"context"
	"strconv"
	"strings"

	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/dto"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type SettingService interface {
	GetSettings(ctx context.Context) (*dto.SettingsResponse, error)
	UpdateSettings(ctx context.Context, input *dto.UpdateSettingsInput) error
}

type settingServiceImpl struct {
	settingRepo   repositories.SettingRepository
	encryptionKey string
}

// NewSettingService creates a SettingService. encryptionKey encrypts the
// mail.password setting before it is stored (see utils.EncryptSecret).
func NewSettingService(settingRepo repositories.SettingRepository, encryptionKey string) SettingService {
	return &settingServiceImpl{settingRepo: settingRepo, encryptionKey: encryptionKey}
}

func (s *settingServiceImpl) GetSettings(ctx context.Context) (*dto.SettingsResponse, error) {
	values, err := s.settingRepo.GetValues(ctx,
		constants.SettingMailHost,
		constants.SettingMailPort,
		constants.SettingMailUsername,
		constants.SettingMailPassword,
		constants.SettingMailFrom,
		constants.SettingFrontendURL,
	)
	if err != nil {
		return nil, err
	}

	// An unparsable port is reported as 0 (not configured).
	port, _ := strconv.Atoi(values[constants.SettingMailPort])

	return &dto.SettingsResponse{
		MailHost:        values[constants.SettingMailHost],
		MailPort:        port,
		MailUsername:    values[constants.SettingMailUsername],
		MailPasswordSet: values[constants.SettingMailPassword] != "",
		MailFrom:        values[constants.SettingMailFrom],
		FrontendURL:     values[constants.SettingFrontendURL],
	}, nil
}

func (s *settingServiceImpl) UpdateSettings(ctx context.Context, input *dto.UpdateSettingsInput) error {
	values := make(map[string]string)

	if input.MailHost != nil {
		values[constants.SettingMailHost] = *input.MailHost
	}
	if input.MailPort != nil {
		values[constants.SettingMailPort] = strconv.Itoa(*input.MailPort)
	}
	if input.MailUsername != nil {
		values[constants.SettingMailUsername] = *input.MailUsername
	}
	if input.MailPassword != nil {
		password := *input.MailPassword
		if password != "" {
			encrypted, err := utils.EncryptSecret(s.encryptionKey, password)
			if err != nil {
				logger.WithContext(ctx).Errorf("Failed to encrypt %s setting: %v", constants.SettingMailPassword, err)
				return apperror.NewInternalServerError("Failed to update settings")
			}
			password = encrypted
		}
		values[constants.SettingMailPassword] = password
	}
	if input.MailFrom != nil {
		values[constants.SettingMailFrom] = *input.MailFrom
	}
	if input.FrontendURL != nil {
		// The mailer appends "/reset-password" to this value.
		values[constants.SettingFrontendURL] = strings.TrimRight(*input.FrontendURL, "/")
	}

	if len(values) == 0 {
		return nil
	}
	return s.settingRepo.SetValues(ctx, values)
}

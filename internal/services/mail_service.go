package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/constants"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"github.com/vfa-khuongdv/golang-cms/pkg/mailer"
)

type MailerService interface {
	SendMailForgotPassword(ctx context.Context, user *models.User) error
}

type mailerServiceImpl struct {
	settingRepo   repositories.SettingRepository
	encryptionKey string
}

var (
	newEmailSender = func(config mailer.GomailSenderConfig) mailer.EmailSender {
		// Return a nil interface, not an interface holding a nil *GomailSender,
		// so the caller's nil check catches an invalid mail configuration.
		if sender := mailer.NewGomailSender(config); sender != nil {
			return sender
		}
		return nil
	}
	parseForgotTemplate = func() (*template.Template, error) {
		return template.ParseFS(mailer.ForgotTemplate, "templates/forgot_template.html")
	}
)

// NewMailerService creates a MailerService. encryptionKey decrypts the
// mail.password setting, which is stored encrypted (see utils.EncryptSecret).
func NewMailerService(settingRepo repositories.SettingRepository, encryptionKey string) MailerService {
	return &mailerServiceImpl{settingRepo: settingRepo, encryptionKey: encryptionKey}
}

func (s *mailerServiceImpl) SendMailForgotPassword(ctx context.Context, user *models.User) error {
	settings, err := s.settingRepo.GetValues(ctx,
		constants.SettingMailHost,
		constants.SettingMailPort,
		constants.SettingMailUsername,
		constants.SettingMailPassword,
		constants.SettingMailFrom,
		constants.SettingFrontendURL,
	)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(settings[constants.SettingMailPort])
	if err != nil {
		logger.WithContext(ctx).Errorf("Invalid %s setting: %v", constants.SettingMailPort, err)
		return apperror.NewInternalServerError("Mail settings are not configured")
	}

	frontendURL := settings[constants.SettingFrontendURL]
	if frontendURL == "" {
		logger.WithContext(ctx).Errorf("Missing %s setting", constants.SettingFrontendURL)
		return apperror.NewInternalServerError("Frontend URL setting is not configured")
	}

	// An empty password is allowed (e.g. local Mailpit without SMTP AUTH).
	password := settings[constants.SettingMailPassword]
	if password != "" {
		password, err = utils.DecryptSecret(s.encryptionKey, password)
		if err != nil {
			logger.WithContext(ctx).Errorf("Failed to decrypt %s setting: %v", constants.SettingMailPassword, err)
			return apperror.NewInternalServerError("Mail settings are not configured")
		}
	}

	sender := newEmailSender(mailer.GomailSenderConfig{
		Host:     settings[constants.SettingMailHost],
		Port:     port,
		Username: settings[constants.SettingMailUsername],
		Password: password,
		From:     settings[constants.SettingMailFrom],
	})
	if sender == nil {
		return apperror.NewInternalServerError("Failed to initialize mail sender")
	}

	tmpl, err := parseForgotTemplate()
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	if user.ResetToken == nil {
		return apperror.NewInternalServerError("user reset token is nil")
	}

	url := frontendURL + "/reset-password?token=" + *user.ResetToken

	data := map[string]interface{}{
		"Name": user.Name,
		"URL":  url,
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		// Log the details but never echo them to the client.
		logger.Errorf("Failed to execute forgot-password template: %v", err)
		return apperror.NewInternalServerError("Failed to generate email content")
	}

	if err := sender.Send([]string{user.Email}, "Reset your password", "", htmlBody.String()); err != nil {
		// Log the details but never echo them to the client.
		logger.Errorf("Failed to send forgot-password email to %s: %v", utils.MaskWithPrefix(user.Email, 4), err)
		return apperror.NewInternalServerError("Failed to send email")
	}
	return nil

}

package services

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/pkg/mailer"
)

type MailerService interface {
	SendMailForgotPassword(user *models.User) error
}

type mailerServiceImpl struct{}

var (
	newEmailSender = func(config mailer.GomailSenderConfig) mailer.EmailSender {
		return mailer.NewGomailSender(config)
	}
	parseForgotTemplate = func() (*template.Template, error) {
		return template.ParseFS(mailer.ForgotTemplate, "templates/forgot_template.html")
	}
)

func NewMailerService() MailerService {
	return &mailerServiceImpl{}
}

func (s *mailerServiceImpl) SendMailForgotPassword(user *models.User) error {

	var config = mailer.GomailSenderConfig{
		Host:     configs.GetEnv("MAIL_HOST", "smtp.gmail.com"),
		Port:     configs.GetEnvAsInt("MAIL_PORT", 587),
		Username: configs.GetEnv("MAIL_USERNAME", ""),
		Password: configs.GetEnv("MAIL_PASSWORD", ""),
		From:     configs.GetEnv("MAIL_FROM", ""),
	}

	sender := newEmailSender(mailer.GomailSenderConfig{
		From:     config.From,
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
	})

	tmpl, err := parseForgotTemplate()
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	url := configs.GetEnv("FRONTEND_URL", "") + "/reset-password?token=" + *user.ResetToken

	data := map[string]interface{}{
		"Name": user.Name,
		"URL":  url,
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return apperror.NewInternalServerError(fmt.Sprintf("error executing template: %+v", err))
	}

	if err := sender.Send([]string{user.Email}, "Reset your password", "", htmlBody.String()); err != nil {
		return apperror.NewInternalServerError(fmt.Sprintf("error sending email: %+v", err))
	}
	return nil

}

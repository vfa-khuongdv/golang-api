package services

import (
	"context"
	"errors"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/mailer"
)

type fakeEmailSender struct {
	sendErr  error
	lastHTML string
	lastTo   []string
}

func (f *fakeEmailSender) Send(to []string, _ string, _ string, html string) error {
	f.lastTo = to
	f.lastHTML = html
	return f.sendErr
}

type fakeSettingRepository struct {
	values map[string]string
	err    error
}

func (f *fakeSettingRepository) GetValues(_ context.Context, _ ...string) (map[string]string, error) {
	return f.values, f.err
}

func validMailSettings() map[string]string {
	return map[string]string{
		"mail_host":     "smtp.example.com",
		"mail_port":     "587",
		"mail_username": "user",
		"mail_password": "secret",
		"mail_from":     "noreply@example.com",
		"frontend_url":  "https://example.com",
	}
}

func TestMailerService_InternalBranches(t *testing.T) {
	originalSender := newEmailSender
	originalParse := parseForgotTemplate
	t.Cleanup(func() {
		newEmailSender = originalSender
		parseForgotTemplate = originalParse
	})

	token := "reset-token"
	user := &models.User{
		Email:      "user@example.com",
		Name:       "User",
		ResetToken: &token,
	}

	settingRepo := &fakeSettingRepository{values: validMailSettings()}

	t.Run("Success", func(t *testing.T) {
		fake := &fakeEmailSender{}
		var gotConfig mailer.GomailSenderConfig
		newEmailSender = func(config mailer.GomailSenderConfig) mailer.EmailSender {
			gotConfig = config
			return fake
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService(settingRepo).SendMailForgotPassword(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, mailer.GomailSenderConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "secret",
			From:     "noreply@example.com",
		}, gotConfig)
		assert.Equal(t, []string{"user@example.com"}, fake.lastTo)
		assert.Contains(t, fake.lastHTML, "https://example.com/reset-password?token=reset-token")
		assert.Contains(t, fake.lastHTML, "User")
	})

	t.Run("SendErrorStillWrapped", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{sendErr: errors.New("smtp fail")}
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService(settingRepo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to send email")
		// The underlying SMTP failure must stay server-side only.
		assert.NotContains(t, err.Error(), "smtp fail")
	})

	t.Run("NilEmailSender", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return nil
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService(settingRepo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to initialize mail sender")
	})

	t.Run("TemplateParseError", func(t *testing.T) {
		parseForgotTemplate = func() (*template.Template, error) {
			return nil, errors.New("parse failure")
		}
		t.Cleanup(func() { parseForgotTemplate = originalParse })

		err := NewMailerService(settingRepo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing template")
	})

	t.Run("TemplateExecuteError", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{}
		}
		t.Cleanup(func() { newEmailSender = originalSender })
		parseForgotTemplate = func() (*template.Template, error) {
			tmpl := template.New("test")
			tmpl = tmpl.Funcs(template.FuncMap{
				"fail": func() (string, error) {
					return "", errors.New("execution failure")
				},
			})
			return template.Must(tmpl.Parse(`{{fail}}`)), nil
		}
		t.Cleanup(func() { parseForgotTemplate = originalParse })

		err := NewMailerService(settingRepo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to generate email content")
		// The template execution error must stay server-side only.
		assert.NotContains(t, err.Error(), "execution failure")
	})
	t.Run("SettingRepositoryError", func(t *testing.T) {
		repoErr := errors.New("db down")
		repo := &fakeSettingRepository{err: repoErr}

		err := NewMailerService(repo).SendMailForgotPassword(context.Background(), user)
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("InvalidMailPort", func(t *testing.T) {
		for _, port := range []string{"", "abc", "58 7"} {
			values := validMailSettings()
			values["mail_port"] = port
			repo := &fakeSettingRepository{values: values}

			err := NewMailerService(repo).SendMailForgotPassword(context.Background(), user)
			assert.Error(t, err, "port %q", port)
			assert.Contains(t, err.Error(), "Mail settings are not configured")
		}
	})

	t.Run("MissingMailPortKey", func(t *testing.T) {
		values := validMailSettings()
		delete(values, "mail_port")
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mail settings are not configured")
	})

	t.Run("MissingFrontendURL", func(t *testing.T) {
		values := validMailSettings()
		delete(values, "frontend_url")
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Frontend URL setting is not configured")
	})

	t.Run("EmptyFrontendURL", func(t *testing.T) {
		values := validMailSettings()
		values["frontend_url"] = ""
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Frontend URL setting is not configured")
	})
}

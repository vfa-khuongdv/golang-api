package services

import (
	"errors"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/pkg/mailer"
)

type fakeEmailSender struct {
	sendErr error
}

func (f *fakeEmailSender) Send(_ []string, _ string, _ string, _ string) error {
	return f.sendErr
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
		Email: "user@example.com",
		Name:  "User",
		ResetToken: &token,
	}

	t.Setenv("FRONTEND_URL", "https://example.com")

	t.Run("Success", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{}
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.NoError(t, err)
	})

	t.Run("SendErrorStillWrapped", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{sendErr: errors.New("smtp fail")}
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending email")
	})

	t.Run("TemplateParseError", func(t *testing.T) {
		parseForgotTemplate = func() (*template.Template, error) {
			return nil, errors.New("parse failure")
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error parsing template")
	})

	t.Run("TemplateExecuteError", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{}
		}
		parseForgotTemplate = func() (*template.Template, error) {
			tmpl := template.New("test")
			tmpl = tmpl.Funcs(template.FuncMap{
				"fail": func() (string, error) {
					return "", errors.New("execution failure")
				},
			})
			return template.Must(tmpl.Parse(`{{fail}}`)), nil
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error executing template")
	})
}

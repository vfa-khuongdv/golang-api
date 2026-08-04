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
	sendErr    error
	lastHTML   string
	lastTo     []string
	lastFrom   string
	lastConfig mailer.GomailSenderConfig
}

func (f *fakeEmailSender) Send(to []string, _ string, _ string, html string) error {
	f.lastTo = to
	f.lastHTML = html
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
		fake := &fakeEmailSender{}
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return fake
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService().SendMailForgotPassword(user)
		assert.NoError(t, err)
		assert.Equal(t, []string{"user@example.com"}, fake.lastTo)
		assert.Contains(t, fake.lastHTML, "https://example.com/reset-password?token=reset-token")
		assert.Contains(t, fake.lastHTML, "User")
	})

	t.Run("SendErrorStillWrapped", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{sendErr: errors.New("smtp fail")}
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending email")
	})

	t.Run("NilEmailSender", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return nil
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to initialize mail sender")
	})

	t.Run("TemplateParseError", func(t *testing.T) {
		parseForgotTemplate = func() (*template.Template, error) {
			return nil, errors.New("parse failure")
		}
		t.Cleanup(func() { parseForgotTemplate = originalParse })

		err := NewMailerService().SendMailForgotPassword(user)
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

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error executing template")
	})
}

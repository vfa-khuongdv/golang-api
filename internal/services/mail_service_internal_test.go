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
	originalParse := parseTemplateFile
	t.Cleanup(func() {
		newEmailSender = originalSender
		parseTemplateFile = originalParse
	})

	token := "reset-token"
	user := &models.User{
		Email: "user@example.com",
		Name:  "User",
		Token: &token,
	}

	t.Setenv("FRONTEND_URL", "https://example.com")

	t.Run("TemplateExecuteError", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{}
		}
		parseTemplateFile = func(_ ...string) (*template.Template, error) {
			return template.Must(template.New("bad").Parse(`{{.Name.Field}}`)), nil
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error executing template")
	})

	t.Run("Success", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{}
		}
		parseTemplateFile = func(_ ...string) (*template.Template, error) {
			return template.Must(template.New("ok").Parse(`Hi {{.Name}} - {{.URL}}`)), nil
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.NoError(t, err)
	})

	t.Run("SendErrorStillWrapped", func(t *testing.T) {
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			return &fakeEmailSender{sendErr: errors.New("smtp fail")}
		}
		parseTemplateFile = func(_ ...string) (*template.Template, error) {
			return template.Must(template.New("ok").Parse(`Hi {{.Name}}`)), nil
		}

		err := NewMailerService().SendMailForgotPassword(user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error sending email")
	})
}

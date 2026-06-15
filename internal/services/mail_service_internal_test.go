package services

import (
	"errors"
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
	t.Cleanup(func() {
		newEmailSender = originalSender
	})

	token := "reset-token"
	user := &models.User{
		Email: "user@example.com",
		Name:  "User",
		Token: &token,
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
}

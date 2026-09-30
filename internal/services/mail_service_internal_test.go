package services

import (
	"context"
	"errors"
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
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

// Low-entropy dummy key so secret scanners do not flag it.
var testEncryptionKey = strings.Repeat("a", 40)

func encryptForTest(t *testing.T, plaintext string) string {
	t.Helper()
	encrypted, err := utils.EncryptSecret(testEncryptionKey, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt test value: %v", err)
	}
	return encrypted
}

func validMailSettings(t *testing.T) map[string]string {
	return map[string]string{
		"mail.host":        "smtp.example.com",
		"mail.port":        "587",
		"mail.username":    "user",
		"mail.password":    encryptForTest(t, "secret"),
		"mail.from":        "noreply@example.com",
		"app.frontend_url": "https://example.com",
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

	settingRepo := &fakeSettingRepository{values: validMailSettings(t)}

	t.Run("Success", func(t *testing.T) {
		fake := &fakeEmailSender{}
		var gotConfig mailer.GomailSenderConfig
		newEmailSender = func(config mailer.GomailSenderConfig) mailer.EmailSender {
			gotConfig = config
			return fake
		}
		t.Cleanup(func() { newEmailSender = originalSender })

		err := NewMailerService(settingRepo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
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

		err := NewMailerService(settingRepo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
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

		err := NewMailerService(settingRepo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to initialize mail sender")
	})

	t.Run("TemplateParseError", func(t *testing.T) {
		parseForgotTemplate = func() (*template.Template, error) {
			return nil, errors.New("parse failure")
		}
		t.Cleanup(func() { parseForgotTemplate = originalParse })

		err := NewMailerService(settingRepo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
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

		err := NewMailerService(settingRepo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to generate email content")
		// The template execution error must stay server-side only.
		assert.NotContains(t, err.Error(), "execution failure")
	})
	t.Run("SettingRepositoryError", func(t *testing.T) {
		repoErr := errors.New("db down")
		repo := &fakeSettingRepository{err: repoErr}

		err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("InvalidMailPort", func(t *testing.T) {
		for _, port := range []string{"", "abc", "58 7"} {
			values := validMailSettings(t)
			values["mail.port"] = port
			repo := &fakeSettingRepository{values: values}

			err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
			assert.Error(t, err, "port %q", port)
			assert.Contains(t, err.Error(), "Mail settings are not configured")
		}
	})

	t.Run("MissingMailPortKey", func(t *testing.T) {
		values := validMailSettings(t)
		delete(values, "mail.port")
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mail settings are not configured")
	})

	t.Run("MissingFrontendURL", func(t *testing.T) {
		values := validMailSettings(t)
		delete(values, "app.frontend_url")
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Frontend URL setting is not configured")
	})

	t.Run("EmptyFrontendURL", func(t *testing.T) {
		values := validMailSettings(t)
		values["app.frontend_url"] = ""
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Frontend URL setting is not configured")
	})
	t.Run("EmptyPasswordSkipsDecryption", func(t *testing.T) {
		var gotConfig mailer.GomailSenderConfig
		newEmailSender = func(config mailer.GomailSenderConfig) mailer.EmailSender {
			gotConfig = config
			return &fakeEmailSender{}
		}
		t.Cleanup(func() { newEmailSender = originalSender })
		values := validMailSettings(t)
		values["mail.username"] = ""
		values["mail.password"] = ""
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.NoError(t, err)
		assert.Empty(t, gotConfig.Password)
	})

	t.Run("PlaintextPasswordRejected", func(t *testing.T) {
		called := false
		newEmailSender = func(_ mailer.GomailSenderConfig) mailer.EmailSender {
			called = true
			return &fakeEmailSender{}
		}
		t.Cleanup(func() { newEmailSender = originalSender })
		values := validMailSettings(t)
		values["mail.password"] = "secret"
		repo := &fakeSettingRepository{values: values}

		err := NewMailerService(repo, testEncryptionKey).SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mail settings are not configured")
		assert.False(t, called, "sender must not be built with an unencrypted password")
	})

	t.Run("PasswordEncryptedWithOtherKey", func(t *testing.T) {
		repo := &fakeSettingRepository{values: validMailSettings(t)}

		err := NewMailerService(repo, testEncryptionKey+"-rotated").SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Mail settings are not configured")
		// Neither the ciphertext nor decryption details leak to the client.
		assert.NotContains(t, err.Error(), utils.EncryptedPrefix)
		assert.NotContains(t, err.Error(), "decrypt")
	})
}

package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
)

type mailerServiceTestSuite struct {
	suite.Suite
	mailerService services.MailerService
}

func (s *mailerServiceTestSuite) SetupTest() {
	s.mailerService = services.NewMailerService()
}

func (s *mailerServiceTestSuite) TestSendMailForgotPassword() {
	s.T().Run("Nil Token", func(t *testing.T) {
		t.Setenv("MAIL_HOST", "smtp.gmail.com")
		t.Setenv("MAIL_PORT", "587")
		t.Setenv("MAIL_USERNAME", "test@example.com")
		t.Setenv("MAIL_PASSWORD", "testpassword")
		t.Setenv("MAIL_FROM", "noreply@example.com")
		t.Setenv("FRONTEND_URL", "https://example.com")

		user := &models.User{
			ID:    1,
			Email: "user@example.com",
			Name:  "Test User",
			ResetToken: nil,
		}

		assert.Panics(t, func() {
			_ = s.mailerService.SendMailForgotPassword(user)
		})
	})
}

func TestMailerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(mailerServiceTestSuite))
}

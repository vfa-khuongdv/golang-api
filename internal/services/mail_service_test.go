package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
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

		err := s.mailerService.SendMailForgotPassword(user)
		assert.Error(t, err)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrInternalServer, appErr.Code)
		}
	})
}

func TestMailerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(mailerServiceTestSuite))
}

package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/apperror"
	"github.com/vfa-khuongdv/golang-cms/tests/mocks"
)

type mailerServiceTestSuite struct {
	suite.Suite
	mailerService services.MailerService
	settingRepo   *mocks.MockSettingRepository
}

func (s *mailerServiceTestSuite) SetupTest() {
	s.settingRepo = new(mocks.MockSettingRepository)
	s.mailerService = services.NewMailerService(s.settingRepo, strings.Repeat("a", 40))
}

func (s *mailerServiceTestSuite) TestSendMailForgotPassword() {
	s.T().Run("Nil Token", func(t *testing.T) {
		s.settingRepo.On("GetValues", mock.Anything, mock.Anything).Return(map[string]string{
			"mail.host":        "smtp.gmail.com",
			"mail.port":        "587",
			"mail.username":    "test@example.com",
			"mail.password":    "",
			"mail.from":        "noreply@example.com",
			"app.frontend_url": "https://example.com",
		}, nil).Once()

		user := &models.User{
			ID:         1,
			Email:      "user@example.com",
			Name:       "Test User",
			ResetToken: nil,
		}

		err := s.mailerService.SendMailForgotPassword(context.Background(), user)
		assert.Error(t, err)
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrInternalServer, appErr.Code)
		}
	})

	s.T().Run("Empty Host Returns Error Instead Of Panicking", func(t *testing.T) {
		s.settingRepo.On("GetValues", mock.Anything, mock.Anything).Return(map[string]string{
			"mail.host":        "",
			"mail.port":        "587",
			"mail.from":        "noreply@example.com",
			"app.frontend_url": "https://example.com",
		}, nil).Once()
		token := "raw-token"
		user := &models.User{ID: 1, Email: "user@example.com", Name: "Test User", ResetToken: &token}

		var err error
		assert.NotPanics(t, func() {
			err = s.mailerService.SendMailForgotPassword(context.Background(), user)
		})
		var appErr *apperror.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, apperror.ErrInternalServer, appErr.Code)
		}
	})
}

func TestMailerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(mailerServiceTestSuite))
}

package mailer_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/pkg/mailer"
	"gopkg.in/gomail.v2"
)

type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(msgs ...*gomail.Message) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func TestGomailSender_Send(t *testing.T) {
	// Arrange
	config := mailer.GomailSenderConfig{
		From:     "test@example.com",
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	}
	mockDialer := new(MockDialer)
	sender := &mailer.GomailSender{
		Config: config,
		Dialer: mockDialer,
	}

	t.Run("should return error if no recipients", func(t *testing.T) {
		// Act
		err := sender.Send([]string{}, "Subject", "text", "html")

		// Assert
		assert.EqualError(t, err, "recipient list cannot be empty")
	})

	t.Run("should return error if subject is empty", func(t *testing.T) {
		// Act
		err := sender.Send([]string{"a@b.com"}, "", "text", "html")

		// Assert
		assert.EqualError(t, err, "email subject cannot be empty")
	})

	t.Run("should return error if both plainText and html are empty", func(t *testing.T) {
		// Act
		err := sender.Send([]string{"a@b.com"}, "Subject", "", "")

		// Assert
		assert.EqualError(t, err, "either plain text or HTML content must be provided")
	})

	t.Run("should send successfully", func(t *testing.T) {
		// Arrange
		mockDialer.On("DialAndSend", mock.Anything).Return(nil).Once()

		// Act
		err := sender.Send([]string{"a@b.com"}, "Subject", "Hello", "<b>Hi</b>")

		// Assert
		assert.NoError(t, err)
		mockDialer.AssertExpectations(t)
	})

	t.Run("should return error from dialer", func(t *testing.T) {
		// Arrange
		mockDialer.On("DialAndSend", mock.Anything).Return(errors.New("smtp error")).Once()

		// Act
		err := sender.Send([]string{"a@b.com"}, "Subject", "text", "")

		// Assert
		assert.EqualError(t, err, "smtp error")
		mockDialer.AssertExpectations(t)
	})
}

func TestNewGomailSender(t *testing.T) {
	// Arrange
	config := mailer.GomailSenderConfig{
		From:     "test@example.com",
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	}

	// Act
	sender := mailer.NewGomailSender(config)

	// Assert
	assert.NotNil(t, sender)
	assert.Equal(t, config.From, sender.Config.From)
	assert.Equal(t, config.Host, sender.Config.Host)
	assert.Equal(t, config.Port, sender.Config.Port)
	assert.Equal(t, config.Username, sender.Config.Username)
	assert.Equal(t, config.Password, sender.Config.Password)
	assert.NotNil(t, sender.Dialer)
}

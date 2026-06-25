package mailer_test

import (
	"errors"
	"testing"

	mail "github.com/wneessen/go-mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vfa-khuongdv/golang-cms/pkg/mailer"
)

type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(msgs ...*mail.Msg) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func TestGomailSender_Send(t *testing.T) {
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
		err := sender.Send([]string{}, "Subject", "text", "html")
		assert.EqualError(t, err, "recipient list cannot be empty")
	})

	t.Run("should return error if subject is empty", func(t *testing.T) {
		err := sender.Send([]string{"a@b.com"}, "", "text", "html")
		assert.EqualError(t, err, "email subject cannot be empty")
	})

	t.Run("should return error if both plainText and html are empty", func(t *testing.T) {
		err := sender.Send([]string{"a@b.com"}, "Subject", "", "")
		assert.EqualError(t, err, "either plain text or HTML content must be provided")
	})

	t.Run("should send successfully", func(t *testing.T) {
		mockDialer.On("DialAndSend", mock.Anything).Return(nil).Once()

		err := sender.Send([]string{"a@b.com"}, "Subject", "Hello", "<b>Hi</b>")

		assert.NoError(t, err)
		mockDialer.AssertExpectations(t)
	})

	t.Run("should return error from dialer", func(t *testing.T) {
		mockDialer.On("DialAndSend", mock.Anything).Return(errors.New("smtp error")).Once()

		err := sender.Send([]string{"a@b.com"}, "Subject", "text", "")

		assert.EqualError(t, err, "smtp error")
		mockDialer.AssertExpectations(t)
	})
}

func TestNewGomailSender_InvalidConfig(t *testing.T) {
	config := mailer.GomailSenderConfig{
		From:     "test@example.com",
		Host:     "smtp.example.com",
		Port:     0,
		Username: "user",
		Password: "pass",
	}

	sender := mailer.NewGomailSender(config)
	assert.Nil(t, sender)
}

func TestGomailSender_Send_InvalidFrom(t *testing.T) {
	mockDialer := new(MockDialer)
	sender := &mailer.GomailSender{
		Config: mailer.GomailSenderConfig{From: "invalid"},
		Dialer: mockDialer,
	}

	err := sender.Send([]string{"a@b.com"}, "Subject", "text", "")
	assert.Error(t, err)
}

func TestGomailSender_Send_InvalidTo(t *testing.T) {
	mockDialer := new(MockDialer)
	sender := &mailer.GomailSender{
		Config: mailer.GomailSenderConfig{From: "test@example.com"},
		Dialer: mockDialer,
	}

	err := sender.Send([]string{"invalid"}, "Subject", "text", "")
	assert.Error(t, err)
}

func TestNewGomailSender(t *testing.T) {
	config := mailer.GomailSenderConfig{
		From:     "test@example.com",
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
	}

	sender := mailer.NewGomailSender(config)

	assert.NotNil(t, sender)
	assert.Equal(t, config.From, sender.Config.From)
	assert.Equal(t, config.Host, sender.Config.Host)
	assert.Equal(t, config.Port, sender.Config.Port)
	assert.Equal(t, config.Username, sender.Config.Username)
	assert.Equal(t, config.Password, sender.Config.Password)
	assert.NotNil(t, sender.Dialer)
}

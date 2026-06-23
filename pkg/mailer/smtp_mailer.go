package mailer

import (
	"errors"

	"gopkg.in/gomail.v2"
)

type EmailSender interface {
	Send(to []string, subject, plainText, html string) error
}

type GomailSenderConfig struct {
	From     string
	Host     string
	Port     int
	Username string
	Password string
}

// Interface for mocking DialAndSend
type DialAndSender interface {
	DialAndSend(m ...*gomail.Message) error
}

type GomailSender struct {
	Config GomailSenderConfig
	Dialer DialAndSender
}

func NewGomailSender(config GomailSenderConfig) *GomailSender {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	return &GomailSender{
		Config: config,
		Dialer: dialer,
	}
}

func (s *GomailSender) Send(to []string, subject, plainText, html string) error {
	if len(to) == 0 {
		return errors.New("recipient list cannot be empty")
	}
	if subject == "" {
		return errors.New("email subject cannot be empty")
	}
	if plainText == "" && html == "" {
		return errors.New("either plain text or HTML content must be provided")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.Config.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	if plainText != "" {
		m.SetBody("text/plain", plainText)
	}
	if html != "" {
		m.AddAlternative("text/html", html)
	}

	if err := s.Dialer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

package mailer

import (
	"errors"

	mail "github.com/wneessen/go-mail"
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

type DialAndSender interface {
	DialAndSend(msgs ...*mail.Msg) error
}

type GomailSender struct {
	Config GomailSenderConfig
	Dialer DialAndSender
}

func NewGomailSender(config GomailSenderConfig) *GomailSender {
	opts := []mail.Option{mail.WithPort(config.Port)}
	if config.Username != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthLogin),
			mail.WithUsername(config.Username),
			mail.WithPassword(config.Password),
		)
	} else {
		// No credentials (e.g. local Mailpit): skip SMTP AUTH and only use
		// STARTTLS when the server offers it. With credentials, the default
		// mandatory TLS policy stays in place so passwords never go out in clear.
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(config.Host, opts...)
	if err != nil {
		return nil
	}
	return &GomailSender{
		Config: config,
		Dialer: client,
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

	m := mail.NewMsg()
	if err := m.From(s.Config.From); err != nil {
		return err
	}
	if err := m.To(to...); err != nil {
		return err
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, plainText)
	if html != "" {
		m.AddAlternativeString(mail.TypeTextHTML, html)
	}

	if err := s.Dialer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

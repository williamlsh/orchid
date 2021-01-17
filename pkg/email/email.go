package email

import (
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// ConfigOptions includes all mail config options.
type ConfigOptions struct {
	From     string
	Host     string
	Port     int
	Username string
	Passwd   string
}

// Mail is a configured email.
type Mail struct {
	logger *zap.SugaredLogger
	ConfigOptions
	to      string
	subject string
}

// New returns a new mail.
func New(logger *zap.SugaredLogger, conf ConfigOptions, to, subject string) Mail {
	return Mail{logger, conf, to, subject}
}

// Send sends email.
func (m Mail) Send(content string) error {
	m.logger.Debugf("Send mail from %s, to %s, subject: %s, content: %s", m.From, m.to, m.subject, content)

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", m.to)
	msg.SetHeader("Subject", m.subject)
	msg.SetBody("text/html", content)

	m.logger.Debugf("Dial mail host: %s port: %d username: %s password: xxx", m.Host, m.Port, m.Username)

	// TODO: Considering changing to mail daemon with only one mail connection for all sending emails.
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Passwd)
	return d.DialAndSend(msg)
}

package email

import "gopkg.in/gomail.v2"

// ConfigOptions includes all mail config options.
type ConfigOptions struct {
	From     string
	Host     string
	Port     int
	Username string
	Passwd   string
}

type Mail struct {
	ConfigOptions
	to      string
	subject string
}

// New returns a new mail.
func New(conf ConfigOptions, to, subject string) Mail {
	return Mail{conf, to, subject}
}

// Send sends email.
func (m Mail) Send(code string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", m.to)
	msg.SetHeader("Subject", "Sign in to xxx")
	msg.SetBody("text/html", "Please click http://localhost/m/callback?token="+code+"&operation=login&state=overseatu")

	// TODO: Considering changing to mail daemon with only one mail connection for all sending emails.
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Passwd)
	return d.DialAndSend(msg)
}

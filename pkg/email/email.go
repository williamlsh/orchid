package email

import "gopkg.in/gomail.v2"

type Config struct {
	from     string
	host     string
	port     int
	username string
	passwd   string
}

type Mail struct {
	Config
	to      string
	subject string
}

// New returns a new mail.
func New(conf Config, to, subject string) Mail {
	return Mail{conf, to, subject}
}

// Send sends email.
func (m Mail) Send(code string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", m.to)
	msg.SetHeader("Subject", "Sign in to xxx")
	msg.SetBody("text/html", "Your sign in verification code is: "+code)

	d := gomail.NewDialer(m.host, m.port, m.username, m.passwd)
	return d.DialAndSend(msg)
}

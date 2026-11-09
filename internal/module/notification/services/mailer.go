package services

import (
	"crypto/tls"
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

type MailerConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromName  string
	FromEmail string
}

type Mailer interface {
	Send(to []string, subject, htmlBody string) error
	Configured() bool
}

type smtpMailer struct {
	cfg MailerConfig
}

func NewMailer(cfg MailerConfig) Mailer {
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) Configured() bool {
	return m.cfg.Host != "" && m.cfg.Username != "" && m.cfg.Password != "" && m.cfg.FromEmail != ""
}

func (m *smtpMailer) Send(to []string, subject, htmlBody string) error {
	if len(to) == 0 {
		return nil
	}

	d := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.Username, m.cfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: false, ServerName: m.cfg.Host}

	sc, err := d.Dial()
	if err != nil {
		return fmt.Errorf("mailer.Dial: %w", err)
	}
	defer sc.Close()

	from := fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.FromEmail)

	for _, recipient := range to {
		msg := gomail.NewMessage()
		msg.SetHeader("From", from)
		msg.SetHeader("To", recipient)
		msg.SetHeader("Subject", subject)
		msg.SetBody("text/html", htmlBody)

		if err := gomail.Send(sc, msg); err != nil {
			// continue on individual failure; return last error
			err = fmt.Errorf("mailer.Send to %s: %w", recipient, err)
			continue
		}
	}
	return nil
}

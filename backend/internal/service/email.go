package service

import (
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

// EmailService sends transactional emails via SMTP, or logs them in console mode.
type EmailService struct {
	dialer      *gomail.Dialer
	from        string
	consoleMode bool
}

// NewEmailService creates an email sender. If host is empty or "console", emails are logged instead of sent.
func NewEmailService(host string, port int, user, pass string) *EmailService {
	if host == "" || host == "console" {
		return &EmailService{consoleMode: true}
	}
	return &EmailService{
		dialer: gomail.NewDialer(host, port, user, pass),
		from:   user,
	}
}

// SendEmail sends an HTML email to the given recipient, or logs it in console mode.
func (es *EmailService) SendEmail(to, subject, htmlBody string) error {
	if es.consoleMode {
		log.Printf("[EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, htmlBody)
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", es.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}
	return nil
}

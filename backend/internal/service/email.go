package service

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

// EmailService sends transactional emails via SMTP, or logs them in console mode for local dev.
type EmailService struct {
	host        string
	port        string
	user        string
	pass        string
	consoleMode bool
}

// NewEmailService creates an email sender. If host is empty or "console", emails are logged instead of sent.
func NewEmailService(host, port, user, pass string) *EmailService {
	if host == "" || host == "console" {
		return &EmailService{host: host, port: port, user: user, pass: pass, consoleMode: true}
	}
	return &EmailService{host: host, port: port, user: user, pass: pass}
}

// SendEmail sends an HTML email to the given recipient via SMTP with TLS, or logs it in console mode.
func (es *EmailService) SendEmail(to, subject, htmlBody string) error {
	if es.consoleMode {
		log.Printf("[EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, htmlBody)
		return nil
	}

	portNum, err := strconv.Atoi(es.port)
	if err != nil {
		return fmt.Errorf("invalid SMTP port %q: %w", es.port, err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", es.user)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(es.host, portNum, es.user, es.pass)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}
	return nil
}

// RenderTemplate performs simple placeholder replacement on a template string.
// Supported placeholders: {{username}}, {{rank}}, {{reward_type}}, {{reward_value}}, {{reward_image}}.
// If reward_image is non-empty, it is wrapped in an <img> tag; otherwise it renders as empty string.
func (es *EmailService) RenderTemplate(template string, data map[string]string) string {
	s := template

	for _, key := range []string{"username", "rank", "reward_type", "reward_value"} {
		if val, ok := data[key]; ok {
			s = strings.ReplaceAll(s, "{{"+key+"}}", val)
		}
	}

	imageHTML := ""
	if imgURL, ok := data["reward_image"]; ok && imgURL != "" {
		imageHTML = fmt.Sprintf(`<img src="%s" width="100">`, html.EscapeString(imgURL))
	}
	s = strings.ReplaceAll(s, "{{reward_image}}", imageHTML)

	// Convert newlines to <br> for HTML rendering
	s = strings.ReplaceAll(s, "\n", "<br>\n")

	return s
}

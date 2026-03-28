package service

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailService_SMTPMode(t *testing.T) {
	es := NewEmailService("smtp.gmail.com", "587", "user@gmail.com", "pass")
	assert.Equal(t, "smtp.gmail.com", es.host)
	assert.Equal(t, "587", es.port)
	assert.Equal(t, "user@gmail.com", es.user)
	assert.Equal(t, "pass", es.pass)
}

func TestNewEmailService_ConsoleMode_WhenHostEmpty(t *testing.T) {
	es := NewEmailService("", "587", "", "")
	assert.True(t, es.consoleMode)
}

func TestNewEmailService_ConsoleMode_WhenHostIsConsole(t *testing.T) {
	es := NewEmailService("console", "587", "", "")
	assert.True(t, es.consoleMode)
}

func TestSendEmail_ConsoleMode_LogsInsteadOfSending(t *testing.T) {
	es := NewEmailService("", "587", "", "")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	err := es.SendEmail("user@example.com", "Test Subject", "<p>Hello</p>")
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "user@example.com")
	assert.Contains(t, output, "Test Subject")
	assert.Contains(t, output, "<p>Hello</p>")
}

func TestSendEmail_SMTPMode_ReturnsErrorOnBadConnection(t *testing.T) {
	es := NewEmailService("localhost", "19999", "test@test.com", "pass")

	err := es.SendEmail("user@example.com", "Test", "<p>Hello</p>")
	assert.Error(t, err)
}

func TestRenderTemplate_AllPlaceholders(t *testing.T) {
	es := NewEmailService("", "587", "", "")

	result := es.RenderTemplate("Hi {{username}}, rank #{{rank}} wins {{reward_type}} ({{reward_value}}) {{reward_image}}", map[string]string{
		"username":     "james",
		"rank":         "1",
		"reward_type":  "$50 Gift Card",
		"reward_value": "50",
		"reward_image": "https://img.png/gc.jpg",
	})

	assert.Contains(t, result, "Hi james")
	assert.Contains(t, result, "rank #1")
	assert.Contains(t, result, "$50 Gift Card")
	assert.Contains(t, result, "(50)")
	assert.Contains(t, result, `<img src="https://img.png/gc.jpg" width="100">`)
}

func TestRenderTemplate_EmptyImage_ProducesEmpty(t *testing.T) {
	es := NewEmailService("", "587", "", "")

	result := es.RenderTemplate("{{reward_image}}", map[string]string{
		"reward_image": "",
	})

	assert.Equal(t, "", result)
}

func TestRenderTemplate_MissingKey_LeavesPlaceholder(t *testing.T) {
	es := NewEmailService("", "587", "", "")

	result := es.RenderTemplate("Hi {{username}}", map[string]string{})

	assert.Equal(t, "Hi {{username}}", result)
}

func TestRenderTemplate_ImageWithSpecialChars_IsEscaped(t *testing.T) {
	es := NewEmailService("", "587", "", "")

	result := es.RenderTemplate("{{reward_image}}", map[string]string{
		"reward_image": "https://img.png/gc.jpg?a=1&b=2",
	})

	assert.Contains(t, result, "&amp;")
	assert.Contains(t, result, `width="100"`)
}

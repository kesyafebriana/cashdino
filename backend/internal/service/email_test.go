package service

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailService_ConsoleMode_WhenHostEmpty(t *testing.T) {
	es := NewEmailService("", 587, "", "")
	assert.True(t, es.consoleMode)
}

func TestNewEmailService_ConsoleMode_WhenHostIsConsole(t *testing.T) {
	es := NewEmailService("console", 587, "", "")
	assert.True(t, es.consoleMode)
}

func TestNewEmailService_SMTPMode_WhenHostProvided(t *testing.T) {
	es := NewEmailService("smtp.gmail.com", 587, "user@gmail.com", "pass")
	assert.False(t, es.consoleMode)
}

func TestSendEmail_ConsoleMode_LogsInsteadOfSending(t *testing.T) {
	es := NewEmailService("", 587, "", "")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	err := es.SendEmail("user@example.com", "Test Subject", "<p>Hello</p>")
	assert.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.Contains(output, "user@example.com"), "should log the recipient")
	assert.True(t, strings.Contains(output, "Test Subject"), "should log the subject")
	assert.True(t, strings.Contains(output, "<p>Hello</p>"), "should log the body")
}

func TestSendEmail_SMTPMode_ReturnsErrorOnBadConnection(t *testing.T) {
	// Use a bad host/port that will fail to connect
	es := NewEmailService("localhost", 19999, "test@test.com", "pass")

	err := es.SendEmail("user@example.com", "Test", "<p>Hello</p>")
	assert.Error(t, err)
}

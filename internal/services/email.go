package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/enielson/launchpad/templates/email"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendAuthCode(ctx context.Context, toEmail, code string) error
}

// SMTPEmailService sends emails using SMTP
type SMTPEmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewSMTPEmailService creates a new SMTP email service with Fastmail settings
func NewSMTPEmailService() *SMTPEmailService {
	return &SMTPEmailService{
		smtpHost:     "smtp.fastmail.com",
		smtpPort:     "587", // TLS port
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("SMTP_FROM_EMAIL"),
		fromName:     "Launchpad",
	}
}

// SendAuthCode sends an email verification code via SMTP
func (s *SMTPEmailService) SendAuthCode(ctx context.Context, toEmail, code string) error {
	// Generate the HTML email using templ
	var htmlBody bytes.Buffer
	if err := email.AuthCodeEmail(code).Render(ctx, &htmlBody); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Create email headers and body
	subject := "Email Verification Code"
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)

	// Build MIME message with HTML
	message := []byte(
		"From: " + from + "\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			htmlBody.String(),
	)

	// Set up authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := s.smtpHost + ":" + s.smtpPort
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %s", toEmail)
	return nil
}

// MockEmailService is a mock implementation of EmailService for development
type MockEmailService struct{}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

// SendAuthCode logs the email instead of actually sending it
func (s *MockEmailService) SendAuthCode(ctx context.Context, toEmail, code string) error {
	// Generate the HTML email using templ
	var buf bytes.Buffer
	if err := email.AuthCodeEmail(code).Render(ctx, &buf); err != nil {
		return err
	}

	// In development, just log the email details
	log.Printf("=== EMAIL SEND (MOCK) ===")
	log.Printf("To: %s", toEmail)
	log.Printf("Subject: Email Verification Code")
	log.Printf("Code: %s", code)
	log.Printf("HTML Body Length: %d bytes", buf.Len())
	log.Printf("=========================")

	return nil
}

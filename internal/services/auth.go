package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	ErrInvalidCode = fmt.Errorf("invalid verification code")
	ErrCodeExpired = fmt.Errorf("verification code expired")
)

// AuthService handles email authentication logic
type AuthService struct {
	emailService EmailService
	codes        sync.Map // map[email]codeData
}

type codeData struct {
	code      string
	expiresAt time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(emailService EmailService) *AuthService {
	return &AuthService{
		emailService: emailService,
	}
}

// SendEmailCode generates and sends a verification code to the given email
func (s *AuthService) SendEmailCode(ctx context.Context, email string) (string, error) {
	// Generate 6-digit code
	code, err := s.generateCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	// Store code with 10 minute expiration
	s.codes.Store(email, codeData{
		code:      code,
		expiresAt: time.Now().Add(10 * time.Minute),
	})

	// TODO remove later, disable emails for the time being
	return code, nil

	// Send email
	if err := s.emailService.SendAuthCode(ctx, email, code); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return code, nil
}

// VerifyCode checks if the provided code matches the stored code for the email
func (s *AuthService) VerifyCode(email, code string) error {
	value, ok := s.codes.Load(email)
	if !ok {
		return ErrInvalidCode
	}

	data := value.(codeData)

	// Check expiration
	if time.Now().After(data.expiresAt) {
		s.codes.Delete(email)
		return ErrCodeExpired
	}

	// Check code match
	if data.code != code {
		return ErrInvalidCode
	}

	// Code is valid, delete it (one-time use)
	s.codes.Delete(email)

	return nil
}

// generateCode creates a random 6-digit numeric code
func (s *AuthService) generateCode() (string, error) {
	// Generate a random number between 100000 and 999999
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}

	code := n.Int64() + 100000
	return fmt.Sprintf("%06d", code), nil
}

// CleanupExpiredCodes removes expired codes from memory (can be called periodically)
func (s *AuthService) CleanupExpiredCodes() {
	now := time.Now()
	s.codes.Range(func(key, value interface{}) bool {
		data := value.(codeData)
		if now.After(data.expiresAt) {
			s.codes.Delete(key)
		}
		return true
	})
}

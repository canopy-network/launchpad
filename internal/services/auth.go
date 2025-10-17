package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
)

var (
	ErrInvalidCode    = fmt.Errorf("invalid verification code")
	ErrCodeExpired    = fmt.Errorf("verification code expired")
	ErrInvalidToken   = fmt.Errorf("invalid session token")
	ErrTokenExpired   = fmt.Errorf("session token expired")
	ErrTokenRevoked   = fmt.Errorf("session token revoked")
	ErrSessionInvalid = fmt.Errorf("session no longer valid")
)

// AuthService handles email authentication and session management
type AuthService struct {
	emailService EmailService
	userRepo     interfaces.UserRepository
	sessionRepo  interfaces.SessionTokenRepository
	codes        sync.Map // map[email]codeData
}

type codeData struct {
	code      string
	expiresAt time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(emailService EmailService, userRepo interfaces.UserRepository, sessionRepo interfaces.SessionTokenRepository) *AuthService {
	return &AuthService{
		emailService: emailService,
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
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

	// Send email - currently disabled
	// if err := s.emailService.SendAuthCode(ctx, email, code); err != nil {
	// 	return "", fmt.Errorf("failed to send email: %w", err)
	// }
	//
	// return code, nil
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

// Session Token Management

// CreateSession creates a new session token for a user
func (s *AuthService) CreateSession(ctx context.Context, userID uuid.UUID, userAgent, ipAddress string) (string, *models.User, error) {
	// Get user to retrieve JWT version
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate random 32-byte token
	token, err := s.generateSecureToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token for storage
	tokenHash := s.hashToken(token)

	// Get first 8 characters for prefix
	tokenPrefix := token[:8]

	// Sanitize user agent to prevent null byte errors
	sanitizedUserAgent := sanitizeString(userAgent)

	// Store IP address as string (already sanitized in CompleteEmailLogin)
	var ip *string
	if ipAddress != "" {
		ip = &ipAddress
	}

	// Create session token model
	sessionToken := &models.SessionToken{
		UserID:             userID,
		TokenHash:          tokenHash,
		TokenPrefix:        tokenPrefix,
		UserAgent:          &sanitizedUserAgent,
		IPAddress:          ip,
		ExpiresAt:          time.Now().Add(30 * 24 * time.Hour), // 30 days
		LastUsedAt:         time.Now(),
		IsRevoked:          false,
		JWTVersionSnapshot: user.JWTVersion,
	}

	// Save to database
	_, err = s.sessionRepo.Create(ctx, sessionToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session: %w", err)
	}

	return token, user, nil
}

// ValidateToken validates a session token and returns the associated user
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, *models.SessionToken, error) {
	// Hash the token
	tokenHash := s.hashToken(token)

	// Get session from database
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	// Check if session is revoked
	if session.IsRevoked {
		return nil, nil, ErrTokenRevoked
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, nil, ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if JWT version matches (global invalidation check)
	if session.JWTVersionSnapshot != user.JWTVersion {
		return nil, nil, ErrSessionInvalid
	}

	// Update last used timestamp (fire and forget, don't block on this)
	go func() {
		bgCtx := context.Background()
		_ = s.sessionRepo.UpdateLastUsed(bgCtx, session.ID)
	}()

	return user, session, nil
}

// RevokeSession revokes a specific session token
func (s *AuthService) RevokeSession(ctx context.Context, token string) error {
	tokenHash := s.hashToken(token)
	return s.sessionRepo.RevokeToken(ctx, tokenHash, models.RevocationReasonUserLogout)
}

// RevokeSessionByID revokes a session by its ID
func (s *AuthService) RevokeSessionByID(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeTokenByID(ctx, sessionID, models.RevocationReasonUserLogout)
}

// RevokeAllUserSessions revokes all active sessions for a user
func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	return s.sessionRepo.RevokeAllUserTokens(ctx, userID, reason)
}

// GetActiveSessions returns all active sessions for a user
func (s *AuthService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]models.SessionToken, error) {
	return s.sessionRepo.GetActiveSessionsByUserID(ctx, userID)
}

// CompleteEmailLogin completes the email login flow after successful verification
func (s *AuthService) CompleteEmailLogin(ctx context.Context, email, userAgent, ipAddress string) (*models.LoginResponse, error) {
	// Sanitize all inputs to prevent null byte errors
	email = sanitizeString(email)
	userAgent = sanitizeString(userAgent)
	ipAddress = sanitizeString(ipAddress)

	// Create or get user
	user, isNew, err := s.userRepo.CreateOrGetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create user: %w", err)
	}

	// Mark email as verified if it's not already
	if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to mark email as verified: %w", err)
	}

	// Create session token
	token, user, err := s.CreateSession(ctx, user.ID, userAgent, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log for debugging
	if isNew {
		fmt.Printf("New user created: %s (ID: %s)\n", email, user.ID)
	}

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// Helper functions

// generateSecureToken generates a cryptographically secure random token
func (s *AuthService) generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken hashes a token using SHA-256
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// sanitizeString removes null bytes (0x00) from strings to prevent PostgreSQL UTF8 encoding errors
func sanitizeString(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

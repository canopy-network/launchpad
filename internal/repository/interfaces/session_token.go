package interfaces

import (
	"context"

	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
)

// SessionTokenRepository defines the interface for session token data operations
type SessionTokenRepository interface {
	// Create creates a new session token
	Create(ctx context.Context, token *models.SessionToken) (*models.SessionToken, error)

	// GetByID retrieves a session token by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.SessionToken, error)

	// GetByTokenHash retrieves a session token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.SessionToken, error)

	// GetActiveSessionsByUserID retrieves all active (non-revoked, non-expired) sessions for a user
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.SessionToken, error)

	// GetAllSessionsByUserID retrieves all sessions for a user (including revoked/expired)
	GetAllSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.SessionToken, error)

	// UpdateLastUsed updates the last_used_at timestamp for a session
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error

	// RevokeToken revokes a specific session token
	RevokeToken(ctx context.Context, tokenHash string, reason string) error

	// RevokeAllUserTokens revokes all active tokens for a user (e.g., on password change, security event)
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error

	// RevokeTokenByID revokes a session token by its ID
	RevokeTokenByID(ctx context.Context, id uuid.UUID, reason string) error

	// DeleteExpiredTokens removes expired tokens older than the retention period
	DeleteExpiredTokens(ctx context.Context, retentionDays int) (int64, error)

	// Delete permanently deletes a session token
	Delete(ctx context.Context, id uuid.UUID) error
}

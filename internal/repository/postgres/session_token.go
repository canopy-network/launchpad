package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type sessionTokenRepository struct {
	db *sqlx.DB
}

// NewSessionTokenRepository creates a new PostgreSQL session token repository
func NewSessionTokenRepository(db *sqlx.DB) interfaces.SessionTokenRepository {
	return &sessionTokenRepository{db: db}
}

// Create creates a new session token
func (r *sessionTokenRepository) Create(ctx context.Context, token *models.SessionToken) (*models.SessionToken, error) {
	query := `
		INSERT INTO session_tokens (
			user_id, token_hash, token_prefix, user_agent, ip_address, device_name,
			expires_at, last_used_at, is_revoked, jwt_version_snapshot
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		token.UserID,
		token.TokenHash,
		token.TokenPrefix,
		database.NullString(token.UserAgent),
		token.IPAddress,
		database.NullString(token.DeviceName),
		token.ExpiresAt,
		token.LastUsedAt,
		token.IsRevoked,
		token.JWTVersionSnapshot,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create session token: %w", err)
	}

	return token, nil
}

// GetByID retrieves a session token by ID
func (r *sessionTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SessionToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_prefix, user_agent, ip_address, device_name,
			   expires_at, last_used_at, is_revoked, revoked_at, revocation_reason,
			   jwt_version_snapshot, created_at, updated_at
		FROM session_tokens
		WHERE id = $1`

	return r.getSessionToken(ctx, query, id)
}

// GetByTokenHash retrieves a session token by its hash
func (r *sessionTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.SessionToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_prefix, user_agent, ip_address, device_name,
			   expires_at, last_used_at, is_revoked, revoked_at, revocation_reason,
			   jwt_version_snapshot, created_at, updated_at
		FROM session_tokens
		WHERE token_hash = $1`

	return r.getSessionToken(ctx, query, tokenHash)
}

// GetActiveSessionsByUserID retrieves all active sessions for a user
func (r *sessionTokenRepository) GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.SessionToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_prefix, user_agent, ip_address, device_name,
			   expires_at, last_used_at, is_revoked, revoked_at, revocation_reason,
			   jwt_version_snapshot, created_at, updated_at
		FROM session_tokens
		WHERE user_id = $1
		  AND is_revoked = FALSE
		  AND expires_at > NOW()
		ORDER BY last_used_at DESC`

	return r.getSessionTokens(ctx, query, userID)
}

// GetAllSessionsByUserID retrieves all sessions for a user
func (r *sessionTokenRepository) GetAllSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.SessionToken, error) {
	query := `
		SELECT id, user_id, token_hash, token_prefix, user_agent, ip_address, device_name,
			   expires_at, last_used_at, is_revoked, revoked_at, revocation_reason,
			   jwt_version_snapshot, created_at, updated_at
		FROM session_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`

	return r.getSessionTokens(ctx, query, userID)
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *sessionTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE session_tokens
		SET last_used_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session token not found")
	}

	return nil
}

// RevokeToken revokes a specific session token
func (r *sessionTokenRepository) RevokeToken(ctx context.Context, tokenHash string, reason string) error {
	query := `
		UPDATE session_tokens
		SET is_revoked = TRUE,
		    revoked_at = NOW(),
		    revocation_reason = $2,
		    updated_at = NOW()
		WHERE token_hash = $1
		  AND is_revoked = FALSE`

	result, err := r.db.ExecContext(ctx, query, tokenHash, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session token not found or already revoked")
	}

	return nil
}

// RevokeTokenByID revokes a session token by its ID
func (r *sessionTokenRepository) RevokeTokenByID(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE session_tokens
		SET is_revoked = TRUE,
		    revoked_at = NOW(),
		    revocation_reason = $2,
		    updated_at = NOW()
		WHERE id = $1
		  AND is_revoked = FALSE`

	result, err := r.db.ExecContext(ctx, query, id, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session token not found or already revoked")
	}

	return nil
}

// RevokeAllUserTokens revokes all active tokens for a user
func (r *sessionTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `
		UPDATE session_tokens
		SET is_revoked = TRUE,
		    revoked_at = NOW(),
		    revocation_reason = $2,
		    updated_at = NOW()
		WHERE user_id = $1
		  AND is_revoked = FALSE`

	_, err := r.db.ExecContext(ctx, query, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

// DeleteExpiredTokens removes expired tokens older than the retention period
func (r *sessionTokenRepository) DeleteExpiredTokens(ctx context.Context, retentionDays int) (int64, error) {
	query := `
		DELETE FROM session_tokens
		WHERE expires_at < NOW() - INTERVAL '1 day' * $1`

	result, err := r.db.ExecContext(ctx, query, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// Delete permanently deletes a session token
func (r *sessionTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM session_tokens WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session token not found")
	}

	return nil
}

// Helper methods

func (r *sessionTokenRepository) getSessionToken(ctx context.Context, query string, arg interface{}) (*models.SessionToken, error) {
	var token models.SessionToken
	var userAgent, deviceName sql.NullString
	var revoked_at sql.NullTime
	var revocation_reason sql.NullString

	err := r.db.QueryRowxContext(ctx, query, arg).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.TokenPrefix,
		&userAgent,
		&token.IPAddress,
		&deviceName,
		&token.ExpiresAt,
		&token.LastUsedAt,
		&token.IsRevoked,
		&revoked_at,
		&revocation_reason,
		&token.JWTVersionSnapshot,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session token not found")
		}
		return nil, fmt.Errorf("failed to get session token: %w", err)
	}

	// Handle nullable fields
	token.UserAgent = database.StringPtr(userAgent)
	token.DeviceName = database.StringPtr(deviceName)
	token.RevocationReason = database.StringPtr(revocation_reason)
	if revoked_at.Valid {
		token.RevokedAt = &revoked_at.Time
	}

	return &token, nil
}

func (r *sessionTokenRepository) getSessionTokens(ctx context.Context, query string, arg interface{}) ([]models.SessionToken, error) {
	rows, err := r.db.QueryxContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query session tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.SessionToken
	for rows.Next() {
		var token models.SessionToken
		var userAgent, deviceName sql.NullString
		var revokedAt sql.NullTime
		var revocationReason sql.NullString

		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.TokenPrefix,
			&userAgent,
			&token.IPAddress,
			&deviceName,
			&token.ExpiresAt,
			&token.LastUsedAt,
			&token.IsRevoked,
			&revokedAt,
			&revocationReason,
			&token.JWTVersionSnapshot,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session token: %w", err)
		}

		// Handle nullable fields
		token.UserAgent = database.StringPtr(userAgent)
		token.DeviceName = database.StringPtr(deviceName)
		token.RevocationReason = database.StringPtr(revocationReason)
		if revokedAt.Valid {
			token.RevokedAt = &revokedAt.Time
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session tokens: %w", err)
	}

	return tokens, nil
}

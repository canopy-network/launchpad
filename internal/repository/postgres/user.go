package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sqlx.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (
			wallet_address, email, username, display_name, bio, avatar_url, website_url,
			twitter_handle, github_username, telegram_handle,
			is_verified, verification_tier, total_chains_created, total_cnpy_invested,
			reputation_score, last_active_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at, updated_at`

	// Sanitize all text fields to prevent null byte errors
	err := r.db.QueryRowxContext(ctx, query,
		sanitizeString(user.WalletAddress),
		database.NullString(sanitizeStringPtr(user.Email)),
		database.NullString(sanitizeStringPtr(user.Username)),
		database.NullString(sanitizeStringPtr(user.DisplayName)),
		database.NullString(sanitizeStringPtr(user.Bio)),
		database.NullString(sanitizeStringPtr(user.AvatarURL)),
		database.NullString(sanitizeStringPtr(user.WebsiteURL)),
		database.NullString(sanitizeStringPtr(user.TwitterHandle)),
		database.NullString(sanitizeStringPtr(user.GithubUsername)),
		database.NullString(sanitizeStringPtr(user.TelegramHandle)),
		user.IsVerified,
		user.VerificationTier,
		user.TotalChainsCreated,
		user.TotalCNPYInvested,
		user.ReputationScore,
		user.LastActiveAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return r.getUserByField(ctx, "id", id.String())
}

// GetByWalletAddress retrieves a user by wallet address
func (r *userRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*models.User, error) {
	return r.getUserByField(ctx, "wallet_address", walletAddress)
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.getUserByField(ctx, "email", email)
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.getUserByField(ctx, "username", username)
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		UPDATE users SET
			wallet_address = $2, email = $3, username = $4, display_name = $5, bio = $6,
			avatar_url = $7, website_url = $8, twitter_handle = $9, github_username = $10,
			telegram_handle = $11, is_verified = $12,
			verification_tier = $13, total_chains_created = $14, total_cnpy_invested = $15,
			reputation_score = $16, last_active_at = $17, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.ID,
		user.WalletAddress,
		database.NullString(user.Email),
		database.NullString(user.Username),
		database.NullString(user.DisplayName),
		database.NullString(user.Bio),
		database.NullString(user.AvatarURL),
		database.NullString(user.WebsiteURL),
		database.NullString(user.TwitterHandle),
		database.NullString(user.GithubUsername),
		database.NullString(user.TelegramHandle),
		user.IsVerified,
		user.VerificationTier,
		user.TotalChainsCreated,
		user.TotalCNPYInvested,
		user.ReputationScore,
		user.LastActiveAt,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves users with filtering and pagination
func (r *userRepository) List(ctx context.Context, filters interfaces.UserFilters, pagination interfaces.Pagination) ([]models.User, int, error) {
	whereClause, args := r.buildUserWhereClause(filters)

	// Count query
	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT * FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, pagination.Limit, pagination.Offset)

	users := []models.User{}
	err = r.db.SelectContext(ctx, &users, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}

	return users, total, nil
}

// ListByVerificationTier retrieves users by verification tier
func (r *userRepository) ListByVerificationTier(ctx context.Context, tier string, pagination interfaces.Pagination) ([]models.User, int, error) {
	filters := interfaces.UserFilters{VerificationTier: &tier}
	return r.List(ctx, filters, pagination)
}

// GetPositionsByUserID retrieves user positions
func (r *userRepository) GetPositionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserVirtualPosition, error) {
	// This would be implemented when user_virtual_positions functionality is needed
	return nil, fmt.Errorf("not implemented")
}

// GetPositionByUserAndChain retrieves user position for specific chain
func (r *userRepository) GetPositionByUserAndChain(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	// This would be implemented when user_virtual_positions functionality is needed
	return nil, fmt.Errorf("not implemented")
}

// UpdatePosition updates user position
func (r *userRepository) UpdatePosition(ctx context.Context, position *models.UserVirtualPosition) (*models.UserVirtualPosition, error) {
	// This would be implemented when user_virtual_positions functionality is needed
	return nil, fmt.Errorf("not implemented")
}

// UpdateChainsCreatedCount updates the count of chains created by user
func (r *userRepository) UpdateChainsCreatedCount(ctx context.Context, userID uuid.UUID, increment int) error {
	query := `
		UPDATE users
		SET total_chains_created = total_chains_created + $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, increment)
	if err != nil {
		return fmt.Errorf("failed to update chains created count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateCNPYInvestment updates total CNPY invested by user
func (r *userRepository) UpdateCNPYInvestment(ctx context.Context, userID uuid.UUID, amount float64) error {
	query := `
		UPDATE users
		SET total_cnpy_invested = total_cnpy_invested + $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to update CNPY investment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateReputationScore updates user reputation score
func (r *userRepository) UpdateReputationScore(ctx context.Context, userID uuid.UUID, score int) error {
	query := `
		UPDATE users
		SET reputation_score = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, score)
	if err != nil {
		return fmt.Errorf("failed to update reputation score: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastActive updates user last active timestamp
func (r *userRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_active_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// CreateOrGetByEmail creates a new user if they don't exist, or returns existing user
func (r *userRepository) CreateOrGetByEmail(ctx context.Context, email string) (*models.User, bool, error) {
	// Sanitize email to remove null bytes (PostgreSQL UTF8 doesn't support them)
	sanitizedEmail := sanitizeString(email)

	// First try to get existing user
	existingUser, err := r.GetByEmail(ctx, sanitizedEmail)
	if err == nil {
		// User exists
		return existingUser, false, nil
	}

	// User doesn't exist, create a new one
	// Generate a wallet address placeholder (user can update later)
	walletAddress := fmt.Sprintf("email_%s", sanitizedEmail)

	user := &models.User{
		WalletAddress:    walletAddress,
		Email:            &sanitizedEmail,
		IsVerified:       false,
		VerificationTier: models.VerificationTierBasic,
		JWTVersion:       0,
	}

	createdUser, err := r.Create(ctx, user)
	if err != nil {
		// Check if there was a race condition and user was created by another request
		existingUser, getErr := r.GetByEmail(ctx, sanitizedEmail)
		if getErr == nil {
			return existingUser, false, nil
		}
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, true, nil
}

// MarkEmailVerified marks a user's email as verified
func (r *userRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified_at = CURRENT_TIMESTAMP,
		    is_verified = TRUE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND email_verified_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// User might already be verified, not an error
		return nil
	}

	return nil
}

// IncrementJWTVersion increments the JWT version to invalidate all existing tokens
func (r *userRepository) IncrementJWTVersion(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET jwt_version = jwt_version + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to increment JWT version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateProfile updates user profile fields (partial update)
func (r *userRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	// Build dynamic query based on provided fields
	setClauses := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}
	argCount := 1

	// Add user ID as first argument
	args = append(args, userID)
	argCount++

	if req.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argCount))
		args = append(args, database.NullString(req.Username))
		argCount++
	}

	if req.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argCount))
		args = append(args, database.NullString(req.DisplayName))
		argCount++
	}

	if req.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argCount))
		args = append(args, database.NullString(req.Bio))
		argCount++
	}

	if req.AvatarURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_url = $%d", argCount))
		args = append(args, database.NullString(req.AvatarURL))
		argCount++
	}

	if req.WebsiteURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("website_url = $%d", argCount))
		args = append(args, database.NullString(req.WebsiteURL))
		argCount++
	}

	if req.TwitterHandle != nil {
		setClauses = append(setClauses, fmt.Sprintf("twitter_handle = $%d", argCount))
		args = append(args, database.NullString(req.TwitterHandle))
		argCount++
	}

	if req.GithubUsername != nil {
		setClauses = append(setClauses, fmt.Sprintf("github_username = $%d", argCount))
		args = append(args, database.NullString(req.GithubUsername))
		argCount++
	}

	if req.TelegramHandle != nil {
		setClauses = append(setClauses, fmt.Sprintf("telegram_handle = $%d", argCount))
		args = append(args, database.NullString(req.TelegramHandle))
		argCount++
	}

	// If no fields to update, just return the current user
	if len(setClauses) == 1 {
		return r.GetByID(ctx, userID)
	}

	// Build and execute update query
	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $1
		RETURNING id, wallet_address, email, username, display_name, bio, avatar_url, website_url,
		          twitter_handle, github_username, telegram_handle,
		          is_verified, verification_tier, email_verified_at, jwt_version,
		          total_chains_created, total_cnpy_invested, reputation_score,
		          created_at, updated_at, last_active_at`,
		fmt.Sprintf("%s", setClauses[0])+func() string {
			result := ""
			for i := 1; i < len(setClauses); i++ {
				result += ", " + setClauses[i]
			}
			return result
		}(),
	)

	var user models.User
	var email, username, displayName, bio, avatarURL, websiteURL sql.NullString
	var twitterHandle, githubUsername, telegramHandle sql.NullString
	var emailVerifiedAt, lastActiveAt sql.NullTime

	err := r.db.QueryRowxContext(ctx, query, args...).Scan(
		&user.ID, &user.WalletAddress, &email, &username, &displayName, &bio,
		&avatarURL, &websiteURL, &twitterHandle, &githubUsername, &telegramHandle,
		&user.IsVerified, &user.VerificationTier, &emailVerifiedAt, &user.JWTVersion,
		&user.TotalChainsCreated, &user.TotalCNPYInvested, &user.ReputationScore,
		&user.CreatedAt, &user.UpdatedAt, &lastActiveAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Unique constraint violation
			if pqErr.Constraint == "users_username_key" {
				return nil, fmt.Errorf("username already taken")
			}
			return nil, fmt.Errorf("duplicate value")
		}
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Handle nullable fields
	user.Email = database.StringPtr(email)
	user.Username = database.StringPtr(username)
	user.DisplayName = database.StringPtr(displayName)
	user.Bio = database.StringPtr(bio)
	user.AvatarURL = database.StringPtr(avatarURL)
	user.WebsiteURL = database.StringPtr(websiteURL)
	user.TwitterHandle = database.StringPtr(twitterHandle)
	user.GithubUsername = database.StringPtr(githubUsername)
	user.TelegramHandle = database.StringPtr(telegramHandle)
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lastActiveAt.Valid {
		user.LastActiveAt = &lastActiveAt.Time
	}

	return &user, nil
}

// Helper methods
func (r *userRepository) getUserByField(ctx context.Context, field, value string) (*models.User, error) {
	query := fmt.Sprintf(`
		SELECT id, wallet_address, email, username, display_name, bio, avatar_url, website_url,
			   twitter_handle, github_username, telegram_handle,
			   is_verified, verification_tier, email_verified_at, jwt_version,
			   total_chains_created, total_cnpy_invested,
			   reputation_score, created_at, updated_at, last_active_at
		FROM users WHERE %s = $1`, field)

	var user models.User
	var email, username, displayName, bio, avatarURL, websiteURL sql.NullString
	var twitterHandle, githubUsername, telegramHandle sql.NullString
	var emailVerifiedAt, lastActiveAt sql.NullTime

	err := r.db.QueryRowxContext(ctx, query, value).Scan(
		&user.ID, &user.WalletAddress, &email, &username, &displayName, &bio,
		&avatarURL, &websiteURL, &twitterHandle, &githubUsername, &telegramHandle,
		&user.IsVerified, &user.VerificationTier, &emailVerifiedAt, &user.JWTVersion,
		&user.TotalChainsCreated, &user.TotalCNPYInvested, &user.ReputationScore,
		&user.CreatedAt, &user.UpdatedAt, &lastActiveAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Handle nullable fields
	user.Email = database.StringPtr(email)
	user.Username = database.StringPtr(username)
	user.DisplayName = database.StringPtr(displayName)
	user.Bio = database.StringPtr(bio)
	user.AvatarURL = database.StringPtr(avatarURL)
	user.WebsiteURL = database.StringPtr(websiteURL)
	user.TwitterHandle = database.StringPtr(twitterHandle)
	user.GithubUsername = database.StringPtr(githubUsername)
	user.TelegramHandle = database.StringPtr(telegramHandle)
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lastActiveAt.Valid {
		user.LastActiveAt = &lastActiveAt.Time
	}

	return &user, nil
}

func (r *userRepository) buildUserWhereClause(filters interfaces.UserFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 0

	if filters.VerificationTier != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("verification_tier = $%d", argCount))
		args = append(args, *filters.VerificationTier)
	}

	if filters.IsVerified != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_verified = $%d", argCount))
		args = append(args, *filters.IsVerified)
	}

	if filters.MinReputation != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("reputation_score >= $%d", argCount))
		args = append(args, *filters.MinReputation)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + fmt.Sprintf("%s", conditions[0])
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	return whereClause, args
}

// sanitizeString removes null bytes (0x00) from strings to prevent PostgreSQL UTF8 encoding errors
func sanitizeString(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

// sanitizeStringPtr removes null bytes from string pointers
func sanitizeStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	cleaned := sanitizeString(*s)
	return &cleaned
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type walletRepository struct {
	db *sqlx.DB
}

// NewWalletRepository creates a new PostgreSQL wallet repository
func NewWalletRepository(db *sqlx.DB) interfaces.WalletRepository {
	return &walletRepository{db: db}
}

// Create creates a new wallet
func (r *walletRepository) Create(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error) {
	query := `
		INSERT INTO wallets (
			user_id, chain_id, address, public_key, encrypted_private_key, salt,
			wallet_name, wallet_description, is_active, is_locked, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		database.NullUUID(wallet.UserID),
		database.NullUUID(wallet.ChainID),
		wallet.Address,
		wallet.PublicKey,
		wallet.EncryptedPrivateKey,
		wallet.Salt,
		database.NullString(wallet.WalletName),
		database.NullString(wallet.WalletDescription),
		wallet.IsActive,
		wallet.IsLocked,
		database.NullUUID(wallet.CreatedBy),
	).Scan(&wallet.ID, &wallet.CreatedAt, &wallet.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("wallet with address already exists")
		}
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

// GetByID retrieves a wallet by ID
func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, user_id, chain_id, address, public_key, encrypted_private_key, salt,
			   wallet_name, wallet_description, is_active, is_locked,
			   last_used_at, password_changed_at, failed_decrypt_attempts, locked_until,
			   created_by, created_at, updated_at
		FROM wallets WHERE id = $1`

	var wallet models.Wallet
	var userID, chainID, createdBy sql.NullString
	var walletName, walletDescription sql.NullString
	var lastUsedAt, passwordChangedAt, lockedUntil sql.NullTime

	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&wallet.ID, &userID, &chainID, &wallet.Address, &wallet.PublicKey,
		&wallet.EncryptedPrivateKey, &wallet.Salt,
		&walletName, &walletDescription, &wallet.IsActive, &wallet.IsLocked,
		&lastUsedAt, &passwordChangedAt, &wallet.FailedDecryptAttempts, &lockedUntil,
		&createdBy, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Handle nullable fields
	wallet.UserID = database.UUIDPtr(userID)
	wallet.ChainID = database.UUIDPtr(chainID)
	wallet.CreatedBy = database.UUIDPtr(createdBy)
	wallet.WalletName = database.StringPtr(walletName)
	wallet.WalletDescription = database.StringPtr(walletDescription)
	if lastUsedAt.Valid {
		wallet.LastUsedAt = &lastUsedAt.Time
	}
	if passwordChangedAt.Valid {
		wallet.PasswordChangedAt = &passwordChangedAt.Time
	}
	if lockedUntil.Valid {
		wallet.LockedUntil = &lockedUntil.Time
	}

	return &wallet, nil
}

// GetByAddress retrieves a wallet by address
func (r *walletRepository) GetByAddress(ctx context.Context, address string) (*models.Wallet, error) {
	query := `
		SELECT id, user_id, chain_id, address, public_key, encrypted_private_key, salt,
			   wallet_name, wallet_description, is_active, is_locked,
			   last_used_at, password_changed_at, failed_decrypt_attempts, locked_until,
			   created_by, created_at, updated_at
		FROM wallets WHERE address = $1`

	var wallet models.Wallet
	var userID, chainID, createdBy sql.NullString
	var walletName, walletDescription sql.NullString
	var lastUsedAt, passwordChangedAt, lockedUntil sql.NullTime

	err := r.db.QueryRowxContext(ctx, query, address).Scan(
		&wallet.ID, &userID, &chainID, &wallet.Address, &wallet.PublicKey,
		&wallet.EncryptedPrivateKey, &wallet.Salt,
		&walletName, &walletDescription, &wallet.IsActive, &wallet.IsLocked,
		&lastUsedAt, &passwordChangedAt, &wallet.FailedDecryptAttempts, &lockedUntil,
		&createdBy, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Handle nullable fields
	wallet.UserID = database.UUIDPtr(userID)
	wallet.ChainID = database.UUIDPtr(chainID)
	wallet.CreatedBy = database.UUIDPtr(createdBy)
	wallet.WalletName = database.StringPtr(walletName)
	wallet.WalletDescription = database.StringPtr(walletDescription)
	if lastUsedAt.Valid {
		wallet.LastUsedAt = &lastUsedAt.Time
	}
	if passwordChangedAt.Valid {
		wallet.PasswordChangedAt = &passwordChangedAt.Time
	}
	if lockedUntil.Valid {
		wallet.LockedUntil = &lockedUntil.Time
	}

	return &wallet, nil
}

// Update updates a wallet
func (r *walletRepository) Update(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error) {
	query := `
		UPDATE wallets SET
			wallet_name = $2, wallet_description = $3, is_active = $4, is_locked = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		wallet.ID,
		database.NullString(wallet.WalletName),
		database.NullString(wallet.WalletDescription),
		wallet.IsActive,
		wallet.IsLocked,
	).Scan(&wallet.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return wallet, nil
}

// Delete deletes a wallet
func (r *walletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM wallets WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// List retrieves wallets with filtering and pagination
func (r *walletRepository) List(ctx context.Context, filters interfaces.WalletFilters, pagination interfaces.Pagination) ([]models.Wallet, int, error) {
	whereClause, args := r.buildWalletWhereClause(filters)

	// Count query
	countQuery := "SELECT COUNT(*) FROM wallets" + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count wallets: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, chain_id, address, public_key, encrypted_private_key, salt,
			   wallet_name, wallet_description, is_active, is_locked,
			   last_used_at, password_changed_at, failed_decrypt_attempts, locked_until,
			   created_by, created_at, updated_at
		FROM wallets
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query wallets: %w", err)
	}
	defer rows.Close()

	wallets := []models.Wallet{}
	for rows.Next() {
		var wallet models.Wallet
		var userID, chainID, createdBy sql.NullString
		var walletName, walletDescription sql.NullString
		var lastUsedAt, passwordChangedAt, lockedUntil sql.NullTime

		err := rows.Scan(
			&wallet.ID, &userID, &chainID, &wallet.Address, &wallet.PublicKey,
			&wallet.EncryptedPrivateKey, &wallet.Salt,
			&walletName, &walletDescription, &wallet.IsActive, &wallet.IsLocked,
			&lastUsedAt, &passwordChangedAt, &wallet.FailedDecryptAttempts, &lockedUntil,
			&createdBy, &wallet.CreatedAt, &wallet.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan wallet: %w", err)
		}

		// Handle nullable fields
		wallet.UserID = database.UUIDPtr(userID)
		wallet.ChainID = database.UUIDPtr(chainID)
		wallet.CreatedBy = database.UUIDPtr(createdBy)
		wallet.WalletName = database.StringPtr(walletName)
		wallet.WalletDescription = database.StringPtr(walletDescription)
		if lastUsedAt.Valid {
			wallet.LastUsedAt = &lastUsedAt.Time
		}
		if passwordChangedAt.Valid {
			wallet.PasswordChangedAt = &passwordChangedAt.Time
		}
		if lockedUntil.Valid {
			wallet.LockedUntil = &lockedUntil.Time
		}

		wallets = append(wallets, wallet)
	}

	return wallets, total, nil
}

// ListByUserID retrieves wallets for a specific user
func (r *walletRepository) ListByUserID(ctx context.Context, userID uuid.UUID, pagination interfaces.Pagination) ([]models.Wallet, int, error) {
	filters := interfaces.WalletFilters{UserID: &userID}
	return r.List(ctx, filters, pagination)
}

// ListByChainID retrieves wallets for a specific chain
func (r *walletRepository) ListByChainID(ctx context.Context, chainID uuid.UUID, pagination interfaces.Pagination) ([]models.Wallet, int, error) {
	filters := interfaces.WalletFilters{ChainID: &chainID}
	return r.List(ctx, filters, pagination)
}

// IncrementFailedAttempts increments the failed decrypt attempts counter
func (r *walletRepository) IncrementFailedAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE wallets
		SET failed_decrypt_attempts = failed_decrypt_attempts + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment failed attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// ResetFailedAttempts resets the failed decrypt attempts counter
func (r *walletRepository) ResetFailedAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE wallets
		SET failed_decrypt_attempts = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// LockWallet locks a wallet for a specified duration
func (r *walletRepository) LockWallet(ctx context.Context, id uuid.UUID, lockDuration int) error {
	lockedUntil := time.Now().Add(time.Duration(lockDuration) * time.Minute)

	query := `
		UPDATE wallets
		SET is_locked = true, locked_until = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, lockedUntil)
	if err != nil {
		return fmt.Errorf("failed to lock wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// UnlockWallet unlocks a wallet
func (r *walletRepository) UnlockWallet(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE wallets
		SET is_locked = false, locked_until = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// UpdateLastUsed updates the last used timestamp
func (r *walletRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE wallets
		SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
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
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// Helper methods
func (r *walletRepository) buildWalletWhereClause(filters interfaces.WalletFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 0

	if filters.UserID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, *filters.UserID)
	}

	if filters.ChainID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("chain_id = $%d", argCount))
		args = append(args, *filters.ChainID)
	}

	if filters.IsActive != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *filters.IsActive)
	}

	if filters.IsLocked != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_locked = $%d", argCount))
		args = append(args, *filters.IsLocked)
	}

	if filters.CreatedBy != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argCount))
		args = append(args, *filters.CreatedBy)
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

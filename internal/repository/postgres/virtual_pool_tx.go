package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// VirtualPoolTxRepository extends VirtualPoolRepository with transaction-aware methods
// that support row-level locking and atomic updates within database transactions.
//
// All *InTx methods accept a *sqlx.Tx parameter and should be called within a transaction.
// The *ForUpdate methods use SELECT ... FOR UPDATE to acquire exclusive row locks,
// preventing concurrent modifications and ensuring consistency.
type VirtualPoolTxRepository interface {
	interfaces.VirtualPoolRepository

	// GetPoolByChainIDForUpdate retrieves a virtual pool with an exclusive row lock.
	// Uses SELECT ... FOR UPDATE to prevent concurrent modifications.
	// Other transactions attempting to read this row will wait until the lock is released.
	GetPoolByChainIDForUpdate(ctx context.Context, tx *sqlx.Tx, chainID uuid.UUID) (*models.VirtualPool, error)

	// UpdatePoolStateInTx updates virtual pool state within a transaction.
	// Should be called after acquiring lock via GetPoolByChainIDForUpdate.
	UpdatePoolStateInTx(ctx context.Context, tx *sqlx.Tx, chainID uuid.UUID, update *interfaces.PoolStateUpdate) error

	// CreateTransactionInTx creates a transaction record within a database transaction.
	// Records the order details and resulting pool state for audit trail.
	CreateTransactionInTx(ctx context.Context, tx *sqlx.Tx, transaction *models.VirtualPoolTransaction) error

	// GetUserPositionForUpdate retrieves a user position with an exclusive row lock.
	// Returns nil if position doesn't exist (not an error - user hasn't traded yet).
	GetUserPositionForUpdate(ctx context.Context, tx *sqlx.Tx, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error)

	// UpsertUserPositionInTx inserts or updates a user position within a transaction.
	// Uses ON CONFLICT to handle both new and existing positions atomically.
	UpsertUserPositionInTx(ctx context.Context, tx *sqlx.Tx, position *models.UserVirtualPosition) error
}

// virtualPoolTxRepository implements transaction-aware virtual pool operations
type virtualPoolTxRepository struct {
	*virtualPoolRepository
}

// NewVirtualPoolTxRepository creates a new transaction-aware virtual pool repository
func NewVirtualPoolTxRepository(db *sqlx.DB) VirtualPoolTxRepository {
	return &virtualPoolTxRepository{
		virtualPoolRepository: &virtualPoolRepository{db: db},
	}
}

// GetPoolByChainIDForUpdate retrieves a virtual pool with FOR UPDATE lock
func (r *virtualPoolTxRepository) GetPoolByChainIDForUpdate(ctx context.Context, tx *sqlx.Tx, chainID uuid.UUID) (*models.VirtualPool, error) {
	query := `
		SELECT id, chain_id, cnpy_reserve, token_reserve, current_price_cnpy, market_cap_usd,
			   total_volume_cnpy, total_transactions, unique_traders, is_active,
			   price_24h_change_percent, volume_24h_cnpy, high_24h_cnpy, low_24h_cnpy,
			   created_at, updated_at
		FROM virtual_pools
		WHERE chain_id = $1
		FOR UPDATE`

	var pool models.VirtualPool
	err := tx.QueryRowxContext(ctx, query, chainID).Scan(
		&pool.ID, &pool.ChainID, &pool.CNPYReserve, &pool.TokenReserve,
		&pool.CurrentPriceCNPY, &pool.MarketCapUSD, &pool.TotalVolumeCNPY,
		&pool.TotalTransactions, &pool.UniqueTraders, &pool.IsActive,
		&pool.Price24hChangePercent, &pool.Volume24hCNPY, &pool.High24hCNPY,
		&pool.Low24hCNPY, &pool.CreatedAt, &pool.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("virtual pool not found for chain_id: %s", chainID)
		}
		return nil, fmt.Errorf("failed to get virtual pool with lock: %w", err)
	}

	return &pool, nil
}

// UpdatePoolStateInTx updates pool state within a transaction
func (r *virtualPoolTxRepository) UpdatePoolStateInTx(ctx context.Context, tx *sqlx.Tx, chainID uuid.UUID, update *interfaces.PoolStateUpdate) error {
	// Build dynamic update query based on non-nil fields
	query := "UPDATE virtual_pools SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{chainID}
	argCount := 1

	if update.CNPYReserve != nil {
		argCount++
		cnpyFloat, _ := update.CNPYReserve.Float64()
		query += fmt.Sprintf(", cnpy_reserve = $%d", argCount)
		args = append(args, cnpyFloat)
	}

	if update.TokenReserve != nil {
		argCount++
		tokenFloat, _ := update.TokenReserve.Float64()
		query += fmt.Sprintf(", token_reserve = $%d", argCount)
		args = append(args, int64(tokenFloat))
	}

	if update.CurrentPriceCNPY != nil {
		argCount++
		priceFloat, _ := update.CurrentPriceCNPY.Float64()
		query += fmt.Sprintf(", current_price_cnpy = $%d", argCount)
		args = append(args, priceFloat)
	}

	if update.MarketCapUSD != nil {
		argCount++
		marketCapFloat, _ := update.MarketCapUSD.Float64()
		query += fmt.Sprintf(", market_cap_usd = $%d", argCount)
		args = append(args, marketCapFloat)
	}

	if update.TotalVolumeCNPY != nil {
		argCount++
		volumeFloat, _ := update.TotalVolumeCNPY.Float64()
		query += fmt.Sprintf(", total_volume_cnpy = $%d", argCount)
		args = append(args, volumeFloat)
	}

	if update.TotalTransactions != nil {
		argCount++
		query += fmt.Sprintf(", total_transactions = $%d", argCount)
		args = append(args, *update.TotalTransactions)
	}

	if update.UniqueTraders != nil {
		argCount++
		query += fmt.Sprintf(", unique_traders = $%d", argCount)
		args = append(args, *update.UniqueTraders)
	}

	if update.Volume24hCNPY != nil {
		argCount++
		volume24hFloat, _ := update.Volume24hCNPY.Float64()
		query += fmt.Sprintf(", volume_24h_cnpy = $%d", argCount)
		args = append(args, volume24hFloat)
	}

	if update.High24hCNPY != nil {
		argCount++
		high24hFloat, _ := update.High24hCNPY.Float64()
		query += fmt.Sprintf(", high_24h_cnpy = $%d", argCount)
		args = append(args, high24hFloat)
	}

	if update.Low24hCNPY != nil {
		argCount++
		low24hFloat, _ := update.Low24hCNPY.Float64()
		query += fmt.Sprintf(", low_24h_cnpy = $%d", argCount)
		args = append(args, low24hFloat)
	}

	if update.Price24hChangePerc != nil {
		argCount++
		price24hChangeFloat, _ := update.Price24hChangePerc.Float64()
		query += fmt.Sprintf(", price_24h_change_percent = $%d", argCount)
		args = append(args, price24hChangeFloat)
	}

	query += " WHERE chain_id = $1"

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update pool state in tx: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("virtual pool not found for chain_id: %s", chainID)
	}

	return nil
}

// CreateTransactionInTx creates a transaction record within a database transaction
func (r *virtualPoolTxRepository) CreateTransactionInTx(ctx context.Context, tx *sqlx.Tx, transaction *models.VirtualPoolTransaction) error {
	query := `
		INSERT INTO virtual_pool_transactions (
			virtual_pool_id, chain_id, user_id, transaction_type, cnpy_amount,
			token_amount, price_per_token_cnpy, trading_fee_cnpy, slippage_percent,
			pool_cnpy_reserve_after, pool_token_reserve_after, market_cap_after_usd,
			transaction_hash, block_height, gas_used
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at`

	err := tx.QueryRowxContext(ctx, query,
		transaction.VirtualPoolID,
		transaction.ChainID,
		transaction.UserID,
		transaction.TransactionType,
		transaction.CNPYAmount,
		transaction.TokenAmount,
		transaction.PricePerTokenCNPY,
		transaction.TradingFeeCNPY,
		transaction.SlippagePercent,
		transaction.PoolCNPYReserveAfter,
		transaction.PoolTokenReserveAfter,
		transaction.MarketCapAfterUSD,
		transaction.TransactionHash,
		transaction.BlockHeight,
		transaction.GasUsed,
	).Scan(&transaction.ID, &transaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction in tx: %w", err)
	}

	return nil
}

// GetUserPositionForUpdate retrieves a user position with FOR UPDATE lock
func (r *virtualPoolTxRepository) GetUserPositionForUpdate(ctx context.Context, tx *sqlx.Tx, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	query := `
		SELECT id, user_id, chain_id, virtual_pool_id, token_balance, total_cnpy_invested,
			   total_cnpy_withdrawn, average_entry_price_cnpy, unrealized_pnl_cnpy,
			   realized_pnl_cnpy, total_return_percent, is_active, first_purchase_at,
			   last_activity_at, created_at, updated_at
		FROM user_virtual_positions
		WHERE user_id = $1 AND chain_id = $2
		FOR UPDATE`

	var position models.UserVirtualPosition
	var firstPurchaseAt, lastActivityAt sql.NullTime

	err := tx.QueryRowxContext(ctx, query, userID, chainID).Scan(
		&position.ID, &position.UserID, &position.ChainID, &position.VirtualPoolID,
		&position.TokenBalance, &position.TotalCNPYInvested, &position.TotalCNPYWithdrawn,
		&position.AverageEntryPriceCNPY, &position.UnrealizedPnlCNPY, &position.RealizedPnlCNPY,
		&position.TotalReturnPercent, &position.IsActive, &firstPurchaseAt,
		&lastActivityAt, &position.CreatedAt, &position.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Position doesn't exist yet, not an error
		}
		return nil, fmt.Errorf("failed to get user position with lock: %w", err)
	}

	if firstPurchaseAt.Valid {
		position.FirstPurchaseAt = &firstPurchaseAt.Time
	}
	if lastActivityAt.Valid {
		position.LastActivityAt = &lastActivityAt.Time
	}

	return &position, nil
}

// UpsertUserPositionInTx inserts or updates a user position within a transaction
func (r *virtualPoolTxRepository) UpsertUserPositionInTx(ctx context.Context, tx *sqlx.Tx, position *models.UserVirtualPosition) error {
	query := `
		INSERT INTO user_virtual_positions (
			user_id, chain_id, virtual_pool_id, token_balance, total_cnpy_invested,
			total_cnpy_withdrawn, average_entry_price_cnpy, unrealized_pnl_cnpy,
			realized_pnl_cnpy, total_return_percent, is_active, first_purchase_at,
			last_activity_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (user_id, chain_id)
		DO UPDATE SET
			token_balance = EXCLUDED.token_balance,
			total_cnpy_invested = EXCLUDED.total_cnpy_invested,
			total_cnpy_withdrawn = EXCLUDED.total_cnpy_withdrawn,
			average_entry_price_cnpy = EXCLUDED.average_entry_price_cnpy,
			unrealized_pnl_cnpy = EXCLUDED.unrealized_pnl_cnpy,
			realized_pnl_cnpy = EXCLUDED.realized_pnl_cnpy,
			total_return_percent = EXCLUDED.total_return_percent,
			is_active = EXCLUDED.is_active,
			last_activity_at = EXCLUDED.last_activity_at,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at`

	err := tx.QueryRowxContext(ctx, query,
		position.UserID,
		position.ChainID,
		position.VirtualPoolID,
		position.TokenBalance,
		position.TotalCNPYInvested,
		position.TotalCNPYWithdrawn,
		position.AverageEntryPriceCNPY,
		position.UnrealizedPnlCNPY,
		position.RealizedPnlCNPY,
		position.TotalReturnPercent,
		position.IsActive,
		position.FirstPurchaseAt,
		position.LastActivityAt,
	).Scan(&position.ID, &position.CreatedAt, &position.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert user position in tx: %w", err)
	}

	return nil
}
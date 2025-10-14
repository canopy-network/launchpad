package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type virtualPoolRepository struct {
	db *sqlx.DB
}

// NewVirtualPoolRepository creates a new PostgreSQL virtual pool repository
func NewVirtualPoolRepository(db *sqlx.DB) interfaces.VirtualPoolRepository {
	return &virtualPoolRepository{db: db}
}

// Create creates a new virtual pool
func (r *virtualPoolRepository) Create(ctx context.Context, pool *models.VirtualPool) (*models.VirtualPool, error) {
	query := `
		INSERT INTO virtual_pools (
			chain_id, cnpy_reserve, token_reserve, current_price_cnpy,
			total_transactions, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, chain_id, cnpy_reserve, token_reserve, current_price_cnpy,
				  market_cap_usd, total_volume_cnpy, total_transactions, unique_traders,
				  is_active, price_24h_change_percent, volume_24h_cnpy, high_24h_cnpy,
				  low_24h_cnpy, created_at, updated_at`

	var created models.VirtualPool
	err := r.db.QueryRowxContext(ctx, query,
		pool.ChainID,
		pool.CNPYReserve,
		pool.TokenReserve,
		pool.CurrentPriceCNPY,
		pool.TotalTransactions,
		pool.IsActive,
	).Scan(
		&created.ID, &created.ChainID, &created.CNPYReserve, &created.TokenReserve,
		&created.CurrentPriceCNPY, &created.MarketCapUSD, &created.TotalVolumeCNPY,
		&created.TotalTransactions, &created.UniqueTraders, &created.IsActive,
		&created.Price24hChangePercent, &created.Volume24hCNPY, &created.High24hCNPY,
		&created.Low24hCNPY, &created.CreatedAt, &created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create virtual pool: %w", err)
	}

	return &created, nil
}

// GetPoolByChainID retrieves a virtual pool by chain ID
func (r *virtualPoolRepository) GetPoolByChainID(ctx context.Context, chainID uuid.UUID) (*models.VirtualPool, error) {
	query := `
		SELECT id, chain_id, cnpy_reserve, token_reserve, current_price_cnpy, market_cap_usd,
			   total_volume_cnpy, total_transactions, unique_traders, is_active,
			   price_24h_change_percent, volume_24h_cnpy, high_24h_cnpy, low_24h_cnpy,
			   created_at, updated_at
		FROM virtual_pools
		WHERE chain_id = $1`

	var pool models.VirtualPool
	err := r.db.QueryRowxContext(ctx, query, chainID).Scan(
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
		return nil, fmt.Errorf("failed to get virtual pool: %w", err)
	}

	return &pool, nil
}

// GetAllPools retrieves all virtual pools with pagination
func (r *virtualPoolRepository) GetAllPools(ctx context.Context, pagination interfaces.Pagination) ([]models.VirtualPool, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM virtual_pools"
	var total int
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count virtual pools: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT id, chain_id, cnpy_reserve, token_reserve, current_price_cnpy, market_cap_usd,
			   total_volume_cnpy, total_transactions, unique_traders, is_active,
			   price_24h_change_percent, volume_24h_cnpy, high_24h_cnpy, low_24h_cnpy,
			   created_at, updated_at
		FROM virtual_pools
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	pools := []models.VirtualPool{}
	err = r.db.SelectContext(ctx, &pools, dataQuery, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query virtual pools: %w", err)
	}

	return pools, total, nil
}

// UpdatePoolState updates the virtual pool state with new values
func (r *virtualPoolRepository) UpdatePoolState(ctx context.Context, chainID uuid.UUID, update *interfaces.PoolStateUpdate) error {
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

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
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

// CreateTransaction creates a new virtual pool transaction record
func (r *virtualPoolRepository) CreateTransaction(ctx context.Context, transaction *models.VirtualPoolTransaction) error {
	query := `
		INSERT INTO virtual_pool_transactions (
			virtual_pool_id, chain_id, user_id, transaction_type, cnpy_amount,
			token_amount, price_per_token_cnpy, trading_fee_cnpy, slippage_percent,
			pool_cnpy_reserve_after, pool_token_reserve_after, market_cap_after_usd,
			transaction_hash, block_height, gas_used
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id, created_at`

	err := r.db.QueryRowxContext(ctx, query,
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
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransactionsByPoolID retrieves transactions for a virtual pool with pagination
func (r *virtualPoolRepository) GetTransactionsByPoolID(ctx context.Context, poolID uuid.UUID, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM virtual_pool_transactions WHERE virtual_pool_id = $1"
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, poolID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT id, virtual_pool_id, chain_id, user_id, transaction_type, cnpy_amount,
			   token_amount, price_per_token_cnpy, trading_fee_cnpy, slippage_percent,
			   pool_cnpy_reserve_after, pool_token_reserve_after, market_cap_after_usd,
			   transaction_hash, block_height, gas_used, created_at
		FROM virtual_pool_transactions
		WHERE virtual_pool_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	transactions := []models.VirtualPoolTransaction{}
	err = r.db.SelectContext(ctx, &transactions, dataQuery, poolID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}

	return transactions, total, nil
}

// GetTransactionsByUserID retrieves transactions for a user with pagination
func (r *virtualPoolRepository) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM virtual_pool_transactions WHERE user_id = $1"
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT id, virtual_pool_id, chain_id, user_id, transaction_type, cnpy_amount,
			   token_amount, price_per_token_cnpy, trading_fee_cnpy, slippage_percent,
			   pool_cnpy_reserve_after, pool_token_reserve_after, market_cap_after_usd,
			   transaction_hash, block_height, gas_used, created_at
		FROM virtual_pool_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	transactions := []models.VirtualPoolTransaction{}
	err = r.db.SelectContext(ctx, &transactions, dataQuery, userID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}

	return transactions, total, nil
}

// GetTransactionsByChainID retrieves transactions for a chain with filters and pagination
func (r *virtualPoolRepository) GetTransactionsByChainID(ctx context.Context, chainID uuid.UUID, filters interfaces.TransactionFilters, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	// Build WHERE clause based on filters
	whereConditions := []string{"chain_id = $1"}
	args := []interface{}{chainID}
	argCount := 1

	if filters.UserID != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, *filters.UserID)
	}

	if filters.TransactionType != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("transaction_type = $%d", argCount))
		args = append(args, filters.TransactionType)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM virtual_pool_transactions WHERE %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, virtual_pool_id, chain_id, user_id, transaction_type, cnpy_amount,
			   token_amount, price_per_token_cnpy, trading_fee_cnpy, slippage_percent,
			   pool_cnpy_reserve_after, pool_token_reserve_after, market_cap_after_usd,
			   transaction_hash, block_height, gas_used, created_at
		FROM virtual_pool_transactions
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argCount+1, argCount+2)

	args = append(args, pagination.Limit, pagination.Offset)

	transactions := []models.VirtualPoolTransaction{}
	err = r.db.SelectContext(ctx, &transactions, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}

	return transactions, total, nil
}

// GetUserPosition retrieves a user's position for a specific chain
func (r *virtualPoolRepository) GetUserPosition(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	query := `
		SELECT id, user_id, chain_id, virtual_pool_id, token_balance, total_cnpy_invested,
			   total_cnpy_withdrawn, average_entry_price_cnpy, unrealized_pnl_cnpy,
			   realized_pnl_cnpy, total_return_percent, is_active, first_purchase_at,
			   last_activity_at, created_at, updated_at
		FROM user_virtual_positions
		WHERE user_id = $1 AND chain_id = $2`

	var position models.UserVirtualPosition
	var firstPurchaseAt, lastActivityAt sql.NullTime

	err := r.db.QueryRowxContext(ctx, query, userID, chainID).Scan(
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
		return nil, fmt.Errorf("failed to get user position: %w", err)
	}

	if firstPurchaseAt.Valid {
		position.FirstPurchaseAt = &firstPurchaseAt.Time
	}
	if lastActivityAt.Valid {
		position.LastActivityAt = &lastActivityAt.Time
	}

	return &position, nil
}

// UpsertUserPosition inserts or updates a user position
func (r *virtualPoolRepository) UpsertUserPosition(ctx context.Context, position *models.UserVirtualPosition) error {
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

	err := r.db.QueryRowxContext(ctx, query,
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
		return fmt.Errorf("failed to upsert user position: %w", err)
	}

	return nil
}

// GetPositionsByChainID retrieves all user positions for a specific chain with pagination
func (r *virtualPoolRepository) GetPositionsByChainID(ctx context.Context, chainID uuid.UUID, pagination interfaces.Pagination) ([]models.UserVirtualPosition, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM user_virtual_positions WHERE chain_id = $1"
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, chainID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count positions: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT id, user_id, chain_id, virtual_pool_id, token_balance, total_cnpy_invested,
			   total_cnpy_withdrawn, average_entry_price_cnpy, unrealized_pnl_cnpy,
			   realized_pnl_cnpy, total_return_percent, is_active, first_purchase_at,
			   last_activity_at, created_at, updated_at
		FROM user_virtual_positions
		WHERE chain_id = $1
		ORDER BY token_balance DESC
		LIMIT $2 OFFSET $3`

	positions := []models.UserVirtualPosition{}
	err = r.db.SelectContext(ctx, &positions, dataQuery, chainID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query positions: %w", err)
	}

	return positions, total, nil
}

// GetPositionsWithUsersByChainID retrieves all positions with user wallet addresses for a chain
func (r *virtualPoolRepository) GetPositionsWithUsersByChainID(ctx context.Context, chainID uuid.UUID) ([]interfaces.UserPositionWithAddress, error) {
	query := `
		SELECT u.wallet_address, uvp.token_balance
		FROM user_virtual_positions uvp
		INNER JOIN users u ON uvp.user_id = u.id
		WHERE uvp.chain_id = $1 AND uvp.token_balance > 0
		ORDER BY uvp.token_balance DESC`

	var results []interfaces.UserPositionWithAddress
	err := r.db.SelectContext(ctx, &results, query, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions with users: %w", err)
	}

	return results, nil
}

// Helper function to convert big.Float to float64 safely
func bigFloatToFloat64(bf *big.Float) float64 {
	if bf == nil {
		return 0
	}
	f, _ := bf.Float64()
	return f
}
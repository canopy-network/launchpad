package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/canopy-network/canopy/lib"
	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/internal/repository/postgres"
	"github.com/enielson/launchpad/pkg/bondingcurve"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Transaction management errors
var (
	// ErrDeadlock indicates a database deadlock was detected
	ErrDeadlock = errors.New("transaction deadlock detected")

	// ErrMaxRetries indicates the maximum number of retry attempts was exceeded
	ErrMaxRetries = errors.New("max transaction retries exceeded")

	// ErrSerialization indicates a transaction serialization failure
	ErrSerialization = errors.New("serialization failure")
)

const (
	// MaxRetries is the maximum number of retry attempts for deadlock/serialization failures.
	// After 3 attempts with exponential backoff (100ms, 200ms, 400ms), the transaction is abandoned.
	MaxRetries = 3

	// RetryDelay is the base delay between retry attempts.
	// Actual delays use exponential backoff: 100ms, 200ms, 400ms for attempts 1, 2, 3.
	RetryDelay = 100 * time.Millisecond
)

// OrderProcessorTx is a transaction-aware order processor that handles buy and sell orders
// from the Canopy blockchain OrderBook with full ACID transaction guarantees.
//
// It uses row-level locking (SELECT FOR UPDATE) to prevent race conditions and automatically
// retries on deadlock or serialization failures. All updates to virtual_pools,
// virtual_pool_transactions, and user_virtual_positions are performed atomically within
// a single database transaction.
//
// Example usage:
//
//	processor := NewOrderProcessorTx(db, poolRepo, userRepo, nil)
//	err := processor.ProcessOrderWithRetry(ctx, order, chainID)
//	if err != nil {
//	    log.Error("Order processing failed", "error", err)
//	}
type OrderProcessorTx struct {
	db       *sqlx.DB
	poolRepo postgres.VirtualPoolTxRepository
	userRepo interfaces.UserRepository
	curve    *bondingcurve.BondingCurve
}

// NewOrderProcessorTx creates a new transaction-aware order processor.
//
// Parameters:
//   - db: Database connection for transaction management
//   - poolRepo: Transaction-aware virtual pool repository
//   - userRepo: User repository for address-to-user mapping
//   - curveConfig: Bonding curve configuration (nil uses default 1% fee)
//
// The returned processor is safe for concurrent use by multiple goroutines.
func NewOrderProcessorTx(
	db *sqlx.DB,
	poolRepo postgres.VirtualPoolTxRepository,
	userRepo interfaces.UserRepository,
	curveConfig *bondingcurve.BondingCurveConfig,
) *OrderProcessorTx {
	if curveConfig == nil {
		curveConfig = bondingcurve.NewBondingCurveConfig()
	}

	return &OrderProcessorTx{
		db:       db,
		poolRepo: poolRepo,
		userRepo: userRepo,
		curve:    bondingcurve.NewBondingCurve(curveConfig),
	}
}

// ProcessOrderWithRetry processes an order with automatic retry on deadlock/serialization failure.
//
// This is the main entry point for order processing and should be used by background workers.
// It automatically detects PostgreSQL deadlock (40P01) and serialization failure (40001) errors
// and retries the transaction up to MaxRetries times with exponential backoff.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - order: SellOrder from Canopy blockchain OrderBook
//   - chainID: UUID of the chain/virtual pool
//
// Returns:
//   - nil on success (order processed and committed)
//   - ErrMaxRetries if all retry attempts exhausted
//   - Other errors for validation failures, insufficient funds, etc. (no retry)
//
// Example:
//
//	err := processor.ProcessOrderWithRetry(ctx, order, chainID)
//	if errors.Is(err, services.ErrMaxRetries) {
//	    // Persistent deadlock - move to dead letter queue
//	    dlq.Enqueue(order)
//	}
func (op *OrderProcessorTx) ProcessOrderWithRetry(ctx context.Context, order *lib.SellOrder, chainID uuid.UUID) error {
	var err error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := RetryDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}

		err = op.ProcessOrder(ctx, order, chainID)
		if err == nil {
			return nil
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Log retry attempt (in production, use proper logging)
		fmt.Printf("Retrying order processing (attempt %d/%d) due to: %v\n", attempt+1, MaxRetries, err)
	}

	return fmt.Errorf("%w: %v", ErrMaxRetries, err)
}

// ProcessOrder processes a single order within a database transaction.
//
// All operations are performed atomically:
//  1. Acquire row locks on virtual pool and user position (FOR UPDATE)
//  2. Calculate bonding curve result
//  3. Update user position (token balance, PnL)
//  4. Create transaction record
//  5. Update pool state (reserves, price, volume)
//  6. Commit transaction (or rollback on any error)
//
// This method does NOT retry on deadlocks. Use ProcessOrderWithRetry instead.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - order: SellOrder from Canopy blockchain
//   - chainID: UUID of the chain/virtual pool
//
// Returns error if validation fails, insufficient funds, or database error.
func (op *OrderProcessorTx) ProcessOrder(ctx context.Context, order *lib.SellOrder, chainID uuid.UUID) error {
	// Validate the order
	if err := op.validateOrder(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	// Determine if this is a buy or sell
	isBuy := order.AmountForSale > 0 && order.RequestedAmount > 0

	// Execute order processing within a transaction
	err := database.Transaction(op.db, func(tx *sqlx.Tx) error {
		if isBuy {
			return op.processBuyOrderInTx(ctx, tx, order, chainID)
		}
		return op.processSellOrderInTx(ctx, tx, order, chainID)
	})

	if err != nil {
		return err
	}

	return nil
}

// processBuyOrderInTx handles a buy order within a transaction
func (op *OrderProcessorTx) processBuyOrderInTx(ctx context.Context, tx *sqlx.Tx, order *lib.SellOrder, chainID uuid.UUID) error {
	// Get the virtual pool with FOR UPDATE lock
	pool, err := op.poolRepo.GetPoolByChainIDForUpdate(ctx, tx, chainID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPoolNotFound, err)
	}

	// Convert order amounts to big.Float for bonding curve calculations
	cnpyAmountIn := big.NewFloat(float64(order.AmountForSale))

	// Create virtual pool for bonding curve
	virtualPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)),
	)

	// Execute the buy on the bonding curve
	result, err := op.curve.Buy(virtualPool, cnpyAmountIn)
	if err != nil {
		if errors.Is(err, bondingcurve.ErrInsufficientReserve) {
			return ErrInsufficientReserves
		}
		return fmt.Errorf("bonding curve buy failed: %w", err)
	}

	// Parse user ID from buyer address
	userID, err := op.extractUserID(order.BuyerReceiveAddress)
	if err != nil {
		return fmt.Errorf("invalid buyer address: %w", err)
	}

	// Get or create user position with FOR UPDATE lock
	position, err := op.poolRepo.GetUserPositionForUpdate(ctx, tx, userID, chainID)
	if err != nil {
		return fmt.Errorf("failed to get user position: %w", err)
	}

	now := time.Now()
	if position == nil {
		// Create new position
		position = &models.UserVirtualPosition{
			UserID:                userID,
			ChainID:               chainID,
			VirtualPoolID:         pool.ID,
			TokenBalance:          0,
			TotalCNPYInvested:     0,
			TotalCNPYWithdrawn:    0,
			AverageEntryPriceCNPY: 0,
			UnrealizedPnlCNPY:     0,
			RealizedPnlCNPY:       0,
			TotalReturnPercent:    0,
			IsActive:              true,
			FirstPurchaseAt:       &now,
			LastActivityAt:        &now,
		}
	}

	// Update user position with new purchase
	tokensReceived, _ := result.AmountOut.Int64()
	cnpySpent, _ := cnpyAmountIn.Float64()

	// Calculate new average entry price
	totalTokensAfter := position.TokenBalance + tokensReceived
	totalInvestedAfter := position.TotalCNPYInvested + cnpySpent
	newAveragePrice := totalInvestedAfter / float64(totalTokensAfter)

	// Calculate fees
	feeAmount := op.curve.GetConfig().CalculateFee(cnpyAmountIn)
	tradingFee, _ := feeAmount.Float64()

	// Update position fields
	position.TokenBalance = totalTokensAfter
	position.TotalCNPYInvested = totalInvestedAfter
	position.AverageEntryPriceCNPY = newAveragePrice
	position.LastActivityAt = &now
	if position.FirstPurchaseAt == nil {
		position.FirstPurchaseAt = &now
	}

	// Calculate unrealized PnL
	currentPrice, _ := result.Price.Float64()
	position.UnrealizedPnlCNPY = (currentPrice - position.AverageEntryPriceCNPY) * float64(position.TokenBalance)

	// Calculate total return percentage
	if position.TotalCNPYInvested > 0 {
		position.TotalReturnPercent = (position.UnrealizedPnlCNPY / position.TotalCNPYInvested) * 100
	}

	// Save user position within transaction
	if err := op.poolRepo.UpsertUserPositionInTx(ctx, tx, position); err != nil {
		return fmt.Errorf("failed to update user position: %w", err)
	}

	// Create transaction record
	newReserveCNPY, _ := result.NewCNPYReserve.Float64()
	newReserveToken, _ := result.NewTokenReserve.Int64()
	priceImpact, _ := result.PriceImpact.Float64()
	pricePerToken, _ := result.Price.Float64()

	transaction := &models.VirtualPoolTransaction{
		VirtualPoolID:         pool.ID,
		ChainID:               chainID,
		UserID:                userID,
		TransactionType:       "buy",
		CNPYAmount:            cnpySpent,
		TokenAmount:           tokensReceived,
		PricePerTokenCNPY:     pricePerToken,
		TradingFeeCNPY:        tradingFee,
		SlippagePercent:       priceImpact,
		PoolCNPYReserveAfter:  newReserveCNPY,
		PoolTokenReserveAfter: newReserveToken,
		MarketCapAfterUSD:     newReserveCNPY,
	}

	if err := op.poolRepo.CreateTransactionInTx(ctx, tx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update pool state within transaction
	poolUpdate := &interfaces.PoolStateUpdate{
		CNPYReserve:      result.NewCNPYReserve,
		TokenReserve:     result.NewTokenReserve,
		CurrentPriceCNPY: result.Price,
		MarketCapUSD:     result.NewCNPYReserve,
		TotalVolumeCNPY:  big.NewFloat(pool.TotalVolumeCNPY + cnpySpent),
	}

	newTxCount := pool.TotalTransactions + 1
	poolUpdate.TotalTransactions = &newTxCount

	if err := op.poolRepo.UpdatePoolStateInTx(ctx, tx, chainID, poolUpdate); err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	return nil
}

// processSellOrderInTx handles a sell order within a transaction
func (op *OrderProcessorTx) processSellOrderInTx(ctx context.Context, tx *sqlx.Tx, order *lib.SellOrder, chainID uuid.UUID) error {
	// Get the virtual pool with FOR UPDATE lock
	pool, err := op.poolRepo.GetPoolByChainIDForUpdate(ctx, tx, chainID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPoolNotFound, err)
	}

	// For sell orders, RequestedAmount represents tokens to sell
	tokenAmountIn := big.NewFloat(float64(order.RequestedAmount))

	// Create virtual pool for bonding curve
	virtualPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)),
	)

	// Execute the sell on the bonding curve
	result, err := op.curve.Sell(virtualPool, tokenAmountIn)
	if err != nil {
		if errors.Is(err, bondingcurve.ErrInsufficientReserve) {
			return ErrInsufficientReserves
		}
		if errors.Is(err, bondingcurve.ErrInsufficientTokens) {
			return ErrInsufficientBalance
		}
		return fmt.Errorf("bonding curve sell failed: %w", err)
	}

	// Parse user ID from seller address
	userID, err := op.extractUserID(order.SellersSendAddress)
	if err != nil {
		return fmt.Errorf("invalid seller address: %w", err)
	}

	// Get user position with FOR UPDATE lock
	position, err := op.poolRepo.GetUserPositionForUpdate(ctx, tx, userID, chainID)
	if err != nil {
		return fmt.Errorf("failed to get user position: %w", err)
	}
	if position == nil {
		return ErrInsufficientBalance
	}

	// Verify user has enough tokens
	tokensSold, _ := tokenAmountIn.Int64()
	if position.TokenBalance < tokensSold {
		return ErrInsufficientBalance
	}

	// Calculate proceeds from sale
	cnpyReceived, _ := result.AmountOut.Float64()
	pricePerToken, _ := result.Price.Float64()

	// Calculate fees
	feeAmount := op.curve.GetConfig().CalculateFee(result.AmountOut)
	tradingFee, _ := feeAmount.Float64()

	// Calculate realized PnL for this sale
	costBasis := position.AverageEntryPriceCNPY * float64(tokensSold)
	realizedPnL := cnpyReceived - costBasis

	now := time.Now()

	// Update position fields
	position.TokenBalance -= tokensSold
	position.TotalCNPYWithdrawn += cnpyReceived
	position.RealizedPnlCNPY += realizedPnL
	position.LastActivityAt = &now

	// Recalculate unrealized PnL with remaining balance
	if position.TokenBalance > 0 {
		currentPrice, _ := result.Price.Float64()
		position.UnrealizedPnlCNPY = (currentPrice - position.AverageEntryPriceCNPY) * float64(position.TokenBalance)
	} else {
		position.UnrealizedPnlCNPY = 0
		position.IsActive = false
	}

	// Calculate total return percentage
	totalValue := position.TotalCNPYWithdrawn + (pricePerToken * float64(position.TokenBalance))
	if position.TotalCNPYInvested > 0 {
		position.TotalReturnPercent = ((totalValue - position.TotalCNPYInvested) / position.TotalCNPYInvested) * 100
	}

	// Save user position within transaction
	if err := op.poolRepo.UpsertUserPositionInTx(ctx, tx, position); err != nil {
		return fmt.Errorf("failed to update user position: %w", err)
	}

	// Create transaction record
	newReserveCNPY, _ := result.NewCNPYReserve.Float64()
	newReserveToken, _ := result.NewTokenReserve.Int64()
	priceImpact, _ := result.PriceImpact.Float64()

	transaction := &models.VirtualPoolTransaction{
		VirtualPoolID:         pool.ID,
		ChainID:               chainID,
		UserID:                userID,
		TransactionType:       "sell",
		CNPYAmount:            cnpyReceived,
		TokenAmount:           tokensSold,
		PricePerTokenCNPY:     pricePerToken,
		TradingFeeCNPY:        tradingFee,
		SlippagePercent:       priceImpact,
		PoolCNPYReserveAfter:  newReserveCNPY,
		PoolTokenReserveAfter: newReserveToken,
		MarketCapAfterUSD:     newReserveCNPY,
	}

	if err := op.poolRepo.CreateTransactionInTx(ctx, tx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update pool state within transaction
	poolUpdate := &interfaces.PoolStateUpdate{
		CNPYReserve:      result.NewCNPYReserve,
		TokenReserve:     result.NewTokenReserve,
		CurrentPriceCNPY: result.Price,
		MarketCapUSD:     result.NewCNPYReserve,
		TotalVolumeCNPY:  big.NewFloat(pool.TotalVolumeCNPY + cnpyReceived),
	}

	newTxCount := pool.TotalTransactions + 1
	poolUpdate.TotalTransactions = &newTxCount

	if err := op.poolRepo.UpdatePoolStateInTx(ctx, tx, chainID, poolUpdate); err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	return nil
}

// validateOrder validates the order structure and fields
func (op *OrderProcessorTx) validateOrder(order *lib.SellOrder) error {
	if order == nil {
		return ErrInvalidOrder
	}

	if order.AmountForSale == 0 && order.RequestedAmount == 0 {
		return ErrZeroAmount
	}

	if order.BuyerReceiveAddress == nil && order.SellersSendAddress == nil {
		return fmt.Errorf("%w: missing user address", ErrInvalidOrder)
	}

	return nil
}

// extractUserID extracts a user ID from an address byte slice
func (op *OrderProcessorTx) extractUserID(address []byte) (uuid.UUID, error) {
	if len(address) == 0 {
		return uuid.Nil, fmt.Errorf("empty address")
	}

	// Try to parse as UUID directly
	addressStr := string(address)
	userID, err := uuid.Parse(addressStr)
	if err == nil {
		return userID, nil
	}

	// If not a UUID, treat as wallet address and look up user
	user, err := op.userRepo.GetByWalletAddress(context.Background(), addressStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user not found for address %s: %w", addressStr, err)
	}

	return user.ID, nil
}

// isRetryableError checks if an error is retryable (deadlock or serialization failure).
//
// Returns true for:
//   - PostgreSQL error code 40P01 (deadlock_detected)
//   - PostgreSQL error code 40001 (serialization_failure)
//   - ErrDeadlock or ErrSerialization
//
// All other errors are considered permanent and will not be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// PostgreSQL error codes:
		// 40001 - serialization_failure
		// 40P01 - deadlock_detected
		switch pqErr.Code {
		case "40001", "40P01":
			return true
		}
	}

	// Check for wrapped errors
	if errors.Is(err, ErrDeadlock) || errors.Is(err, ErrSerialization) {
		return true
	}

	return false
}

// GetConfig returns the bonding curve configuration.
//
// The configuration includes the fee rate in basis points (e.g., 100 = 1%).
// This can be used to inspect current fee settings or for simulation purposes.
func (op *OrderProcessorTx) GetConfig() *bondingcurve.BondingCurveConfig {
	return op.curve.GetConfig()
}

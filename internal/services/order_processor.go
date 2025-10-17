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
	"github.com/enielson/launchpad/pkg/bondingcurve"
	"github.com/google/uuid"
)

var (
	ErrInvalidOrder         = errors.New("invalid order")
	ErrPoolNotFound         = errors.New("virtual pool not found")
	ErrInsufficientReserves = errors.New("insufficient reserves in pool")
	ErrInsufficientBalance  = errors.New("insufficient user balance")
	ErrInvalidOrderType     = errors.New("invalid order type")
	ErrZeroAmount           = errors.New("order amount must be greater than zero")
	ErrUserNotFound         = errors.New("user not found")
)

// OrderProcessor handles processing of orders from Canopy OrderBook
type OrderProcessor struct {
	poolRepo interfaces.VirtualPoolRepository
	userRepo interfaces.UserRepository
	curve    *bondingcurve.BondingCurve
}

// NewOrderProcessor creates a new order processor service
func NewOrderProcessor(
	poolRepo interfaces.VirtualPoolRepository,
	userRepo interfaces.UserRepository,
	curveConfig *bondingcurve.BondingCurveConfig,
) *OrderProcessor {
	if curveConfig == nil {
		curveConfig = bondingcurve.NewBondingCurveConfig()
	}

	return &OrderProcessor{
		poolRepo: poolRepo,
		userRepo: userRepo,
		curve:    bondingcurve.NewBondingCurve(curveConfig),
	}
}

// ProcessOrder processes a single order from the Canopy OrderBook
// This determines if it's a buy or sell and executes the appropriate action
func (op *OrderProcessor) ProcessOrder(ctx context.Context, order *lib.SellOrder, chainID uuid.UUID) error {
	// Validate the order
	if err := op.validateOrder(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	// Determine if this is a buy or sell based on order fields
	// For a SellOrder: AmountForSale is CNPY, RequestedAmount is tokens
	// If user is selling CNPY for tokens -> BUY tokens
	// If user is selling tokens for CNPY -> SELL tokens
	isBuy := order.AmountForSale > 0 && order.RequestedAmount > 0

	if isBuy {
		return op.processBuyOrder(ctx, order, chainID)
	}

	return op.processSellOrder(ctx, order, chainID)
}

// processBuyOrder handles a buy order (user buying tokens with CNPY)
func (op *OrderProcessor) processBuyOrder(ctx context.Context, order *lib.SellOrder, chainID uuid.UUID) error {
	// Get the virtual pool for this chain
	pool, err := op.poolRepo.GetPoolByChainID(ctx, chainID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPoolNotFound, err)
	}

	// Convert order amounts to big.Float for bonding curve calculations
	cnpyAmountIn := big.NewFloat(float64(order.AmountForSale))

	// Create virtual pool for bonding curve
	virtualPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)), // total supply = token reserve for virtual pool
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

	// Get or create user position
	position, err := op.poolRepo.GetUserPosition(ctx, userID, chainID)
	if err != nil {
		return fmt.Errorf("failed to get user position: %w", err)
	}

	now := time.Now()
	if position == nil {
		// Create new position
		position = &models.UserVirtualLPPosition{
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
	pricePerToken, _ := result.Price.Float64()
	cnpySpent, _ := cnpyAmountIn.Float64()

	// Calculate new average entry price
	totalTokensAfter := position.TokenBalance + tokensReceived
	totalInvestedAfter := position.TotalCNPYInvested + cnpySpent
	newAveragePrice := totalInvestedAfter / float64(totalTokensAfter)

	// Calculate fees (from bonding curve config)
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

	// Save user position
	if err := op.poolRepo.UpsertUserPosition(ctx, position); err != nil {
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
		TransactionType:       "buy",
		CNPYAmount:            cnpySpent,
		TokenAmount:           tokensReceived,
		PricePerTokenCNPY:     pricePerToken,
		TradingFeeCNPY:        tradingFee,
		SlippagePercent:       priceImpact,
		PoolCNPYReserveAfter:  newReserveCNPY,
		PoolTokenReserveAfter: newReserveToken,
		MarketCapAfterUSD:     newReserveCNPY, // Simplified: market cap = CNPY reserve
	}

	if err := op.poolRepo.CreateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update pool state
	poolUpdate := &interfaces.PoolStateUpdate{
		CNPYReserve:      result.NewCNPYReserve,
		TokenReserve:     result.NewTokenReserve,
		CurrentPriceCNPY: result.Price,
		MarketCapUSD:     result.NewCNPYReserve, // Simplified
		TotalVolumeCNPY:  big.NewFloat(pool.TotalVolumeCNPY + cnpySpent),
	}

	// Increment transaction count
	newTxCount := pool.TotalTransactions + 1
	poolUpdate.TotalTransactions = &newTxCount

	if err := op.poolRepo.UpdatePoolState(ctx, chainID, poolUpdate); err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	return nil
}

// processSellOrder handles a sell order (user selling tokens for CNPY)
func (op *OrderProcessor) processSellOrder(ctx context.Context, order *lib.SellOrder, chainID uuid.UUID) error {
	// Get the virtual pool for this chain
	pool, err := op.poolRepo.GetPoolByChainID(ctx, chainID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPoolNotFound, err)
	}

	// For sell orders, RequestedAmount represents CNPY desired, AmountForSale represents tokens to sell
	// But we need to work backwards from the tokens being sold
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

	// Get user position
	position, err := op.poolRepo.GetUserPosition(ctx, userID, chainID)
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
		position.IsActive = false // No more tokens held
	}

	// Calculate total return percentage
	totalValue := position.TotalCNPYWithdrawn + (pricePerToken * float64(position.TokenBalance))
	if position.TotalCNPYInvested > 0 {
		position.TotalReturnPercent = ((totalValue - position.TotalCNPYInvested) / position.TotalCNPYInvested) * 100
	}

	// Save user position
	if err := op.poolRepo.UpsertUserPosition(ctx, position); err != nil {
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

	if err := op.poolRepo.CreateTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update pool state
	poolUpdate := &interfaces.PoolStateUpdate{
		CNPYReserve:      result.NewCNPYReserve,
		TokenReserve:     result.NewTokenReserve,
		CurrentPriceCNPY: result.Price,
		MarketCapUSD:     result.NewCNPYReserve,
		TotalVolumeCNPY:  big.NewFloat(pool.TotalVolumeCNPY + cnpyReceived),
	}

	// Increment transaction count
	newTxCount := pool.TotalTransactions + 1
	poolUpdate.TotalTransactions = &newTxCount

	if err := op.poolRepo.UpdatePoolState(ctx, chainID, poolUpdate); err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	return nil
}

// validateOrder validates the order structure and fields
func (op *OrderProcessor) validateOrder(order *lib.SellOrder) error {
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
// This assumes the address can be converted to a UUID or wallet address
// that maps to a user in the system
func (op *OrderProcessor) extractUserID(address []byte) (uuid.UUID, error) {
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
	// This requires the user to exist in the system
	user, err := op.userRepo.GetByWalletAddress(context.Background(), addressStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user not found for address %s: %w", addressStr, err)
	}

	return user.ID, nil
}

// GetConfig returns the bonding curve configuration
func (op *OrderProcessor) GetConfig() *bondingcurve.BondingCurveConfig {
	return op.curve.GetConfig()
}

// SimulateBuy simulates a buy order without executing it
func (op *OrderProcessor) SimulateBuy(ctx context.Context, chainID uuid.UUID, cnpyAmount float64) (*bondingcurve.TradeResult, error) {
	pool, err := op.poolRepo.GetPoolByChainID(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPoolNotFound, err)
	}

	virtualPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)),
	)

	return op.curve.SimulateBuy(virtualPool, big.NewFloat(cnpyAmount))
}

// SimulateSell simulates a sell order without executing it
func (op *OrderProcessor) SimulateSell(ctx context.Context, chainID uuid.UUID, tokenAmount int64) (*bondingcurve.TradeResult, error) {
	pool, err := op.poolRepo.GetPoolByChainID(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPoolNotFound, err)
	}

	virtualPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)),
	)

	return op.curve.SimulateSell(virtualPool, big.NewFloat(float64(tokenAmount)))
}

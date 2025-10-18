package fakevolume

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/bondingcurve"
	"github.com/google/uuid"
)

// Worker generates fake trading volume for virtual pools
type Worker struct {
	chainRepo   interfaces.ChainRepository
	poolRepo    interfaces.VirtualPoolRepository
	userRepo    interfaces.UserRepository
	interval    time.Duration
	stopChan    chan struct{}
	done        chan struct{}
	fakeUserIDs []uuid.UUID // Pool of fake user IDs to rotate through
}

// Config holds configuration for the fake volume worker
type Config struct {
	// Interval is how often to generate fake transactions (default: 30 seconds)
	Interval time.Duration
	// NumFakeUsers is the number of fake users to create and rotate through (default: 10)
	NumFakeUsers int
}

// DefaultConfig returns default configuration for the worker
func DefaultConfig() Config {
	return Config{
		Interval:     30 * time.Second,
		NumFakeUsers: 10,
	}
}

// NewWorker creates a new fake volume worker
func NewWorker(chainRepo interfaces.ChainRepository, poolRepo interfaces.VirtualPoolRepository, userRepo interfaces.UserRepository, config Config) *Worker {
	if config.Interval == 0 {
		config.Interval = 30 * time.Second
	}
	if config.NumFakeUsers == 0 {
		config.NumFakeUsers = 10
	}

	return &Worker{
		chainRepo:   chainRepo,
		poolRepo:    poolRepo,
		userRepo:    userRepo,
		interval:    config.Interval,
		stopChan:    make(chan struct{}),
		done:        make(chan struct{}),
		fakeUserIDs: make([]uuid.UUID, 0, config.NumFakeUsers),
	}
}

// Start begins the fake volume worker
func (w *Worker) Start() error {
	log.Printf("[FakeVolume Worker] Starting fake volume generation (interval: %v)", w.interval)

	// Initialize fake user pool on startup
	if err := w.initializeFakeUsers(); err != nil {
		return fmt.Errorf("failed to initialize fake users: %w", err)
	}

	go w.run()

	return nil
}

// Stop gracefully stops the fake volume worker
func (w *Worker) Stop() error {
	log.Println("[FakeVolume Worker] Stopping...")
	close(w.stopChan)

	// Wait for worker to finish current operation
	select {
	case <-w.done:
		log.Println("[FakeVolume Worker] Stopped")
	case <-time.After(10 * time.Second):
		log.Println("[FakeVolume Worker] Stop timeout")
	}

	return nil
}

// run is the main worker loop
func (w *Worker) run() {
	defer close(w.done)

	// Run generation immediately on start
	w.generateTransactions()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.generateTransactions()
		case <-w.stopChan:
			return
		}
	}
}

// generateTransactions creates fake transactions for all active chains
func (w *Worker) generateTransactions() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Query all chains with virtual_active status
	pagination := interfaces.Pagination{
		Page:  1,
		Limit: 100, // Get up to 100 chains per tick
	}
	chains, _, err := w.chainRepo.ListByStatus(ctx, models.ChainStatusVirtualActive, pagination)
	if err != nil {
		log.Printf("[FakeVolume Worker] Error querying chains: %v", err)
		return
	}

	if len(chains) == 0 {
		log.Printf("[FakeVolume Worker] No active chains found")
		return
	}

	log.Printf("[FakeVolume Worker] Generating transactions for %d active chain(s)", len(chains))

	// Generate 1-3 transactions per chain per interval
	for i := range chains {
		numTransactions := randomInt(1, 3)
		for j := 0; j < numTransactions; j++ {
			if err := w.generateSingleTransaction(ctx, &chains[i]); err != nil {
				log.Printf("[FakeVolume Worker] Error generating transaction for chain %s: %v", chains[i].ChainName, err)
			}
		}
	}
}

// generateSingleTransaction creates a single fake transaction
func (w *Worker) generateSingleTransaction(ctx context.Context, chain *models.Chain) error {
	// Get the virtual pool for this chain
	pool, err := w.poolRepo.GetPoolByChainID(ctx, chain.ID)
	if err != nil {
		return fmt.Errorf("failed to get virtual pool: %w", err)
	}

	// Determine transaction type (60% buy, 40% sell)
	isBuy := randomInt(1, 100) <= 60

	// Get or create a random user
	user, err := w.getOrCreateRandomUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get random user: %w", err)
	}

	if isBuy {
		return w.executeBuy(ctx, chain, pool, user)
	}
	return w.executeSell(ctx, chain, pool, user)
}

// executeBuy executes a fake buy transaction
func (w *Worker) executeBuy(ctx context.Context, chain *models.Chain, pool *models.VirtualPool, user *models.User) error {
	// Random CNPY amount between 0.1 and 10.0 CNPY
	cnpyAmount := randomFloat(0.1, 10.0)

	// Convert database pool to bonding curve pool
	bcPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)),
	)

	// Create bonding curve instance with default config
	curve := bondingcurve.NewBondingCurve(nil)

	// Execute the buy
	result, err := curve.Buy(bcPool, big.NewFloat(cnpyAmount))
	if err != nil {
		return fmt.Errorf("bonding curve buy failed: %w", err)
	}

	// Update pool state in database
	totalTransactions := pool.TotalTransactions + 1
	totalVolumeCNPY := pool.TotalVolumeCNPY + cnpyAmount
	totalVolumeCNPYBigFloat := big.NewFloat(totalVolumeCNPY)
	update := &interfaces.PoolStateUpdate{
		CNPYReserve:       result.NewCNPYReserve,
		TokenReserve:      result.NewTokenReserve,
		CurrentPriceCNPY:  result.Price,
		TotalVolumeCNPY:   totalVolumeCNPYBigFloat,
		TotalTransactions: &totalTransactions,
	}

	err = w.poolRepo.UpdatePoolState(ctx, chain.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	// Record transaction
	tokensOutFloat, _ := result.AmountOut.Float64()
	priceFloat, _ := result.Price.Float64()
	newCNPYReserveFloat, _ := result.NewCNPYReserve.Float64()
	newTokenReserveFloat, _ := result.NewTokenReserve.Float64()

	transaction := &models.VirtualPoolTransaction{
		VirtualPoolID:         pool.ID,
		ChainID:               chain.ID,
		UserID:                user.ID,
		TransactionType:       "buy",
		CNPYAmount:            cnpyAmount,
		TokenAmount:           int64(tokensOutFloat),
		PricePerTokenCNPY:     priceFloat,
		TradingFeeCNPY:        0.0,
		SlippagePercent:       0,
		TransactionHash:       nil,
		BlockHeight:           nil,
		GasUsed:               nil,
		PoolCNPYReserveAfter:  newCNPYReserveFloat,
		PoolTokenReserveAfter: int64(newTokenReserveFloat),
		MarketCapAfterUSD:     0,
	}

	err = w.poolRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Update user position
	err = w.updateUserPosition(ctx, user, pool, chain, cnpyAmount, tokensOutFloat, priceFloat, true)
	if err != nil {
		log.Printf("[FakeVolume Worker] Warning: Failed to update user position: %v", err)
	}

	log.Printf("[FakeVolume Worker] BUY: Chain=%s, CNPY=%.4f, Tokens=%.2f, Price=%.8f",
		chain.ChainName, cnpyAmount, tokensOutFloat, priceFloat)

	return nil
}

// executeSell executes a fake sell transaction
func (w *Worker) executeSell(ctx context.Context, chain *models.Chain, pool *models.VirtualPool, user *models.User) error {
	// Get user position (create if doesn't exist with some tokens)
	position, err := w.poolRepo.GetUserPosition(ctx, user.ID, chain.ID)
	if err != nil {
		return fmt.Errorf("failed to get user position: %w", err)
	}

	// If user has no position or insufficient tokens, give them some tokens first
	if position == nil || position.TokenBalance < 1000 {
		// Give user tokens by doing a buy first
		return w.executeBuy(ctx, chain, pool, user)
	}

	// Random token amount to sell (10% to 50% of balance)
	sellPercent := randomFloat(0.1, 0.5)
	tokenAmount := float64(position.TokenBalance) * sellPercent

	// Convert database pool to bonding curve pool
	bcPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)),
	)

	// Create bonding curve instance with default config
	curve := bondingcurve.NewBondingCurve(nil)

	// Execute the sell
	result, err := curve.Sell(bcPool, big.NewFloat(tokenAmount))
	if err != nil {
		return fmt.Errorf("bonding curve sell failed: %w", err)
	}

	cnpyOutFloat, _ := result.AmountOut.Float64()
	priceFloat, _ := result.Price.Float64()
	newCNPYReserveFloat, _ := result.NewCNPYReserve.Float64()
	newTokenReserveFloat, _ := result.NewTokenReserve.Float64()

	// Update pool state in database
	totalTransactions := pool.TotalTransactions + 1
	totalVolumeCNPY := pool.TotalVolumeCNPY + cnpyOutFloat
	totalVolumeCNPYBigFloat := big.NewFloat(totalVolumeCNPY)
	update := &interfaces.PoolStateUpdate{
		CNPYReserve:       result.NewCNPYReserve,
		TokenReserve:      result.NewTokenReserve,
		CurrentPriceCNPY:  result.Price,
		TotalVolumeCNPY:   totalVolumeCNPYBigFloat,
		TotalTransactions: &totalTransactions,
	}

	err = w.poolRepo.UpdatePoolState(ctx, chain.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	// Record transaction
	transaction := &models.VirtualPoolTransaction{
		VirtualPoolID:         pool.ID,
		ChainID:               chain.ID,
		UserID:                user.ID,
		TransactionType:       "sell",
		CNPYAmount:            cnpyOutFloat,
		TokenAmount:           int64(tokenAmount),
		PricePerTokenCNPY:     priceFloat,
		TradingFeeCNPY:        0.0,
		SlippagePercent:       0,
		TransactionHash:       nil,
		BlockHeight:           nil,
		GasUsed:               nil,
		PoolCNPYReserveAfter:  newCNPYReserveFloat,
		PoolTokenReserveAfter: int64(newTokenReserveFloat),
		MarketCapAfterUSD:     0,
	}

	err = w.poolRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Update user position (selling)
	err = w.updateUserPosition(ctx, user, pool, chain, cnpyOutFloat, tokenAmount, priceFloat, false)
	if err != nil {
		log.Printf("[FakeVolume Worker] Warning: Failed to update user position: %v", err)
	}

	log.Printf("[FakeVolume Worker] SELL: Chain=%s, Tokens=%.2f, CNPY=%.4f, Price=%.8f",
		chain.ChainName, tokenAmount, cnpyOutFloat, priceFloat)

	return nil
}

// updateUserPosition updates or creates a user's virtual position
func (w *Worker) updateUserPosition(ctx context.Context, user *models.User, pool *models.VirtualPool, chain *models.Chain, cnpyAmount, tokenAmount, price float64, isBuy bool) error {
	position, err := w.poolRepo.GetUserPosition(ctx, user.ID, chain.ID)
	if err != nil {
		return fmt.Errorf("failed to get user position: %w", err)
	}

	now := time.Now()

	if position == nil {
		// Create new position (only for buys)
		if !isBuy {
			return nil
		}

		position = &models.UserVirtualLPPosition{
			UserID:                user.ID,
			ChainID:               chain.ID,
			VirtualPoolID:         pool.ID,
			TokenBalance:          int64(tokenAmount),
			TotalCNPYInvested:     cnpyAmount,
			TotalCNPYWithdrawn:    0,
			AverageEntryPriceCNPY: price,
			UnrealizedPnlCNPY:     0,
			RealizedPnlCNPY:       0,
			TotalReturnPercent:    0,
			IsActive:              true,
			FirstPurchaseAt:       &now,
			LastActivityAt:        &now,
		}
	} else {
		// Update existing position
		if isBuy {
			oldBalance := float64(position.TokenBalance)
			newBalance := oldBalance + tokenAmount

			// Calculate new average entry price (weighted average)
			oldInvestment := position.TotalCNPYInvested
			newInvestment := oldInvestment + cnpyAmount
			position.AverageEntryPriceCNPY = newInvestment / newBalance

			// Update balances
			position.TokenBalance = int64(newBalance)
			position.TotalCNPYInvested = newInvestment
		} else {
			// Sell
			newBalance := float64(position.TokenBalance) - tokenAmount
			if newBalance < 0 {
				newBalance = 0
			}

			position.TokenBalance = int64(newBalance)
			position.TotalCNPYWithdrawn += cnpyAmount

			// Calculate realized PnL on this sale
			costBasis := position.AverageEntryPriceCNPY * tokenAmount
			realizedPnL := cnpyAmount - costBasis
			position.RealizedPnlCNPY += realizedPnL
		}

		position.LastActivityAt = &now

		// Calculate unrealized PnL for remaining position
		if position.TokenBalance > 0 {
			currentValueCNPY := price * float64(position.TokenBalance)
			costBasis := position.AverageEntryPriceCNPY * float64(position.TokenBalance)
			position.UnrealizedPnlCNPY = currentValueCNPY - costBasis

			// Calculate total return percentage
			totalInvestment := position.TotalCNPYInvested
			if totalInvestment > 0 {
				totalPnL := position.UnrealizedPnlCNPY + position.RealizedPnlCNPY
				position.TotalReturnPercent = (totalPnL / totalInvestment) * 100
			}
		} else {
			position.UnrealizedPnlCNPY = 0
			position.IsActive = false
		}
	}

	// Upsert the position
	err = w.poolRepo.UpsertUserPosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to upsert user position: %w", err)
	}

	return nil
}

// initializeFakeUsers creates a pool of fake users to rotate through
func (w *Worker) initializeFakeUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numUsers := cap(w.fakeUserIDs)
	log.Printf("[FakeVolume Worker] Initializing pool of %d fake users...", numUsers)

	for i := 0; i < numUsers; i++ {
		walletAddress := fmt.Sprintf("0xfake%032d", i)

		// Try to get existing user first
		user, err := w.userRepo.GetByWalletAddress(ctx, walletAddress)
		if err == nil && user != nil {
			w.fakeUserIDs = append(w.fakeUserIDs, user.ID)
			continue
		}

		// Create new fake user
		newUser := &models.User{
			WalletAddress:    walletAddress,
			Username:         nil,
			IsVerified:       false,
			VerificationTier: "basic",
		}

		user, err = w.userRepo.Create(ctx, newUser)
		if err != nil {
			return fmt.Errorf("failed to create fake user %d: %w", i, err)
		}

		w.fakeUserIDs = append(w.fakeUserIDs, user.ID)
	}

	log.Printf("[FakeVolume Worker] Initialized %d fake users", len(w.fakeUserIDs))
	return nil
}

// getOrCreateRandomUser gets a random user from the fake user pool
func (w *Worker) getOrCreateRandomUser(ctx context.Context) (*models.User, error) {
	if len(w.fakeUserIDs) == 0 {
		return nil, fmt.Errorf("fake user pool is empty")
	}

	// Pick a random user from the pool
	randomIndex := randomInt(0, len(w.fakeUserIDs)-1)
	userID := w.fakeUserIDs[randomIndex]

	// Get the user from database
	user, err := w.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fake user: %w", err)
	}

	return user, nil
}

// Helper functions

// randomInt returns a random integer between min and max (inclusive)
func randomInt(min, max int) int {
	if min >= max {
		return min
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return int(n.Int64()) + min
}

// randomFloat returns a random float64 between min and max
func randomFloat(min, max float64) float64 {
	// Generate random bytes
	var b [8]byte
	rand.Read(b[:])

	// Convert to uint64
	n := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56

	// Normalize to [0, 1)
	f := float64(n) / float64(1<<64)

	// Scale to [min, max)
	return min + f*(max-min)
}

// generateRandomAddress generates a random wallet address
func generateRandomAddress() string {
	b := make([]byte, 20)
	rand.Read(b)
	return "0x" + hex.EncodeToString(b)
}

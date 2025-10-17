package newblock

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/canopy-network/canopy/fsm"
	"github.com/canopy-network/canopy/lib"
	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/bondingcurve"
	"github.com/enielson/launchpad/pkg/sub"
)

// RPCClient defines the interface for fetching blockchain data
type RPCClient interface {
	TransactionsByHeight(height uint64, page lib.PageParams) (*lib.Page, lib.ErrorI)
}

// Worker manages the root chain subscription and transaction processing
type Worker struct {
	subscription *sub.Subscription
	rpcClient    RPCClient
	chainRepo    interfaces.ChainRepository
	poolRepo     interfaces.VirtualPoolRepository
	userRepo     interfaces.UserRepository
	logger       sub.Logger
}

// Config holds the configuration for the newblock worker
type Config struct {
	RootChainURL    string // WebSocket URL for subscription
	RootChainID     uint64 // Chain ID to subscribe to
	RootChainRPCURL string // HTTP URL for RPC client
}

// NewWorker creates a new root chain event worker
func NewWorker(config Config, rpcClient RPCClient, chainRepo interfaces.ChainRepository, poolRepo interfaces.VirtualPoolRepository, userRepo interfaces.UserRepository) *Worker {
	logger := NewLogger()

	// Create subscription config
	subConfig := sub.Config{
		ChainId: config.RootChainID,
		Url:     config.RootChainURL,
	}

	// Create worker instance
	worker := &Worker{
		rpcClient: rpcClient,
		chainRepo: chainRepo,
		poolRepo:  poolRepo,
		userRepo:  userRepo,
		logger:    logger,
	}

	// Create subscription with event handler
	worker.subscription = sub.NewSubscription(subConfig, worker.handleRootChainEvent, logger)

	return worker
}

// Start begins the worker's subscription to root chain events
func (w *Worker) Start() error {
	log.Printf("[NewBlock Worker] Starting root chain subscription...")
	return w.subscription.Start()
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop() error {
	log.Printf("[NewBlock Worker] Stopping root chain subscription...")
	return w.subscription.Stop()
}

// IsConnected returns whether the worker is connected to the root chain
func (w *Worker) IsConnected() bool {
	return w.subscription.IsConnected()
}

// HandleRootChainEventForTest exposes the event handler for integration testing
// This allows tests to simulate receiving RootChainInfo events without needing
// a live Canopy node connection
func (w *Worker) HandleRootChainEventForTest(info *lib.RootChainInfo) error {
	return w.handleRootChainEvent(info)
}

// handleRootChainEvent processes new root chain info received from the subscription
func (w *Worker) handleRootChainEvent(info *lib.RootChainInfo) error {
	log.Printf("[NewBlock Worker] Processing block at height %d from root chain %d",
		info.Height, info.RootChainId)

	// Fetch transactions at this height using the RPC client
	pageParams := lib.PageParams{
		PageNumber: 1,
		PerPage:    100, // Adjust as needed
	}

	page, err := w.rpcClient.TransactionsByHeight(info.Height, pageParams)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions at height %d: %w", info.Height, err)
	}

	// Process transactions
	if page != nil && page.Results != nil {
		// Type assert the Results to TxResults
		txResults, ok := page.Results.(*lib.TxResults)
		if !ok {
			return fmt.Errorf("unexpected page results type at height %d", info.Height)
		}

		if len(*txResults) > 0 {
			log.Printf("[NewBlock Worker] Found %d transactions at height %d", len(*txResults), info.Height)

			for i, txResult := range *txResults {
				// Only process send transactions
				if txResult.MessageType != fsm.MessageSendName {
					continue
				}
				w.processTransaction(context.Background(), txResult, i, len(*txResults), info.Height)
			}
		} else {
			log.Printf("[NewBlock Worker] No transactions found at height %d", info.Height)
		}
	} else {
		log.Printf("[NewBlock Worker] No transactions found at height %d", info.Height)
	}

	return nil
}

// processTransaction processes a single transaction from a block
func (w *Worker) processTransaction(ctx context.Context, txResult *lib.TxResult, index int, total int, height uint64) {
	// Look up chain by recipient address (destination of send transaction)
	recipientAddress := hex.EncodeToString(txResult.Recipient)
	chain, err := w.chainRepo.GetByAddress(ctx, recipientAddress)
	if err != nil {
		// Chain not found or error - skip this transaction
		log.Printf("[NewBlock Worker] Transaction %d/%d at height %d: No chain found for recipient address %s",
			index+1, total, height, recipientAddress)
		return
	}

	// Found a chain matching this transaction's recipient
	log.Printf("[NewBlock Worker] Transaction %d/%d at height %d: Hash=%s, MessageType=%s, Sender=%s, Chain=%s (ID: %s)",
		index+1, total, height, txResult.TxHash, txResult.MessageType,
		txResult.Sender, chain.ChainName, chain.ID)

	// Extract the send amount from the transaction
	amount, err := w.extractSendAmount(txResult)
	if err != nil {
		log.Printf("[NewBlock Worker] Failed to extract send amount: %v", err)
		return
	}

	// Process the deposit to the virtual pool
	err = w.processDeposit(ctx, chain, amount, txResult.Sender)
	if err != nil {
		log.Printf("[NewBlock Worker] Failed to process deposit: %v", err)
		return
	}
}

// extractSendAmount extracts the CNPY amount from a send transaction
func (w *Worker) extractSendAmount(txResult *lib.TxResult) (uint64, error) {
	// Check if this is a send transaction
	if txResult.MessageType != fsm.MessageSendName {
		return 0, fmt.Errorf("not a send transaction: %s", txResult.MessageType)
	}

	// Get the transaction from the result
	if txResult.Transaction == nil || txResult.Transaction.Msg == nil {
		return 0, fmt.Errorf("transaction or message is nil")
	}

	// Unmarshal the Any message into MessageSend
	msg, err := lib.FromAny(txResult.Transaction.Msg)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal transaction message: %w", err)
	}

	// Type assert to MessageSend
	sendMsg, ok := msg.(*fsm.MessageSend)
	if !ok {
		return 0, fmt.Errorf("message is not MessageSend type")
	}

	return sendMsg.Amount, nil
}

// processDeposit handles a CNPY deposit to a chain's virtual pool
func (w *Worker) processDeposit(ctx context.Context, chain *models.Chain, amount uint64, sender []byte) error {
	log.Printf("[NewBlock Worker] Processing deposit: Chain=%s, Amount=%d uCNPY, Sender=%x",
		chain.ChainName, amount, sender)

	// Get the virtual pool for this chain
	pool, err := w.poolRepo.GetPoolByChainID(ctx, chain.ID)
	if err != nil {
		return fmt.Errorf("failed to get virtual pool: %w", err)
	}

	// Convert amount from micro-CNPY (uint64) to CNPY (big.Float)
	// 1 CNPY = 1,000,000 uCNPY
	cnpyAmount := new(big.Float).SetUint64(amount)
	cnpyAmount.Quo(cnpyAmount, big.NewFloat(1000000))

	// Convert database pool to bonding curve pool
	bcPool := bondingcurve.NewVirtualPool(
		big.NewFloat(pool.CNPYReserve),
		big.NewFloat(float64(pool.TokenReserve)),
		big.NewFloat(float64(pool.TokenReserve)), // Using token reserve as total supply for now
	)

	// Create bonding curve instance with default config
	curve := bondingcurve.NewBondingCurve(nil)

	// Execute the buy (deposit CNPY, receive tokens)
	result, err := curve.Buy(bcPool, cnpyAmount)
	if err != nil {
		return fmt.Errorf("bonding curve buy failed: %w", err)
	}

	log.Printf("[NewBlock Worker] Deposit result: TokensOut=%.6f, NewCNPYReserve=%.6f, NewTokenReserve=%.6f, Price=%.8f",
		result.AmountOut, result.NewCNPYReserve, result.NewTokenReserve, result.Price)

	// Update pool state in database
	totalTransactions := pool.TotalTransactions + 1
	update := &interfaces.PoolStateUpdate{
		CNPYReserve:       result.NewCNPYReserve,
		TokenReserve:      result.NewTokenReserve,
		CurrentPriceCNPY:  result.Price,
		TotalTransactions: &totalTransactions,
	}

	err = w.poolRepo.UpdatePoolState(ctx, chain.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update pool state: %w", err)
	}

	log.Printf("[NewBlock Worker] Successfully processed deposit for chain %s: CNPY %.6f → Tokens %.6f (Price: %.8f CNPY/token)",
		chain.ChainName, cnpyAmount, result.AmountOut, result.Price)

	// Record transaction in virtual_pool_transactions table
	user, err := w.recordTransaction(ctx, pool, chain, sender, cnpyAmount, result)
	if err != nil {
		log.Printf("[NewBlock Worker] Warning: Failed to record transaction: %v", err)
		// Don't fail the entire deposit if transaction recording fails
	}

	// Update or create user_virtual_positions for the sender
	if user != nil {
		err = w.updateUserPosition(ctx, user, pool, chain, cnpyAmount, result)
		if err != nil {
			log.Printf("[NewBlock Worker] Warning: Failed to update user position: %v", err)
			// Don't fail the entire deposit if position update fails
		}
	}

	return nil
}

// recordTransaction creates a record in virtual_pool_transactions table and returns the user
func (w *Worker) recordTransaction(ctx context.Context, pool *models.VirtualPool, chain *models.Chain, sender []byte, cnpyAmount *big.Float, result *bondingcurve.TradeResult) (*models.User, error) {
	// Convert sender address to hex string with 0x prefix (database stores addresses with 0x prefix)
	senderAddress := "0x" + hex.EncodeToString(sender)

	// Look up or create user by wallet address
	user, err := w.userRepo.GetByWalletAddress(ctx, senderAddress)
	if err != nil {
		// If user doesn't exist, create a new one
		newUser := &models.User{
			WalletAddress: senderAddress,
			Username:      nil, // Will be set when user completes profile
			IsVerified:    false,
		}
		user, err = w.userRepo.Create(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create user for address %s: %w", senderAddress, err)
		}
		log.Printf("[NewBlock Worker] Created new user for wallet address: %s", senderAddress)
	}

	// Convert big.Float amounts to float64
	cnpyFloat, _ := cnpyAmount.Float64()
	tokensOutFloat, _ := result.AmountOut.Float64()
	priceFloat, _ := result.Price.Float64()
	newCNPYReserveFloat, _ := result.NewCNPYReserve.Float64()
	newTokenReserveFloat, _ := result.NewTokenReserve.Float64()

	// Calculate trading fee (assumed to be included in the bonding curve result)
	// For now, we'll set it to 0 or calculate based on config
	tradingFeeCNPY := 0.0

	// Create transaction record
	transaction := &models.VirtualPoolTransaction{
		VirtualPoolID:         pool.ID,
		ChainID:               chain.ID,
		UserID:                user.ID,
		TransactionType:       "buy",
		CNPYAmount:            cnpyFloat,
		TokenAmount:           int64(tokensOutFloat),
		PricePerTokenCNPY:     priceFloat,
		TradingFeeCNPY:        tradingFeeCNPY,
		SlippagePercent:       0, // Could be calculated if needed
		TransactionHash:       nil,
		BlockHeight:           nil,
		GasUsed:               nil,
		PoolCNPYReserveAfter:  newCNPYReserveFloat,
		PoolTokenReserveAfter: int64(newTokenReserveFloat),
		MarketCapAfterUSD:     0, // Could be calculated: price * total_supply
	}

	err = w.poolRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	log.Printf("[NewBlock Worker] Recorded transaction: User=%s, Type=buy, CNPY=%.6f, Tokens=%d",
		user.ID, cnpyFloat, transaction.TokenAmount)

	return user, nil
}

// updateUserPosition updates or creates a user's virtual position
func (w *Worker) updateUserPosition(ctx context.Context, user *models.User, pool *models.VirtualPool, chain *models.Chain, cnpyAmount *big.Float, result *bondingcurve.TradeResult) error {
	// Get existing position or create new one
	position, err := w.poolRepo.GetUserPosition(ctx, user.ID, chain.ID)
	if err != nil {
		return fmt.Errorf("failed to get user position: %w", err)
	}

	// Convert amounts
	cnpyFloat, _ := cnpyAmount.Float64()
	tokensOutFloat, _ := result.AmountOut.Float64()
	priceFloat, _ := result.Price.Float64()

	now := time.Now()

	if position == nil {
		// Create new position
		position = &models.UserVirtualLPPosition{
			UserID:                user.ID,
			ChainID:               chain.ID,
			VirtualPoolID:         pool.ID,
			TokenBalance:          int64(tokensOutFloat),
			TotalCNPYInvested:     cnpyFloat,
			TotalCNPYWithdrawn:    0,
			AverageEntryPriceCNPY: priceFloat,
			UnrealizedPnlCNPY:     0, // Will be calculated based on current price vs entry price
			RealizedPnlCNPY:       0,
			TotalReturnPercent:    0,
			IsActive:              true,
			FirstPurchaseAt:       &now,
			LastActivityAt:        &now,
		}
	} else {
		// Update existing position
		oldBalance := float64(position.TokenBalance)
		newBalance := oldBalance + tokensOutFloat

		// Calculate new average entry price (weighted average)
		oldInvestment := position.TotalCNPYInvested
		newInvestment := oldInvestment + cnpyFloat
		position.AverageEntryPriceCNPY = newInvestment / newBalance

		// Update balances
		position.TokenBalance = int64(newBalance)
		position.TotalCNPYInvested = newInvestment
		position.LastActivityAt = &now

		// Calculate unrealized PnL
		// PnL = (current_price - avg_entry_price) * token_balance
		currentValueCNPY := priceFloat * newBalance
		position.UnrealizedPnlCNPY = currentValueCNPY - newInvestment

		// Calculate total return percentage
		if newInvestment > 0 {
			position.TotalReturnPercent = (position.UnrealizedPnlCNPY / newInvestment) * 100
		}
	}

	// Upsert the position
	err = w.poolRepo.UpsertUserPosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to upsert user position: %w", err)
	}

	log.Printf("[NewBlock Worker] Updated user position: User=%s, Chain=%s, Balance=%d tokens, Invested=%.6f CNPY, PnL=%.6f CNPY",
		user.ID, chain.ChainName, position.TokenBalance, position.TotalCNPYInvested, position.UnrealizedPnlCNPY)

	return nil
}

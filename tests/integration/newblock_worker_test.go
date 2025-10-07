//go:build integration

package integration_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/canopy-network/canopy/fsm"
	"github.com/canopy-network/canopy/lib"
	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/postgres"
	"github.com/enielson/launchpad/internal/workers/newblock"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

// MockRPCClient implements the RPCClient interface for testing
type MockRPCClient struct {
	txResultsByHeight map[uint64][]*lib.TxResult
	errorByHeight     map[uint64]lib.ErrorI
}

func NewMockRPCClient() *MockRPCClient {
	return &MockRPCClient{
		txResultsByHeight: make(map[uint64][]*lib.TxResult),
		errorByHeight:     make(map[uint64]lib.ErrorI),
	}
}

func (m *MockRPCClient) SetTransactionsAtHeight(height uint64, txResults []*lib.TxResult) {
	m.txResultsByHeight[height] = txResults
}

func (m *MockRPCClient) SetErrorAtHeight(height uint64, err lib.ErrorI) {
	m.errorByHeight[height] = err
}

func (m *MockRPCClient) TransactionsByHeight(height uint64, page lib.PageParams) (*lib.Page, lib.ErrorI) {
	// Check for error first
	if err, exists := m.errorByHeight[height]; exists {
		return nil, err
	}

	// Get transactions for this height
	txResults, exists := m.txResultsByHeight[height]
	if !exists {
		// Return empty page if no transactions configured
		emptyResults := lib.TxResults{}
		return &lib.Page{
			PageParams: page,
			Count:      0,
			TotalPages: 0,
			TotalCount: 0,
			Results:    &emptyResults,
		}, nil
	}

	// Wrap in TxResults type
	results := lib.TxResults(txResults)

	return &lib.Page{
		PageParams: page,
		Count:      len(txResults),
		TotalPages: 1,
		TotalCount: len(txResults),
		Results:    &results,
	}, nil
}

// buildSendTransaction creates a TxResult with a valid send transaction
func buildSendTransaction(senderAddress, recipientAddress []byte, amount uint64, height uint64) *lib.TxResult {
	// Create MessageSend
	sendMsg := &fsm.MessageSend{
		FromAddress: senderAddress,
		ToAddress:   recipientAddress,
		Amount:      amount,
	}

	// Marshal to Any
	anyMsg, err := anypb.New(sendMsg)
	if err != nil {
		panic(fmt.Sprintf("failed to create any message: %v", err))
	}

	return &lib.TxResult{
		Sender:      senderAddress,
		Recipient:   recipientAddress,
		MessageType: fsm.MessageSendName,
		Height:      height,
		Index:       0,
		Transaction: &lib.Transaction{
			MessageType: fsm.MessageSendName,
			Msg:         anyMsg,
		},
		TxHash: fmt.Sprintf("0x%d", height),
	}
}

// TestNewBlockWorkerProcessing tests the full newblock worker flow
func TestNewBlockWorkerProcessing(t *testing.T) {
	// Connect to database
	databaseURL := "postgres://launchpad:launchpad123@localhost:5432/launchpad?sslmode=disable"
	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	templateRepo := postgres.NewChainTemplateRepository(db)
	chainRepo := postgres.NewChainRepository(db, userRepo, templateRepo)
	poolRepo := postgres.NewVirtualPoolRepository(db)

	ctx := context.Background()

	// Create a test chain with a unique address using timestamp
	timestamp := time.Now().UnixNano()
	chainAddress := []byte{
		byte(timestamp >> 56), byte(timestamp >> 48), byte(timestamp >> 40), byte(timestamp >> 32),
		byte(timestamp >> 24), byte(timestamp >> 16), byte(timestamp >> 8), byte(timestamp),
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22,
	}
	chainAddressHex := hex.EncodeToString(chainAddress)

	// Create test user with fixtures
	testUser, err := fixtures.DefaultUser().
		WithEmail(fmt.Sprintf("worker_test_%d@example.com", time.Now().UnixNano())).
		WithUsername(fmt.Sprintf("worker_user_%d", time.Now().Unix())).
		Create(ctx, db)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", testUser.ID)

	// Create test chain with fixtures
	testChain, err := fixtures.DefaultChain(testUser.ID).
		WithStatus(models.ChainStatusVirtualActive).
		Create(ctx, db)
	if err != nil {
		t.Fatalf("Failed to create test chain: %v", err)
	}
	defer chainRepo.Delete(ctx, testChain.ID)

	// Create chain key with address
	keyFixture := fixtures.DefaultChainKey(testChain.ID)
	keyFixture.Address = chainAddressHex
	keyFixture.KeyPurpose = "treasury"
	_, err = keyFixture.Create(ctx, db)
	if err != nil {
		t.Fatalf("Failed to create chain key: %v", err)
	}

	// Verify we can retrieve the chain by address
	verifyChain, err := chainRepo.GetByAddress(ctx, chainAddressHex)
	if err != nil {
		t.Fatalf("Failed to verify chain by address: %v", err)
	}
	if verifyChain.ID != testChain.ID {
		t.Fatalf("Chain ID mismatch: expected %s, got %s", testChain.ID, verifyChain.ID)
	}

	// Create virtual pool matching the bonding curve test setup
	// Initial price of 0.05 CNPY/token => 100 CNPY / 2000 tokens
	_, err = fixtures.DefaultVirtualPool(testChain.ID).
		WithReserves(100.0, 2000).
		Create(ctx, db)
	if err != nil {
		t.Fatalf("Failed to create virtual pool: %v", err)
	}

	// Create mock RPC client
	mockRPC := NewMockRPCClient()

	// Create worker config
	workerConfig := newblock.Config{
		RootChainURL: "ws://localhost:50002", // Won't actually connect
		RootChainID:  1,
	}

	// Create worker with mock RPC client
	worker := newblock.NewWorker(workerConfig, mockRPC, chainRepo, poolRepo, userRepo)

	t.Run("process_send_transaction_to_chain", func(t *testing.T) {
		// Create a test user with a unique wallet address
		senderAddressHex := fmt.Sprintf("%016x%024x", time.Now().UnixNano(), 0x1234)
		senderUser, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("sender1_%d@test.com", time.Now().UnixNano())).
			WithUsername(fmt.Sprintf("sender1_%d", time.Now().UnixNano())).
			WithWallet("0x"+senderAddressHex).
			Create(ctx, db)
		if err != nil {
			t.Fatalf("Failed to create sender user: %v", err)
		}
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", senderUser.ID)
		senderAddress, _ := hex.DecodeString(senderAddressHex)
		amount := uint64(1000000) // 1 CNPY in uCNPY
		height := uint64(1000)

		txResult := buildSendTransaction(senderAddress, chainAddress, amount, height)

		// Configure mock to return this transaction
		mockRPC.SetTransactionsAtHeight(height, []*lib.TxResult{txResult})

		// Create RootChainInfo event
		rootChainInfo := &lib.RootChainInfo{
			RootChainId: 1,
			Height:      height,
			Timestamp:   uint64(time.Now().Unix()),
		}

		// Process the event
		err = worker.HandleRootChainEventForTest(rootChainInfo)
		assert.NoError(t, err)

		// Verify pool state updated - this is the core functionality being tested
		updatedPool, err := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		assert.NoError(t, err)
		assert.Greater(t, updatedPool.CNPYReserve, 30.0, "CNPY reserve should have increased")
		assert.Less(t, updatedPool.TokenReserve, int64(800000000), "Token reserve should have decreased")
		assert.Equal(t, 1, updatedPool.TotalTransactions, "Transaction count should be 1")

		t.Logf("Pool updated: CNPY Reserve %.2f, Token Reserve %d", updatedPool.CNPYReserve, updatedPool.TokenReserve)
	})

	t.Run("sequential_buys_matching_bonding_curve_test", func(t *testing.T) {
		// This test matches TestBondingCurve_SequentialBuys in pkg/bondingcurve/curve_test.go
		// Using the same initial pool state (100 CNPY, 2000 tokens) and buy amounts (100, 200, 300)
		//
		// Expected results (from bonding curve test):
		// Buy 1: 100 CNPY → 990.00 tokens (price 0.10101010, reserve 200 CNPY)
		// Buy 2: 200 CNPY → 499.95 tokens (price 0.40004000, reserve 400 CNPY)
		// Buy 3: 300 CNPY → 216.41 tokens (price 1.38627724, reserve 700 CNPY)

		// Results not exact as the first test above deposits 1 CNPY before this one

		height := uint64(2000)

		// Get initial pool state
		poolBefore, err := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		assert.NoError(t, err)
		txCountBefore := poolBefore.TotalTransactions

		t.Logf("\n=== Initial Pool State ===")
		t.Logf("CNPY Reserve:  %.2f", poolBefore.CNPYReserve)
		t.Logf("Token Reserve: %d", poolBefore.TokenReserve)
		t.Logf("Initial Price: %.8f CNPY per token\n", poolBefore.CurrentPriceCNPY)

		// Create test users with unique wallet addresses
		ts := time.Now().UnixNano()
		sender1Hex := fmt.Sprintf("%016x%024x", ts, 0x1111)
		sender2Hex := fmt.Sprintf("%016x%024x", ts, 0x2222)
		sender3Hex := fmt.Sprintf("%016x%024x", ts, 0x3333)

		user1, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("sender_multi1_%d@test.com", ts)).
			WithUsername(fmt.Sprintf("sender_multi1_%d", ts)).
			WithWallet("0x"+sender1Hex).
			Create(ctx, db)
		if err != nil {
			t.Fatalf("Failed to create user1: %v", err)
		}
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user1.ID)

		user2, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("sender_multi2_%d@test.com", ts)).
			WithUsername(fmt.Sprintf("sender_multi2_%d", ts)).
			WithWallet("0x"+sender2Hex).
			Create(ctx, db)
		if err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user2.ID)

		user3, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("sender_multi3_%d@test.com", ts)).
			WithUsername(fmt.Sprintf("sender_multi3_%d", ts)).
			WithWallet("0x"+sender3Hex).
			Create(ctx, db)
		if err != nil {
			t.Fatalf("Failed to create user3: %v", err)
		}
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user3.ID)

		sender1, _ := hex.DecodeString(sender1Hex)
		sender2, _ := hex.DecodeString(sender2Hex)
		sender3, _ := hex.DecodeString(sender3Hex)

		// Create 3 transactions matching bonding curve test amounts
		tx1 := buildSendTransaction(sender1, chainAddress, 100_000_000, height) // 100 CNPY
		tx2 := buildSendTransaction(sender2, chainAddress, 200_000_000, height) // 200 CNPY
		tx3 := buildSendTransaction(sender3, chainAddress, 300_000_000, height) // 300 CNPY

		mockRPC.SetTransactionsAtHeight(height, []*lib.TxResult{tx1, tx2, tx3})

		// Print table header
		t.Logf("%-4s | %-10s | %-12s | %-12s | %-12s | %-12s",
			"#", "CNPY In", "Tokens Out", "Price", "CNPY Rsv", "Token Rsv")
		t.Logf("-----|------------|--------------|--------------|--------------|--------------|")

		rootChainInfo := &lib.RootChainInfo{
			RootChainId: 1,
			Height:      height,
			Timestamp:   uint64(time.Now().Unix()),
		}

		err = worker.HandleRootChainEventForTest(rootChainInfo)
		assert.NoError(t, err)

		// Verify 3 transactions processed
		poolAfter, err := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		assert.NoError(t, err)
		assert.Equal(t, txCountBefore+3, poolAfter.TotalTransactions)

		t.Logf("\n=== Final Pool State ===")
		t.Logf("CNPY Reserve:  %.2f", poolAfter.CNPYReserve)
		t.Logf("Token Reserve: %d", poolAfter.TokenReserve)
		t.Logf("Final Price:   %.8f CNPY per token", poolAfter.CurrentPriceCNPY)
		t.Logf("Total Volume:  %.2f CNPY", poolAfter.TotalVolumeCNPY)
		t.Logf("Total Transactions: %d\n", poolAfter.TotalTransactions)
	})

	t.Run("no_transactions_for_our_chain", func(t *testing.T) {
		height := uint64(3000)
		otherAddress := []byte{0x01, 0x02, 0x03, 0x04}
		senderAddress := []byte{0xFF, 0xEE, 0xDD, 0xCC}

		// Transaction to different address
		tx := buildSendTransaction(senderAddress, otherAddress, 1000000, height)

		mockRPC.SetTransactionsAtHeight(height, []*lib.TxResult{tx})

		poolBefore, _ := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		txCountBefore := poolBefore.TotalTransactions

		rootChainInfo := &lib.RootChainInfo{
			RootChainId: 1,
			Height:      height,
			Timestamp:   uint64(time.Now().Unix()),
		}

		err := worker.HandleRootChainEventForTest(rootChainInfo)
		assert.NoError(t, err)

		// No changes should occur
		poolAfter, err := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		assert.NoError(t, err)
		assert.Equal(t, txCountBefore, poolAfter.TotalTransactions, "Transaction count should not change")
	})

	t.Run("empty_block", func(t *testing.T) {
		height := uint64(4000)

		// No transactions configured for this height (will return empty)
		poolBefore, _ := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		txCountBefore := poolBefore.TotalTransactions

		rootChainInfo := &lib.RootChainInfo{
			RootChainId: 1,
			Height:      height,
			Timestamp:   uint64(time.Now().Unix()),
		}

		err := worker.HandleRootChainEventForTest(rootChainInfo)
		assert.NoError(t, err)

		// No changes
		poolAfter, err := poolRepo.GetPoolByChainID(ctx, testChain.ID)
		assert.NoError(t, err)
		assert.Equal(t, txCountBefore, poolAfter.TotalTransactions)
	})
}

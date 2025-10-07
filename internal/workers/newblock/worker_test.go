package newblock

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/canopy-network/canopy/fsm"
	"github.com/canopy-network/canopy/lib"
	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/anypb"
)

// MockRPCClient mocks the RPCClient interface
type MockRPCClient struct {
	mock.Mock
}

func (m *MockRPCClient) TransactionsByHeight(height uint64, page lib.PageParams) (*lib.Page, lib.ErrorI) {
	args := m.Called(height, page)
	if args.Get(0) == nil {
		if args.Get(1) == nil {
			return nil, nil
		}
		return nil, args.Get(1).(lib.ErrorI)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*lib.Page), nil
	}
	return args.Get(0).(*lib.Page), args.Get(1).(lib.ErrorI)
}

// MockChainRepository mocks the ChainRepository interface
type MockChainRepository struct {
	mock.Mock
}

func (m *MockChainRepository) GetByAddress(ctx context.Context, address string) (*models.Chain, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chain), args.Error(1)
}

func (m *MockChainRepository) Create(ctx context.Context, chain *models.Chain) (*models.Chain, error) {
	args := m.Called(ctx, chain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chain), args.Error(1)
}

func (m *MockChainRepository) GetByID(ctx context.Context, id uuid.UUID, includeRelations []string) (*models.Chain, error) {
	args := m.Called(ctx, id, includeRelations)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chain), args.Error(1)
}

func (m *MockChainRepository) GetByName(ctx context.Context, name string) (*models.Chain, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chain), args.Error(1)
}

func (m *MockChainRepository) Update(ctx context.Context, chain *models.Chain) (*models.Chain, error) {
	args := m.Called(ctx, chain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chain), args.Error(1)
}

func (m *MockChainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChainRepository) List(ctx context.Context, filters interfaces.ChainFilters, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	args := m.Called(ctx, filters, pagination)
	return args.Get(0).([]models.Chain), args.Int(1), args.Error(2)
}

func (m *MockChainRepository) ListByCreator(ctx context.Context, creatorID uuid.UUID, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	args := m.Called(ctx, creatorID, pagination)
	return args.Get(0).([]models.Chain), args.Int(1), args.Error(2)
}

func (m *MockChainRepository) ListByTemplate(ctx context.Context, templateID uuid.UUID, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	args := m.Called(ctx, templateID, pagination)
	return args.Get(0).([]models.Chain), args.Int(1), args.Error(2)
}

func (m *MockChainRepository) ListByStatus(ctx context.Context, status string, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	args := m.Called(ctx, status, pagination)
	return args.Get(0).([]models.Chain), args.Int(1), args.Error(2)
}

func (m *MockChainRepository) CreateRepository(ctx context.Context, repo *models.ChainRepository) (*models.ChainRepository, error) {
	args := m.Called(ctx, repo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChainRepository), args.Error(1)
}

func (m *MockChainRepository) UpdateRepository(ctx context.Context, repo *models.ChainRepository) (*models.ChainRepository, error) {
	args := m.Called(ctx, repo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChainRepository), args.Error(1)
}

func (m *MockChainRepository) GetRepositoryByChainID(ctx context.Context, chainID uuid.UUID) (*models.ChainRepository, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChainRepository), args.Error(1)
}

func (m *MockChainRepository) DeleteRepository(ctx context.Context, chainID uuid.UUID) error {
	args := m.Called(ctx, chainID)
	return args.Error(0)
}

func (m *MockChainRepository) CreateSocialLinks(ctx context.Context, chainID uuid.UUID, links []models.ChainSocialLink) error {
	args := m.Called(ctx, chainID, links)
	return args.Error(0)
}

func (m *MockChainRepository) UpdateSocialLinks(ctx context.Context, chainID uuid.UUID, links []models.ChainSocialLink) error {
	args := m.Called(ctx, chainID, links)
	return args.Error(0)
}

func (m *MockChainRepository) GetSocialLinksByChainID(ctx context.Context, chainID uuid.UUID) ([]models.ChainSocialLink, error) {
	args := m.Called(ctx, chainID)
	return args.Get(0).([]models.ChainSocialLink), args.Error(1)
}

func (m *MockChainRepository) DeleteSocialLinksByChainID(ctx context.Context, chainID uuid.UUID) error {
	args := m.Called(ctx, chainID)
	return args.Error(0)
}

func (m *MockChainRepository) CreateAssets(ctx context.Context, chainID uuid.UUID, assets []models.ChainAsset) error {
	args := m.Called(ctx, chainID, assets)
	return args.Error(0)
}

func (m *MockChainRepository) UpdateAssets(ctx context.Context, chainID uuid.UUID, assets []models.ChainAsset) error {
	args := m.Called(ctx, chainID, assets)
	return args.Error(0)
}

func (m *MockChainRepository) GetAssetsByChainID(ctx context.Context, chainID uuid.UUID) ([]models.ChainAsset, error) {
	args := m.Called(ctx, chainID)
	return args.Get(0).([]models.ChainAsset), args.Error(1)
}

func (m *MockChainRepository) DeleteAssetsByChainID(ctx context.Context, chainID uuid.UUID) error {
	args := m.Called(ctx, chainID)
	return args.Error(0)
}

func (m *MockChainRepository) GetVirtualPoolByChainID(ctx context.Context, chainID uuid.UUID) (*models.VirtualPool, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VirtualPool), args.Error(1)
}

func (m *MockChainRepository) GetTransactionsByChainID(ctx context.Context, chainID uuid.UUID, filters interfaces.TransactionFilters, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	args := m.Called(ctx, chainID, filters, pagination)
	return args.Get(0).([]models.VirtualPoolTransaction), args.Int(1), args.Error(2)
}

func (m *MockChainRepository) CreateChainKey(ctx context.Context, key *models.ChainKey) (*models.ChainKey, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChainKey), args.Error(1)
}

func (m *MockChainRepository) GetChainKeyByChainID(ctx context.Context, chainID uuid.UUID, purpose string) (*models.ChainKey, error) {
	args := m.Called(ctx, chainID, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChainKey), args.Error(1)
}

// MockVirtualPoolRepository mocks the VirtualPoolRepository interface
type MockVirtualPoolRepository struct {
	mock.Mock
}

func (m *MockVirtualPoolRepository) GetPoolByChainID(ctx context.Context, chainID uuid.UUID) (*models.VirtualPool, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VirtualPool), args.Error(1)
}

func (m *MockVirtualPoolRepository) UpdatePoolState(ctx context.Context, chainID uuid.UUID, update *interfaces.PoolStateUpdate) error {
	args := m.Called(ctx, chainID, update)
	return args.Error(0)
}

func (m *MockVirtualPoolRepository) CreateTransaction(ctx context.Context, transaction *models.VirtualPoolTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockVirtualPoolRepository) GetTransactionsByPoolID(ctx context.Context, poolID uuid.UUID, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	args := m.Called(ctx, poolID, pagination)
	return args.Get(0).([]models.VirtualPoolTransaction), args.Int(1), args.Error(2)
}

func (m *MockVirtualPoolRepository) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	args := m.Called(ctx, userID, pagination)
	return args.Get(0).([]models.VirtualPoolTransaction), args.Int(1), args.Error(2)
}

func (m *MockVirtualPoolRepository) GetUserPosition(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	args := m.Called(ctx, userID, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserVirtualPosition), args.Error(1)
}

func (m *MockVirtualPoolRepository) UpsertUserPosition(ctx context.Context, position *models.UserVirtualPosition) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockVirtualPoolRepository) GetPositionsByChainID(ctx context.Context, chainID uuid.UUID, pagination interfaces.Pagination) ([]models.UserVirtualPosition, int, error) {
	args := m.Called(ctx, chainID, pagination)
	return args.Get(0).([]models.UserVirtualPosition), args.Int(1), args.Error(2)
}

// MockUserRepository mocks the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*models.User, error) {
	args := m.Called(ctx, walletAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, filters interfaces.UserFilters, pagination interfaces.Pagination) ([]models.User, int, error) {
	args := m.Called(ctx, filters, pagination)
	return args.Get(0).([]models.User), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) UpdateActivity(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetPositionByUserAndChain(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	args := m.Called(ctx, userID, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserVirtualPosition), args.Error(1)
}

func (m *MockUserRepository) UpdatePosition(ctx context.Context, position *models.UserVirtualPosition) (*models.UserVirtualPosition, error) {
	args := m.Called(ctx, position)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserVirtualPosition), args.Error(1)
}

func (m *MockUserRepository) ListByVerificationTier(ctx context.Context, tier string, pagination interfaces.Pagination) ([]models.User, int, error) {
	args := m.Called(ctx, tier, pagination)
	return args.Get(0).([]models.User), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) GetPositionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserVirtualPosition, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserVirtualPosition), args.Error(1)
}

func (m *MockUserRepository) UpdateChainsCreatedCount(ctx context.Context, userID uuid.UUID, increment int) error {
	args := m.Called(ctx, userID, increment)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateCNPYInvestment(ctx context.Context, userID uuid.UUID, amount float64) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateReputationScore(ctx context.Context, userID uuid.UUID, score int) error {
	args := m.Called(ctx, userID, score)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Helper functions to create test fixtures

// setupStandardUserMocks sets up the standard user repository mocks for transaction processing
func setupStandardUserMocks(userRepo *MockUserRepository, poolRepo *MockVirtualPoolRepository, senderAddress []byte, chainID uuid.UUID) {
	senderAddressHex := hex.EncodeToString(senderAddress)
	testUser := &models.User{
		ID:            uuid.New(),
		WalletAddress: senderAddressHex,
		IsVerified:    false,
	}
	userRepo.On("GetByWalletAddress", mock.Anything, senderAddressHex).Return(testUser, nil)

	// Mock CreateTransaction to be called
	poolRepo.On("CreateTransaction", mock.Anything, mock.MatchedBy(func(tx *models.VirtualPoolTransaction) bool {
		return tx.TransactionType == "buy" && tx.ChainID == chainID
	})).Return(nil)

	// Mock GetUserPosition (returning nil indicates no existing position)
	poolRepo.On("GetUserPosition", mock.Anything, mock.Anything, chainID).Return(nil, nil)

	// Mock UpsertUserPosition to be called
	poolRepo.On("UpsertUserPosition", mock.Anything, mock.MatchedBy(func(pos *models.UserVirtualPosition) bool {
		return pos.UserID == testUser.ID && pos.ChainID == chainID
	})).Return(nil)
}

// buildTxResultWithValidSend creates a TxResult with a valid send transaction
func buildTxResultWithValidSend(recipientAddress []byte, senderAddress []byte, amount uint64) *lib.TxResult {
	// Create a MessageSend
	sendMsg := &fsm.MessageSend{
		FromAddress: senderAddress,
		ToAddress:   recipientAddress,
		Amount:      amount,
	}

	// Marshal it to Any
	anyMsg, err := anypb.New(sendMsg)
	if err != nil {
		panic(fmt.Sprintf("failed to create any message: %v", err))
	}

	return &lib.TxResult{
		Sender:      senderAddress,
		Recipient:   recipientAddress,
		MessageType: fsm.MessageSendName,
		Height:      1000,
		Index:       0,
		Transaction: &lib.Transaction{
			MessageType: fsm.MessageSendName,
			Msg:         anyMsg,
		},
		TxHash: "0xabc123",
	}
}

// buildTxResultWithNilTransaction creates a TxResult with nil transaction
func buildTxResultWithNilTransaction(recipientAddress []byte, senderAddress []byte) *lib.TxResult {
	return &lib.TxResult{
		Sender:      senderAddress,
		Recipient:   recipientAddress,
		MessageType: fsm.MessageSendName,
		Height:      1000,
		Index:       0,
		Transaction: nil,
		TxHash:      "0xabc123",
	}
}

// buildTxResultWithNilMessage creates a TxResult with nil message in transaction
func buildTxResultWithNilMessage(recipientAddress []byte, senderAddress []byte) *lib.TxResult {
	return &lib.TxResult{
		Sender:      senderAddress,
		Recipient:   recipientAddress,
		MessageType: fsm.MessageSendName,
		Height:      1000,
		Index:       0,
		Transaction: &lib.Transaction{
			MessageType: fsm.MessageSendName,
			Msg:         nil,
		},
		TxHash: "0xabc123",
	}
}

// buildChain creates a test chain with specified parameters
func buildChain(chainID uuid.UUID, chainName string, creatorID uuid.UUID) *models.Chain {
	return &models.Chain{
		ID:                   chainID,
		ChainName:            chainName,
		TokenSymbol:          "TEST",
		ConsensusMechanism:   "PoS",
		TokenTotalSupply:     1000000000,
		GraduationThreshold:  100000.0,
		CreationFeeCNPY:      10.0,
		InitialCNPYReserve:   30.0,
		InitialTokenSupply:   800000000,
		BondingCurveSlope:    0.5,
		Status:               models.ChainStatusVirtualActive,
		IsGraduated:          false,
		CreatedBy:            creatorID,
		ValidatorMinStake:    1000.0,
	}
}

// buildVirtualPool creates a test virtual pool with specified reserves
func buildVirtualPool(poolID uuid.UUID, chainID uuid.UUID, cnpyReserve float64, tokenReserve int64, totalTransactions int) *models.VirtualPool {
	return &models.VirtualPool{
		ID:                    poolID,
		ChainID:               chainID,
		CNPYReserve:           cnpyReserve,
		TokenReserve:          tokenReserve,
		CurrentPriceCNPY:      cnpyReserve / float64(tokenReserve),
		MarketCapUSD:          0,
		TotalVolumeCNPY:       0,
		TotalTransactions:     totalTransactions,
		UniqueTraders:         0,
		IsActive:              true,
		Price24hChangePercent: 0,
		Volume24hCNPY:         0,
		High24hCNPY:           0,
		Low24hCNPY:            0,
	}
}

func TestWorker_processTransaction(t *testing.T) {
	// Common test fixtures
	chainID := uuid.New()
	creatorID := uuid.New()
	poolID := uuid.New()
	recipientAddress := []byte{0x01, 0x02, 0x03, 0x04}
	recipientAddressHex := hex.EncodeToString(recipientAddress)
	senderAddress := []byte{0x05, 0x06, 0x07, 0x08}

	tests := []struct {
		name           string
		txResult       *lib.TxResult
		index          int
		total          int
		height         uint64
		setupMocks     func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository)
		expectedLogs   []string // Expected log patterns (for manual verification)
		activeForm     string
	}{
		{
			name:     "successful transaction processing with valid send",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 1000000), // 1 CNPY in uCNPY
			index:    0,
			total:    1,
			height:   1000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "TestChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				// Expect UpdatePoolState to be called with increased reserves
				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					// Verify that reserves have increased
					return update.CNPYReserve != nil &&
						update.TokenReserve != nil &&
						update.CurrentPriceCNPY != nil &&
						update.TotalTransactions != nil &&
						*update.TotalTransactions == 1
				})).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Transaction 1/1 at height 1000",
				"Chain=TestChain",
				"Successfully processed deposit",
			},
			activeForm: "Processing successful transaction with valid send",
		},
		{
			name:     "chain not found for recipient address",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 1000000),
			index:    0,
			total:    1,
			height:   1000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				// Chain not found error
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(nil, fmt.Errorf("chain not found"))
			},
			expectedLogs: []string{
				"No chain found for recipient address",
			},
			activeForm: "Processing transaction when chain not found",
		},
		{
			name:     "nil transaction in TxResult",
			txResult: buildTxResultWithNilTransaction(recipientAddress, senderAddress),
			index:    1,
			total:    3,
			height:   2000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "TestChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)
			},
			expectedLogs: []string{
				"Failed to extract send amount",
				"transaction or message is nil",
			},
			activeForm: "Processing transaction with nil transaction object",
		},
		{
			name:     "nil message in transaction",
			txResult: buildTxResultWithNilMessage(recipientAddress, senderAddress),
			index:    0,
			total:    1,
			height:   3000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "TestChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)
			},
			expectedLogs: []string{
				"Failed to extract send amount",
				"transaction or message is nil",
			},
			activeForm: "Processing transaction with nil message",
		},
		{
			name:     "pool not found for chain",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 5000000), // 5 CNPY
			index:    0,
			total:    1,
			height:   4000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "TestChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				// Pool not found
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(nil, fmt.Errorf("pool not found"))
			},
			expectedLogs: []string{
				"Failed to process deposit",
				"failed to get virtual pool",
			},
			activeForm: "Processing transaction when pool not found",
		},
		{
			name:     "pool state update fails",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 2000000), // 2 CNPY
			index:    0,
			total:    1,
			height:   5000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "TestChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				// UpdatePoolState fails
				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.Anything).Return(fmt.Errorf("database connection lost"))
			},
			expectedLogs: []string{
				"Failed to process deposit",
				"failed to update pool state",
			},
			activeForm: "Processing transaction when pool state update fails",
		},
		{
			name:     "large amount transaction (100 CNPY)",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 100000000), // 100 CNPY in uCNPY
			index:    5,
			total:    10,
			height:   6000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "LargeChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				// Pool with larger reserves to handle large transaction
				pool := buildVirtualPool(poolID, chainID, 1000.0, 8000000000, 50)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					// For large deposits, CNPY reserve should increase significantly
					return update.CNPYReserve != nil &&
						update.CNPYReserve.Cmp(big.NewFloat(1000.0)) > 0 && // Greater than original reserve
						update.TotalTransactions != nil &&
						*update.TotalTransactions == 51
				})).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Transaction 6/10 at height 6000",
				"Successfully processed deposit",
			},
			activeForm: "Processing large amount transaction (100 CNPY)",
		},
		{
			name:     "zero amount transaction - bonding curve rejects",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 0), // 0 CNPY
			index:    0,
			total:    1,
			height:   7000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "ZeroChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				// Note: UpdatePoolState should NOT be called because bonding curve rejects zero amounts
			},
			expectedLogs: []string{
				"Processing deposit: Chain=ZeroChain, Amount=0 uCNPY",
				"Failed to process deposit",
				"bonding curve buy failed",
			},
			activeForm: "Processing zero amount transaction (bonding curve rejects)",
		},
		{
			name:     "context already cancelled before processing",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 1000000),
			index:    0,
			total:    1,
			height:   8000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				// GetByAddress should be called but return context cancelled
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(nil, context.Canceled)
			},
			expectedLogs: []string{
				"No chain found for recipient address",
			},
			activeForm: "Processing with cancelled context",
		},
		{
			name:     "multiple transactions same block - first of three",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 500000), // 0.5 CNPY
			index:    0,
			total:    3,
			height:   9000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "MultiTxChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 100)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					return update.TotalTransactions != nil && *update.TotalTransactions == 101
				})).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Transaction 1/3 at height 9000",
			},
			activeForm: "Processing first of multiple transactions in block",
		},
		{
			name:     "multiple transactions same block - second of three",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 750000), // 0.75 CNPY
			index:    1,
			total:    3,
			height:   9000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "MultiTxChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				// Pool state after first transaction
				pool := buildVirtualPool(poolID, chainID, 30.5, 799000000, 101)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					return update.TotalTransactions != nil && *update.TotalTransactions == 102
				})).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Transaction 2/3 at height 9000",
			},
			activeForm: "Processing second of multiple transactions in block",
		},
		{
			name:     "multiple transactions same block - last of three",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 250000), // 0.25 CNPY
			index:    2,
			total:    3,
			height:   9000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "MultiTxChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				// Pool state after first two transactions
				pool := buildVirtualPool(poolID, chainID, 31.25, 798000000, 102)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					return update.TotalTransactions != nil && *update.TotalTransactions == 103
				})).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Transaction 3/3 at height 9000",
			},
			activeForm: "Processing last of multiple transactions in block",
		},
		{
			name:     "transaction with very small reserve pool (near graduation)",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 10000000), // 10 CNPY
			index:    0,
			total:    1,
			height:   10000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "NearGraduationChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				// Pool with small token reserve (most tokens sold)
				pool := buildVirtualPool(poolID, chainID, 95000.0, 5000000, 1000) // Very low token reserve
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.Anything).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Successfully processed deposit",
			},
			activeForm: "Processing transaction with small reserve pool",
		},
		{
			name:     "transaction processing with fresh pool (first transaction)",
			txResult: buildTxResultWithValidSend(recipientAddress, senderAddress, 1000000), // 1 CNPY
			index:    0,
			total:    1,
			height:   11000,
			setupMocks: func(chainRepo *MockChainRepository, poolRepo *MockVirtualPoolRepository, userRepo *MockUserRepository) {
				chain := buildChain(chainID, "FreshChain", creatorID)
				chainRepo.On("GetByAddress", mock.Anything, recipientAddressHex).Return(chain, nil)

				// Fresh pool with initial reserves, zero transactions
				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					// First transaction should increment from 0 to 1
					return update.TotalTransactions != nil && *update.TotalTransactions == 1
				})).Return(nil)

				// Setup user mocks for transaction recording
				setupStandardUserMocks(userRepo, poolRepo, senderAddress, chainID)
			},
			expectedLogs: []string{
				"Successfully processed deposit",
			},
			activeForm: "Processing first transaction on fresh pool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repositories
			chainRepo := new(MockChainRepository)
			poolRepo := new(MockVirtualPoolRepository)
			userRepo := new(MockUserRepository)

			// Setup mocks for this test case
			tt.setupMocks(chainRepo, poolRepo, userRepo)

			// Create worker with mocks
			worker := &Worker{
				chainRepo: chainRepo,
				poolRepo:  poolRepo,
				userRepo:  userRepo,
				logger:    NewLogger(),
			}

			// Create context (cancelled for context cancellation test)
			ctx := context.Background()
			if tt.name == "context already cancelled before processing" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel() // Cancel immediately
			}

			// Execute the function under test
			worker.processTransaction(ctx, tt.txResult, tt.index, tt.total, tt.height)

			// Verify all expected mock calls were made
			chainRepo.AssertExpectations(t)
			poolRepo.AssertExpectations(t)
		})
	}
}

func TestWorker_extractSendAmount(t *testing.T) {
	senderAddress := []byte{0x01, 0x02}
	recipientAddress := []byte{0x03, 0x04}

	tests := []struct {
		name          string
		txResult      *lib.TxResult
		expectedAmt   uint64
		expectError   bool
		errorContains string
		activeForm    string
	}{
		{
			name:        "valid send transaction with standard amount",
			txResult:    buildTxResultWithValidSend(recipientAddress, senderAddress, 1000000),
			expectedAmt: 1000000,
			expectError: false,
			activeForm:  "Extracting amount from valid send transaction",
		},
		{
			name:        "valid send transaction with large amount",
			txResult:    buildTxResultWithValidSend(recipientAddress, senderAddress, 1000000000000),
			expectedAmt: 1000000000000,
			expectError: false,
			activeForm:  "Extracting large amount from send transaction",
		},
		{
			name:        "valid send transaction with zero amount",
			txResult:    buildTxResultWithValidSend(recipientAddress, senderAddress, 0),
			expectedAmt: 0,
			expectError: false,
			activeForm:  "Extracting zero amount from send transaction",
		},
		{
			name:          "nil transaction",
			txResult:      buildTxResultWithNilTransaction(recipientAddress, senderAddress),
			expectedAmt:   0,
			expectError:   true,
			errorContains: "transaction or message is nil",
			activeForm:    "Extracting amount from nil transaction",
		},
		{
			name:          "nil message",
			txResult:      buildTxResultWithNilMessage(recipientAddress, senderAddress),
			expectedAmt:   0,
			expectError:   true,
			errorContains: "transaction or message is nil",
			activeForm:    "Extracting amount from nil message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := &Worker{
				logger: NewLogger(),
			}

			amount, err := worker.extractSendAmount(tt.txResult)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmt, amount)
			}
		})
	}
}

func TestWorker_processDeposit(t *testing.T) {
	chainID := uuid.New()
	creatorID := uuid.New()
	poolID := uuid.New()
	senderAddress := []byte{0x01, 0x02, 0x03, 0x04}

	tests := []struct {
		name          string
		chain         *models.Chain
		amount        uint64
		sender        []byte
		setupMocks    func(poolRepo *MockVirtualPoolRepository)
		expectError   bool
		errorContains string
		activeForm    string
	}{
		{
			name:   "successful deposit with standard amount",
			chain:  buildChain(chainID, "TestChain", creatorID),
			amount: 1000000, // 1 CNPY
			sender: senderAddress,
			setupMocks: func(poolRepo *MockVirtualPoolRepository) {
				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)
				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.Anything).Return(nil)
			},
			expectError: false,
			activeForm:  "Processing successful deposit",
		},
		{
			name:   "pool not found for chain",
			chain:  buildChain(chainID, "TestChain", creatorID),
			amount: 1000000,
			sender: senderAddress,
			setupMocks: func(poolRepo *MockVirtualPoolRepository) {
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(nil, fmt.Errorf("pool not found"))
			},
			expectError:   true,
			errorContains: "failed to get virtual pool",
			activeForm:    "Processing deposit when pool not found",
		},
		{
			name:   "pool state update fails",
			chain:  buildChain(chainID, "TestChain", creatorID),
			amount: 1000000,
			sender: senderAddress,
			setupMocks: func(poolRepo *MockVirtualPoolRepository) {
				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)
				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.Anything).Return(fmt.Errorf("database error"))
			},
			expectError:   true,
			errorContains: "failed to update pool state",
			activeForm:    "Processing deposit when update fails",
		},
		{
			name:   "large amount deposit (100 CNPY)",
			chain:  buildChain(chainID, "LargeChain", creatorID),
			amount: 100000000, // 100 CNPY
			sender: senderAddress,
			setupMocks: func(poolRepo *MockVirtualPoolRepository) {
				pool := buildVirtualPool(poolID, chainID, 1000.0, 8000000000, 50)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)
				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.Anything).Return(nil)
			},
			expectError: false,
			activeForm:  "Processing large amount deposit",
		},
		{
			name:   "zero amount deposit - bonding curve rejects",
			chain:  buildChain(chainID, "ZeroChain", creatorID),
			amount: 0,
			sender: senderAddress,
			setupMocks: func(poolRepo *MockVirtualPoolRepository) {
				pool := buildVirtualPool(poolID, chainID, 30.0, 800000000, 0)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)
				// UpdatePoolState should NOT be called - bonding curve rejects zero amounts
			},
			expectError:   true,
			errorContains: "bonding curve buy failed",
			activeForm:    "Processing zero amount deposit (bonding curve rejects)",
		},
		{
			name:   "deposit to pool with many transactions",
			chain:  buildChain(chainID, "PopularChain", creatorID),
			amount: 5000000, // 5 CNPY
			sender: senderAddress,
			setupMocks: func(poolRepo *MockVirtualPoolRepository) {
				pool := buildVirtualPool(poolID, chainID, 5000.0, 400000000, 9999)
				poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)
				poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.MatchedBy(func(update *interfaces.PoolStateUpdate) bool {
					return update.TotalTransactions != nil && *update.TotalTransactions == 10000
				})).Return(nil)
			},
			expectError: false,
			activeForm:  "Processing deposit to popular pool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolRepo := new(MockVirtualPoolRepository)
			userRepo := new(MockUserRepository)
			tt.setupMocks(poolRepo)

			// Setup user mocks for successful cases
			if !tt.expectError {
				setupStandardUserMocks(userRepo, poolRepo, tt.sender, tt.chain.ID)
			}

			worker := &Worker{
				poolRepo: poolRepo,
				userRepo: userRepo,
				logger:   NewLogger(),
			}

			err := worker.processDeposit(context.Background(), tt.chain, tt.amount, tt.sender)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			poolRepo.AssertExpectations(t)
		})
	}
}

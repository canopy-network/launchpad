package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/canopy-network/canopy/lib"
	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/bondingcurve"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVirtualPoolRepository is a mock implementation of VirtualPoolRepository
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

func (m *MockVirtualPoolRepository) Create(ctx context.Context, pool *models.VirtualPool) (*models.VirtualPool, error) {
	args := m.Called(ctx, pool)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VirtualPool), args.Error(1)
}

func (m *MockVirtualPoolRepository) GetAllPools(ctx context.Context, pagination interfaces.Pagination) ([]models.VirtualPool, int, error) {
	args := m.Called(ctx, pagination)
	return args.Get(0).([]models.VirtualPool), args.Int(1), args.Error(2)
}

func (m *MockVirtualPoolRepository) GetTransactionsByChainID(ctx context.Context, chainID uuid.UUID, filters interfaces.TransactionFilters, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	args := m.Called(ctx, chainID, filters, pagination)
	return args.Get(0).([]models.VirtualPoolTransaction), args.Int(1), args.Error(2)
}

func (m *MockVirtualPoolRepository) GetPositionsWithUsersByChainID(ctx context.Context, chainID uuid.UUID) ([]interfaces.UserPositionWithAddress, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interfaces.UserPositionWithAddress), args.Error(1)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
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
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
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

func (m *MockUserRepository) ListByVerificationTier(ctx context.Context, tier string, pagination interfaces.Pagination) ([]models.User, int, error) {
	args := m.Called(ctx, tier, pagination)
	return args.Get(0).([]models.User), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) GetPositionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserVirtualPosition, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.UserVirtualPosition), args.Error(1)
}

func (m *MockUserRepository) GetPositionByUserAndChain(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	args := m.Called(ctx, userID, chainID)
	return args.Get(0).(*models.UserVirtualPosition), args.Error(1)
}

func (m *MockUserRepository) UpdatePosition(ctx context.Context, position *models.UserVirtualPosition) (*models.UserVirtualPosition, error) {
	args := m.Called(ctx, position)
	return args.Get(0).(*models.UserVirtualPosition), args.Error(1)
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

func (m *MockUserRepository) CreateOrGetByEmail(ctx context.Context, email string) (*models.User, bool, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*models.User), args.Bool(1), args.Error(2)
}

func (m *MockUserRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementJWTVersion(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestNewOrderProcessor(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)

	t.Run("with custom config", func(t *testing.T) {
		config := &bondingcurve.BondingCurveConfig{
			FeeRateBasisPoints: 200, // 2%
		}
		processor := NewOrderProcessor(poolRepo, userRepo, config)
		assert.NotNil(t, processor)
		assert.Equal(t, uint64(200), processor.GetConfig().FeeRateBasisPoints)
	})

	t.Run("with nil config uses default", func(t *testing.T) {
		processor := NewOrderProcessor(poolRepo, userRepo, nil)
		assert.NotNil(t, processor)
		assert.Equal(t, uint64(100), processor.GetConfig().FeeRateBasisPoints) // Default 1%
	})
}

func TestValidateOrder(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)
	processor := NewOrderProcessor(poolRepo, userRepo, nil)

	t.Run("valid buy order", func(t *testing.T) {
		order := &lib.SellOrder{
			AmountForSale:        1000,
			RequestedAmount:      80000,
			BuyerReceiveAddress:  []byte(uuid.New().String()),
		}
		err := processor.validateOrder(order)
		assert.NoError(t, err)
	})

	t.Run("nil order", func(t *testing.T) {
		err := processor.validateOrder(nil)
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("zero amounts", func(t *testing.T) {
		order := &lib.SellOrder{
			AmountForSale:   0,
			RequestedAmount: 0,
		}
		err := processor.validateOrder(order)
		assert.ErrorIs(t, err, ErrZeroAmount)
	})

	t.Run("missing addresses", func(t *testing.T) {
		order := &lib.SellOrder{
			AmountForSale:   1000,
			RequestedAmount: 80000,
		}
		err := processor.validateOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing user address")
	})
}

func TestProcessBuyOrder(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)
	processor := NewOrderProcessor(poolRepo, userRepo, nil)

	chainID := uuid.New()
	userID := uuid.New()
	poolID := uuid.New()

	pool := &models.VirtualPool{
		ID:           poolID,
		ChainID:      chainID,
		CNPYReserve:  10000.0,
		TokenReserve: 800000000,
		CurrentPriceCNPY: 0.0000125,
		TotalVolumeCNPY: 5000.0,
		TotalTransactions: 10,
	}

	order := &lib.SellOrder{
		AmountForSale:       100,
		RequestedAmount:     8000,
		BuyerReceiveAddress: []byte(userID.String()),
	}

	t.Run("successful buy with new position", func(t *testing.T) {
		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()
		poolRepo.On("GetUserPosition", mock.Anything, userID, chainID).Return(nil, nil).Once()
		poolRepo.On("UpsertUserPosition", mock.Anything, mock.AnythingOfType("*models.UserVirtualPosition")).Return(nil).Once()
		poolRepo.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.VirtualPoolTransaction")).Return(nil).Once()
		poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.AnythingOfType("*interfaces.PoolStateUpdate")).Return(nil).Once()

		err := processor.processBuyOrder(context.Background(), order, chainID)
		assert.NoError(t, err)
		poolRepo.AssertExpectations(t)
	})

	t.Run("successful buy with existing position", func(t *testing.T) {
		now := time.Now()
		existingPosition := &models.UserVirtualPosition{
			ID:                    uuid.New(),
			UserID:                userID,
			ChainID:               chainID,
			VirtualPoolID:         poolID,
			TokenBalance:          5000,
			TotalCNPYInvested:     60.0,
			AverageEntryPriceCNPY: 0.012,
			FirstPurchaseAt:       &now,
		}

		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()
		poolRepo.On("GetUserPosition", mock.Anything, userID, chainID).Return(existingPosition, nil).Once()
		poolRepo.On("UpsertUserPosition", mock.Anything, mock.AnythingOfType("*models.UserVirtualPosition")).Return(nil).Once()
		poolRepo.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.VirtualPoolTransaction")).Return(nil).Once()
		poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.AnythingOfType("*interfaces.PoolStateUpdate")).Return(nil).Once()

		err := processor.processBuyOrder(context.Background(), order, chainID)
		assert.NoError(t, err)
		poolRepo.AssertExpectations(t)
	})

	t.Run("pool not found", func(t *testing.T) {
		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(nil, errors.New("not found")).Once()

		err := processor.processBuyOrder(context.Background(), order, chainID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPoolNotFound)
		poolRepo.AssertExpectations(t)
	})
}

func TestProcessSellOrder(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)
	processor := NewOrderProcessor(poolRepo, userRepo, nil)

	chainID := uuid.New()
	userID := uuid.New()
	poolID := uuid.New()

	pool := &models.VirtualPool{
		ID:           poolID,
		ChainID:      chainID,
		CNPYReserve:  10000.0,
		TokenReserve: 800000000,
		CurrentPriceCNPY: 0.0000125,
		TotalVolumeCNPY: 5000.0,
		TotalTransactions: 10,
	}

	now := time.Now()
	existingPosition := &models.UserVirtualPosition{
		ID:                    uuid.New(),
		UserID:                userID,
		ChainID:               chainID,
		VirtualPoolID:         poolID,
		TokenBalance:          10000,
		TotalCNPYInvested:     120.0,
		AverageEntryPriceCNPY: 0.012,
		FirstPurchaseAt:       &now,
	}

	order := &lib.SellOrder{
		RequestedAmount:    5000, // Selling 5000 tokens
		SellersSendAddress: []byte(userID.String()),
	}

	t.Run("successful sell", func(t *testing.T) {
		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()
		poolRepo.On("GetUserPosition", mock.Anything, userID, chainID).Return(existingPosition, nil).Once()
		poolRepo.On("UpsertUserPosition", mock.Anything, mock.AnythingOfType("*models.UserVirtualPosition")).Return(nil).Once()
		poolRepo.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*models.VirtualPoolTransaction")).Return(nil).Once()
		poolRepo.On("UpdatePoolState", mock.Anything, chainID, mock.AnythingOfType("*interfaces.PoolStateUpdate")).Return(nil).Once()

		err := processor.processSellOrder(context.Background(), order, chainID)
		assert.NoError(t, err)
		poolRepo.AssertExpectations(t)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		smallPosition := &models.UserVirtualPosition{
			ID:            uuid.New(),
			UserID:        userID,
			ChainID:       chainID,
			VirtualPoolID: poolID,
			TokenBalance:  100, // Less than order amount
		}

		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()
		poolRepo.On("GetUserPosition", mock.Anything, userID, chainID).Return(smallPosition, nil).Once()

		err := processor.processSellOrder(context.Background(), order, chainID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientBalance)
		poolRepo.AssertExpectations(t)
	})

	t.Run("position not found", func(t *testing.T) {
		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()
		poolRepo.On("GetUserPosition", mock.Anything, userID, chainID).Return(nil, nil).Once()

		err := processor.processSellOrder(context.Background(), order, chainID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientBalance)
		poolRepo.AssertExpectations(t)
	})
}

func TestSimulateBuy(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)
	processor := NewOrderProcessor(poolRepo, userRepo, nil)

	chainID := uuid.New()
	pool := &models.VirtualPool{
		ID:           uuid.New(),
		ChainID:      chainID,
		CNPYReserve:  10000.0,
		TokenReserve: 800000000,
	}

	t.Run("successful simulation", func(t *testing.T) {
		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()

		result, err := processor.SimulateBuy(context.Background(), chainID, 100.0)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.AmountOut)
		assert.True(t, result.AmountOut.Sign() > 0)
		poolRepo.AssertExpectations(t)
	})
}

func TestSimulateSell(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)
	processor := NewOrderProcessor(poolRepo, userRepo, nil)

	chainID := uuid.New()
	pool := &models.VirtualPool{
		ID:           uuid.New(),
		ChainID:      chainID,
		CNPYReserve:  10000.0,
		TokenReserve: 800000000,
	}

	t.Run("successful simulation", func(t *testing.T) {
		poolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil).Once()

		result, err := processor.SimulateSell(context.Background(), chainID, 5000)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.AmountOut)
		assert.True(t, result.AmountOut.Sign() > 0)
		poolRepo.AssertExpectations(t)
	})
}

func TestExtractUserID(t *testing.T) {
	poolRepo := new(MockVirtualPoolRepository)
	userRepo := new(MockUserRepository)
	processor := NewOrderProcessor(poolRepo, userRepo, nil)

	t.Run("valid UUID address", func(t *testing.T) {
		userID := uuid.New()
		address := []byte(userID.String())

		extractedID, err := processor.extractUserID(address)
		assert.NoError(t, err)
		assert.Equal(t, userID, extractedID)
	})

	t.Run("wallet address lookup", func(t *testing.T) {
		userID := uuid.New()
		walletAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
		user := &models.User{
			ID:            userID,
			WalletAddress: walletAddress,
		}

		userRepo.On("GetByWalletAddress", mock.Anything, walletAddress).Return(user, nil).Once()

		extractedID, err := processor.extractUserID([]byte(walletAddress))
		assert.NoError(t, err)
		assert.Equal(t, userID, extractedID)
		userRepo.AssertExpectations(t)
	})

	t.Run("empty address", func(t *testing.T) {
		_, err := processor.extractUserID([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty address")
	})

	t.Run("user not found", func(t *testing.T) {
		walletAddress := "unknown_wallet"
		userRepo.On("GetByWalletAddress", mock.Anything, walletAddress).Return(nil, errors.New("not found")).Once()

		_, err := processor.extractUserID([]byte(walletAddress))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		userRepo.AssertExpectations(t)
	})
}
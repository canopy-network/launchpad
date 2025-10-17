// Package mocks provides mock implementations of repository interfaces for testing.
// These mocks use testify/mock and can be used across all test packages.
package mocks

import (
	"context"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

// MockChainRepository is a mock implementation of interfaces.ChainRepository
type MockChainRepository struct {
	mock.Mock
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

// MockVirtualPoolRepository is a mock implementation of interfaces.VirtualPoolRepository
type MockVirtualPoolRepository struct {
	mock.Mock
}

func (m *MockVirtualPoolRepository) Create(ctx context.Context, pool *models.VirtualPool) (*models.VirtualPool, error) {
	args := m.Called(ctx, pool)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VirtualPool), args.Error(1)
}

func (m *MockVirtualPoolRepository) GetPoolByChainID(ctx context.Context, chainID uuid.UUID) (*models.VirtualPool, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VirtualPool), args.Error(1)
}

func (m *MockVirtualPoolRepository) GetAllPools(ctx context.Context, pagination interfaces.Pagination) ([]models.VirtualPool, int, error) {
	args := m.Called(ctx, pagination)
	return args.Get(0).([]models.VirtualPool), args.Int(1), args.Error(2)
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

func (m *MockVirtualPoolRepository) GetTransactionsByChainID(ctx context.Context, chainID uuid.UUID, filters interfaces.TransactionFilters, pagination interfaces.Pagination) ([]models.VirtualPoolTransaction, int, error) {
	args := m.Called(ctx, chainID, filters, pagination)
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

func (m *MockVirtualPoolRepository) GetPositionsWithUsersByChainID(ctx context.Context, chainID uuid.UUID) ([]interfaces.UserPositionWithAddress, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interfaces.UserPositionWithAddress), args.Error(1)
}

// MockUserRepository is a mock implementation of interfaces.UserRepository
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

// MockVirtualPoolTxRepository is a mock implementation with transaction support
// It embeds MockVirtualPoolRepository to inherit base methods
type MockVirtualPoolTxRepository struct {
	MockVirtualPoolRepository
}

func (m *MockVirtualPoolTxRepository) GetPoolByChainIDForUpdate(ctx context.Context, tx *sqlx.Tx, chainID uuid.UUID) (*models.VirtualPool, error) {
	args := m.Called(ctx, tx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VirtualPool), args.Error(1)
}

func (m *MockVirtualPoolTxRepository) UpdatePoolStateInTx(ctx context.Context, tx *sqlx.Tx, chainID uuid.UUID, update *interfaces.PoolStateUpdate) error {
	args := m.Called(ctx, tx, chainID, update)
	return args.Error(0)
}

func (m *MockVirtualPoolTxRepository) CreateTransactionInTx(ctx context.Context, tx *sqlx.Tx, transaction *models.VirtualPoolTransaction) error {
	args := m.Called(ctx, tx, transaction)
	return args.Error(0)
}

func (m *MockVirtualPoolTxRepository) GetUserPositionForUpdate(ctx context.Context, tx *sqlx.Tx, userID, chainID uuid.UUID) (*models.UserVirtualPosition, error) {
	args := m.Called(ctx, tx, userID, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserVirtualPosition), args.Error(1)
}

func (m *MockVirtualPoolTxRepository) UpsertUserPositionInTx(ctx context.Context, tx *sqlx.Tx, position *models.UserVirtualPosition) error {
	args := m.Called(ctx, tx, position)
	return args.Error(0)
}

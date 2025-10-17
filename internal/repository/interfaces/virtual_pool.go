package interfaces

import (
	"context"
	"math/big"

	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
)

// VirtualPoolRepository defines the interface for virtual pool data operations
type VirtualPoolRepository interface {
	// Pool state operations
	Create(ctx context.Context, pool *models.VirtualPool) (*models.VirtualPool, error)
	GetPoolByChainID(ctx context.Context, chainID uuid.UUID) (*models.VirtualPool, error)
	GetAllPools(ctx context.Context, pagination Pagination) ([]models.VirtualPool, int, error)
	UpdatePoolState(ctx context.Context, chainID uuid.UUID, update *PoolStateUpdate) error

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *models.VirtualPoolTransaction) error
	GetTransactionsByPoolID(ctx context.Context, poolID uuid.UUID, pagination Pagination) ([]models.VirtualPoolTransaction, int, error)
	GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, pagination Pagination) ([]models.VirtualPoolTransaction, int, error)
	GetTransactionsByChainID(ctx context.Context, chainID uuid.UUID, filters TransactionFilters, pagination Pagination) ([]models.VirtualPoolTransaction, int, error)

	// User position operations
	GetUserPosition(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualLPPosition, error)
	UpsertUserPosition(ctx context.Context, position *models.UserVirtualLPPosition) error
	GetPositionsByChainID(ctx context.Context, chainID uuid.UUID, pagination Pagination) ([]models.UserVirtualLPPosition, int, error)
	GetPositionsWithUsersByChainID(ctx context.Context, chainID uuid.UUID) ([]UserPositionWithAddress, error)
}

// UserPositionWithAddress contains position data with user's wallet address
type UserPositionWithAddress struct {
	WalletAddress string
	TokenBalance  int64
}

// PoolStateUpdate represents the fields to update in a virtual pool
type PoolStateUpdate struct {
	CNPYReserve        *big.Float
	TokenReserve       *big.Float
	CurrentPriceCNPY   *big.Float
	MarketCapUSD       *big.Float
	TotalVolumeCNPY    *big.Float
	TotalTransactions  *int
	UniqueTraders      *int
	Volume24hCNPY      *big.Float
	High24hCNPY        *big.Float
	Low24hCNPY         *big.Float
	Price24hChangePerc *big.Float
}

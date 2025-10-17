package interfaces

import (
	"context"
	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
)

// GraduatedPoolRepository defines the interface for graduated pool data operations
type GraduatedPoolRepository interface {
	// Pool state operations
	Create(ctx context.Context, pool *models.GraduatedPool) (*models.GraduatedPool, error)
	GetPoolByChainID(ctx context.Context, chainID uuid.UUID) (*models.GraduatedPool, error)
	GetAllPools(ctx context.Context, pagination Pagination) ([]models.GraduatedPool, int, error)
	UpdatePoolState(ctx context.Context, chainID uuid.UUID, update *PoolStateUpdate) error

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *models.GraduatedPoolTransaction) error
	GetTransactionsByPoolID(ctx context.Context, poolID uuid.UUID, pagination Pagination) ([]models.GraduatedPoolTransaction, int, error)
	GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, pagination Pagination) ([]models.GraduatedPoolTransaction, int, error)
	GetTransactionsByChainID(ctx context.Context, chainID uuid.UUID, filters TransactionFilters, pagination Pagination) ([]models.GraduatedPoolTransaction, int, error)

	// User position operations
	GetUserPosition(ctx context.Context, userID, chainID uuid.UUID) (*models.UserGraduatedLPPosition, error)
	UpsertUserPosition(ctx context.Context, position *models.UserGraduatedLPPosition) error
	GetPositionsByChainID(ctx context.Context, chainID uuid.UUID, pagination Pagination) ([]models.UserGraduatedLPPosition, int, error)
	GetPositionsWithUsersByChainID(ctx context.Context, chainID uuid.UUID) ([]UserPositionWithAddress, error)
}

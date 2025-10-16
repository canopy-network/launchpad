package interfaces

import (
	"context"

	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
)

// WalletRepository defines the interface for wallet data operations
type WalletRepository interface {
	// Wallet CRUD operations
	Create(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	GetByAddress(ctx context.Context, address string) (*models.Wallet, error)
	Update(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Wallet listing and filtering
	List(ctx context.Context, filters WalletFilters, pagination Pagination) ([]models.Wallet, int, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, pagination Pagination) ([]models.Wallet, int, error)
	ListByChainID(ctx context.Context, chainID uuid.UUID, pagination Pagination) ([]models.Wallet, int, error)

	// Security operations
	IncrementFailedAttempts(ctx context.Context, id uuid.UUID) error
	ResetFailedAttempts(ctx context.Context, id uuid.UUID) error
	LockWallet(ctx context.Context, id uuid.UUID, lockDuration int) error // lockDuration in minutes
	UnlockWallet(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

// WalletFilters represents filters for wallet queries
type WalletFilters struct {
	UserID    *uuid.UUID
	ChainID   *uuid.UUID
	IsActive  *bool
	IsLocked  *bool
	CreatedBy *uuid.UUID
}

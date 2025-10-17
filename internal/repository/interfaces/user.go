package interfaces

import (
	"context"

	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByWalletAddress(ctx context.Context, walletAddress string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// User listing and filtering
	List(ctx context.Context, filters UserFilters, pagination Pagination) ([]models.User, int, error)
	ListByVerificationTier(ctx context.Context, tier string, pagination Pagination) ([]models.User, int, error)

	// User positions and portfolio
	GetPositionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.UserVirtualLPPosition, error)
	GetPositionByUserAndChain(ctx context.Context, userID, chainID uuid.UUID) (*models.UserVirtualLPPosition, error)
	UpdatePosition(ctx context.Context, position *models.UserVirtualLPPosition) (*models.UserVirtualLPPosition, error)

	// User statistics updates
	UpdateChainsCreatedCount(ctx context.Context, userID uuid.UUID, increment int) error
	UpdateCNPYInvestment(ctx context.Context, userID uuid.UUID, amount float64) error
	UpdateReputationScore(ctx context.Context, userID uuid.UUID, score int) error
	UpdateLastActive(ctx context.Context, userID uuid.UUID) error

	// Authentication and session management
	CreateOrGetByEmail(ctx context.Context, email string) (*models.User, bool, error) // Returns (user, isNew, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	IncrementJWTVersion(ctx context.Context, userID uuid.UUID) error

	// Profile management
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error)
}

// UserFilters represents filters for user queries
type UserFilters struct {
	VerificationTier *string
	IsVerified       *bool
	MinReputation    *int
}

package interfaces

import (
	"context"

	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
)

// ChainRepository defines the interface for chain data operations
type ChainRepository interface {
	// Chain CRUD operations
	Create(ctx context.Context, chain *models.Chain) (*models.Chain, error)
	GetByID(ctx context.Context, id uuid.UUID, includeRelations []string) (*models.Chain, error)
	GetByName(ctx context.Context, name string) (*models.Chain, error)
	GetByAddress(ctx context.Context, address string) (*models.Chain, error)
	Update(ctx context.Context, chain *models.Chain) (*models.Chain, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Chain listing and filtering
	List(ctx context.Context, filters ChainFilters, pagination Pagination) ([]models.Chain, int, error)
	ListByCreator(ctx context.Context, creatorID uuid.UUID, pagination Pagination) ([]models.Chain, int, error)
	ListByTemplate(ctx context.Context, templateID uuid.UUID, pagination Pagination) ([]models.Chain, int, error)
	ListByStatus(ctx context.Context, status string, pagination Pagination) ([]models.Chain, int, error)

	// Repository operations
	CreateRepository(ctx context.Context, repo *models.ChainRepository) (*models.ChainRepository, error)
	UpdateRepository(ctx context.Context, repo *models.ChainRepository) (*models.ChainRepository, error)
	GetRepositoryByChainID(ctx context.Context, chainID uuid.UUID) (*models.ChainRepository, error)
	DeleteRepository(ctx context.Context, chainID uuid.UUID) error

	// Social links operations
	CreateSocialLinks(ctx context.Context, chainID uuid.UUID, links []models.ChainSocialLink) error
	UpdateSocialLinks(ctx context.Context, chainID uuid.UUID, links []models.ChainSocialLink) error
	GetSocialLinksByChainID(ctx context.Context, chainID uuid.UUID) ([]models.ChainSocialLink, error)
	DeleteSocialLinksByChainID(ctx context.Context, chainID uuid.UUID) error

	// Assets operations
	CreateAssets(ctx context.Context, chainID uuid.UUID, assets []models.ChainAsset) error
	UpdateAssets(ctx context.Context, chainID uuid.UUID, assets []models.ChainAsset) error
	GetAssetsByChainID(ctx context.Context, chainID uuid.UUID) ([]models.ChainAsset, error)
	DeleteAssetsByChainID(ctx context.Context, chainID uuid.UUID) error

	// Chain key operations
	CreateChainKey(ctx context.Context, key *models.ChainKey) (*models.ChainKey, error)
	GetChainKeyByChainID(ctx context.Context, chainID uuid.UUID, purpose string) (*models.ChainKey, error)
}

// ChainTemplateRepository defines the interface for chain template operations
type ChainTemplateRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.ChainTemplate, error)
	List(ctx context.Context, filters TemplateFilters, pagination Pagination) ([]models.ChainTemplate, int, error)
	GetByCategory(ctx context.Context, category string) ([]models.ChainTemplate, error)
	GetActive(ctx context.Context) ([]models.ChainTemplate, error)
}

// Filter structures
type ChainFilters struct {
	Status     string
	CreatedBy  *uuid.UUID
	TemplateID *uuid.UUID
	Include    []string
}

type TransactionFilters struct {
	UserID          *uuid.UUID
	TransactionType string
}

type TemplateFilters struct {
	Category        string
	ComplexityLevel string
	IsActive        *bool
}

type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

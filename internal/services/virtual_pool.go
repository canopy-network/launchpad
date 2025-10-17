package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
)

type VirtualPoolService struct {
	virtualPoolRepo interfaces.VirtualPoolRepository
}

func NewVirtualPoolService(virtualPoolRepo interfaces.VirtualPoolRepository) *VirtualPoolService {
	return &VirtualPoolService{
		virtualPoolRepo: virtualPoolRepo,
	}
}

// GetAllPools retrieves all virtual pools with pagination
func (s *VirtualPoolService) GetAllPools(ctx context.Context, page, limit int) ([]models.VirtualPool, *models.Pagination, error) {
	// Build pagination
	pagination := interfaces.Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Get all pools
	pools, total, err := s.virtualPoolRepo.GetAllPools(ctx, pagination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get virtual pools: %w", err)
	}

	// Build pagination response
	paginationResp := &models.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
		Pages: (total + limit - 1) / limit, // Ceiling division
	}

	return pools, paginationResp, nil
}

// GetPool retrieves virtual pool data for a chain
func (s *VirtualPoolService) GetPool(ctx context.Context, chainID string) (*models.VirtualPool, error) {
	chainUUID, err := uuid.Parse(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Get virtual pool using VirtualPoolRepository
	pool, err := s.virtualPoolRepo.GetPoolByChainID(ctx, chainUUID)
	if err != nil {
		if strings.Contains(err.Error(), "virtual pool not found") {
			return nil, fmt.Errorf("virtual pool not found for chain")
		}
		return nil, fmt.Errorf("failed to get virtual pool: %w", err)
	}

	return pool, nil
}

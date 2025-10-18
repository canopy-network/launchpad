package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
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

// GetPriceHistory retrieves OHLC price history for a chain
func (s *VirtualPoolService) GetPriceHistory(ctx context.Context, chainID string, startTime, endTime *time.Time) ([]models.PriceHistoryCandle, error) {
	// Parse and validate chain ID
	chainUUID, err := uuid.Parse(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Set default time range: last 24 hours if not provided
	now := time.Now()
	var start, end time.Time

	if startTime != nil {
		start = *startTime
	} else {
		start = now.Add(-24 * time.Hour)
	}

	if endTime != nil {
		end = *endTime
	} else {
		end = now
	}

	// Validate time range
	if end.Before(start) || end.Equal(start) {
		return nil, fmt.Errorf("end_time must be after start_time")
	}

	// Get price history from repository
	repoCandles, err := s.virtualPoolRepo.GetPriceHistory(ctx, chainUUID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	// Convert repository candles to model candles
	candles := make([]models.PriceHistoryCandle, len(repoCandles))
	for i, rc := range repoCandles {
		candles[i] = models.PriceHistoryCandle{
			Timestamp:  rc.Timestamp,
			Open:       rc.Open,
			High:       rc.High,
			Low:        rc.Low,
			Close:      rc.Close,
			Volume:     rc.Volume,
			TradeCount: rc.TradeCount,
		}
	}

	return candles, nil
}

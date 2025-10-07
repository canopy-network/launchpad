package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/keygen"
	"github.com/google/uuid"
)

var (
	ErrChainNotFound         = errors.New("chain not found")
	ErrChainAlreadyExists    = errors.New("chain already exists")
	ErrChainNotInDraftStatus = errors.New("chain is not in draft status")
	ErrUnauthorized          = errors.New("unauthorized")
)

type ChainService struct {
	chainRepo       interfaces.ChainRepository
	templateRepo    interfaces.ChainTemplateRepository
	userRepo        interfaces.UserRepository
	virtualPoolRepo interfaces.VirtualPoolRepository
}

func NewChainService(chainRepo interfaces.ChainRepository, templateRepo interfaces.ChainTemplateRepository, userRepo interfaces.UserRepository, virtualPoolRepo interfaces.VirtualPoolRepository) *ChainService {
	return &ChainService{
		chainRepo:       chainRepo,
		templateRepo:    templateRepo,
		userRepo:        userRepo,
		virtualPoolRepo: virtualPoolRepo,
	}
}

// CreateChain creates a new chain
func (s *ChainService) CreateChain(ctx context.Context, req *models.CreateChainRequest, userID string) (*models.Chain, error) {
	// Parse user ID
	createdBy, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate template exists (if provided)
	var templateID *uuid.UUID
	var template *models.ChainTemplate
	if req.TemplateID != nil && *req.TemplateID != "" {
		parsedTemplateID, err := uuid.Parse(*req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("invalid template ID: %w", err)
		}
		templateID = &parsedTemplateID

		if s.templateRepo != nil {
			template, err = s.templateRepo.GetByID(ctx, parsedTemplateID)
			if err != nil {
				return nil, fmt.Errorf("template not found: %w", err)
			}
		} else {
			return nil, fmt.Errorf("template repository not available")
		}
	}

	// Check if chain name already exists
	existingChain, err := s.chainRepo.GetByName(ctx, req.ChainName)
	if err == nil && existingChain != nil {
		return nil, ErrChainAlreadyExists
	}

	// Create chain with defaults from template and request
	defaultConsensus := "tendermint"
	defaultTokenSupply := int64(1000000000)
	if template != nil {
		// Template only provides default token supply now
		defaultTokenSupply = template.DefaultTokenSupply
	}

	chain := &models.Chain{
		ChainName:                  req.ChainName,
		TokenSymbol:                strings.ToUpper(req.TokenSymbol),
		ChainDescription:           req.ChainDescription,
		TemplateID:                 templateID,
		ConsensusMechanism:         s.getStringValueOrDefault(&req.ConsensusMechanism, defaultConsensus),
		TokenTotalSupply:           s.getInt64ValueOrDefault(req.TokenTotalSupply, defaultTokenSupply),
		GraduationThreshold:        s.getFloat64ValueOrDefault(req.GraduationThreshold, 50000.00),
		CreationFeeCNPY:            s.getFloat64ValueOrDefault(req.CreationFeeCNPY, 100.00000000),
		InitialCNPYReserve:         s.getFloat64ValueOrDefault(req.InitialCNPYReserve, 10000.00000000),
		InitialTokenSupply:         s.getInt64ValueOrDefault(req.InitialTokenSupply, 800000000),
		BondingCurveSlope:          s.getFloat64ValueOrDefault(req.BondingCurveSlope, 0.00000001),
		ValidatorMinStake:          s.getFloat64ValueOrDefault(req.ValidatorMinStake, 1000.00000000),
		CreatorInitialPurchaseCNPY: s.getFloat64ValueOrDefault(req.CreatorInitialPurchaseCNPY, 0),
		Status:                     models.ChainStatusDraft,
		IsGraduated:                false,
		CreatedBy:                  createdBy,
	}

	// Save to database
	createdChain, err := s.chainRepo.Create(ctx, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain: %w", err)
	}

	// Generate and encrypt keypair for the chain
	// Using a default password for now - in production, this should be configurable or derived from user secrets
	defaultPassword := "changeme" // TODO: make this configurable or user-provided
	_, encryptedKeyPair, err := keygen.GenerateEncryptedKeyPair(defaultPassword)
	if err != nil {
		// Rollback chain creation on key generation failure
		// In production, this should use a transaction
		_ = s.chainRepo.Delete(ctx, createdChain.ID)
		return nil, fmt.Errorf("failed to generate chain keypair: %w", err)
	}

	// Convert encrypted keypair to ChainKey model
	// addressBytes, err := hex.DecodeString(encryptedKeyPair.Address)
	// if err != nil {
	// 	_ = s.chainRepo.Delete(ctx, createdChain.ID)
	// 	return nil, fmt.Errorf("failed to decode address: %w", err)
	// }

	publicKeyBytes, err := hex.DecodeString(encryptedKeyPair.PublicKey)
	if err != nil {
		_ = s.chainRepo.Delete(ctx, createdChain.ID)
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	saltBytes, err := hex.DecodeString(encryptedKeyPair.Salt)
	if err != nil {
		_ = s.chainRepo.Delete(ctx, createdChain.ID)
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	chainKey := &models.ChainKey{
		ChainID:             createdChain.ID,
		Address:             encryptedKeyPair.Address,
		PublicKey:           publicKeyBytes,
		EncryptedPrivateKey: encryptedKeyPair.EncryptedPrivateKey,
		Salt:                saltBytes,
		KeyNickname:         nil,
		KeyPurpose:          models.KeyPurposeChainOperation,
		IsActive:            true,
		RotationCount:       0,
	}

	// Save chain key to database
	_, err = s.chainRepo.CreateChainKey(ctx, chainKey)
	if err != nil {
		// Rollback chain creation on key storage failure
		_ = s.chainRepo.Delete(ctx, createdChain.ID)
		return nil, fmt.Errorf("failed to store chain key: %w", err)
	}

	// Update user's chain count (if user repository is available)
	if s.userRepo != nil {
		err = s.userRepo.UpdateChainsCreatedCount(ctx, createdBy, 1)
		if err != nil {
			// Log error but don't fail the chain creation
			// In a production system, you might want to use a proper logger
			fmt.Printf("Failed to update user chain count: %v\n", err)
		}
	}

	// Load relations for response (only load what's available)
	includeRelations := []string{}
	if s.templateRepo != nil {
		includeRelations = append(includeRelations, "template")
	}
	// Don't include creator since user repository is not implemented yet
	return s.chainRepo.GetByID(ctx, createdChain.ID, includeRelations)
}

// GetChains retrieves chains with filtering and pagination
func (s *ChainService) GetChains(ctx context.Context, status, createdBy string, include []string, page, limit int) ([]models.Chain, *models.Pagination, error) {
	// Build filters
	filters := interfaces.ChainFilters{
		Status:  status,
		Include: include,
	}

	if createdBy != "" {
		createdByUUID, err := uuid.Parse(createdBy)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid created_by UUID: %w", err)
		}
		filters.CreatedBy = &createdByUUID
	}

	// Build pagination
	pagination := interfaces.Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Get chains
	chains, total, err := s.chainRepo.List(ctx, filters, pagination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chains: %w", err)
	}

	// Build pagination response
	paginationResp := &models.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
		Pages: (total + limit - 1) / limit, // Ceiling division
	}

	return chains, paginationResp, nil
}

// GetChainByID retrieves a chain by ID
func (s *ChainService) GetChainByID(ctx context.Context, id string, include string) (*models.Chain, error) {
	chainID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	includeRelations := []string{}
	if include != "" {
		includeRelations = strings.Split(include, ",")
	}

	chain, err := s.chainRepo.GetByID(ctx, chainID, includeRelations)
	if err != nil {
		if err.Error() == "chain not found" {
			return nil, ErrChainNotFound
		}
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	return chain, nil
}

// DeleteChain deletes a chain (only if in draft status)
func (s *ChainService) DeleteChain(ctx context.Context, chainID string, userID string) error {
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return err
	}

	// Validate chain is in draft status
	if chain.Status != models.ChainStatusDraft {
		return fmt.Errorf("can only delete chains in draft status")
	}

	// Delete the chain
	err = s.chainRepo.Delete(ctx, chain.ID)
	if err != nil {
		return fmt.Errorf("failed to delete chain: %w", err)
	}

	// Update user's chain count (if user repository is available)
	if s.userRepo != nil {
		createdByUUID, _ := uuid.Parse(userID)
		err = s.userRepo.UpdateChainsCreatedCount(ctx, createdByUUID, -1)
		if err != nil {
			// Log error but don't fail the deletion
			fmt.Printf("Failed to update user chain count: %v\n", err)
		}
	}

	return nil
}

// GetVirtualPool retrieves virtual pool data for a chain
func (s *ChainService) GetVirtualPool(ctx context.Context, chainID string) (*models.VirtualPool, error) {
	chainUUID, err := uuid.Parse(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Verify chain exists
	_, err = s.chainRepo.GetByID(ctx, chainUUID, nil)
	if err != nil {
		if err.Error() == "chain not found" {
			return nil, ErrChainNotFound
		}
		return nil, fmt.Errorf("failed to get chain: %w", err)
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

// GetTransactions retrieves virtual pool transactions for a chain
func (s *ChainService) GetTransactions(ctx context.Context, chainID string, userID, transactionType string, page, limit int) ([]models.VirtualPoolTransaction, *models.Pagination, error) {
	chainUUID, err := uuid.Parse(chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Verify chain exists
	_, err = s.chainRepo.GetByID(ctx, chainUUID, nil)
	if err != nil {
		if err.Error() == "chain not found" {
			return nil, nil, ErrChainNotFound
		}
		return nil, nil, fmt.Errorf("failed to get chain: %w", err)
	}

	// Build filters
	filters := interfaces.TransactionFilters{
		TransactionType: transactionType,
	}

	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid user ID: %w", err)
		}
		filters.UserID = &userUUID
	}

	// Build pagination
	pagination := interfaces.Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Get transactions using VirtualPoolRepository
	transactions, total, err := s.virtualPoolRepo.GetTransactionsByChainID(ctx, chainUUID, filters, pagination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Build pagination response
	paginationResp := &models.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
		Pages: (total + limit - 1) / limit,
	}

	return transactions, paginationResp, nil
}

// Helper methods
func (s *ChainService) getChainAndValidateOwnership(ctx context.Context, chainID, userID string) (*models.Chain, error) {
	chainUUID, err := uuid.Parse(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	chain, err := s.chainRepo.GetByID(ctx, chainUUID, nil)
	if err != nil {
		if err.Error() == "chain not found" {
			return nil, ErrChainNotFound
		}
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	// Validate ownership
	if chain.CreatedBy != userUUID {
		return nil, ErrUnauthorized
	}

	return chain, nil
}

func (s *ChainService) getStringValueOrDefault(ptr *string, defaultValue string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return defaultValue
}

func (s *ChainService) getInt64ValueOrDefault(ptr *int64, defaultValue int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func (s *ChainService) getFloat64ValueOrDefault(ptr *float64, defaultValue float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func (s *ChainService) getBoolValueOrDefault(ptr *bool, defaultValue bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func (s *ChainService) getIntValueOrDefault(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

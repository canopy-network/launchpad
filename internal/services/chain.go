package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

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
	ErrRepositoryNotFound    = errors.New("repository not found")
	ErrAssetNotFound         = errors.New("asset not found")
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
		TokenName:                  req.TokenName,
		TokenSymbol:                strings.ToUpper(req.TokenSymbol),
		ChainDescription:           req.ChainDescription,
		TemplateID:                 templateID,
		ConsensusMechanism:         s.getStringValueOrDefault(&req.ConsensusMechanism, defaultConsensus),
		TokenTotalSupply:           s.getInt64ValueOrDefault(req.TokenTotalSupply, defaultTokenSupply),
		BlockTimeSeconds:           req.BlockTimeSeconds,
		UpgradeBlockHeight:         req.UpgradeBlockHeight,
		BlockRewardAmount:          req.BlockRewardAmount,
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

	// Create GitHub repository if provided
	if req.GithubURL != nil && *req.GithubURL != "" {
		repo := &models.ChainRepository{
			ChainID:            createdChain.ID,
			GithubURL:          *req.GithubURL,
			RepositoryName:     extractRepoName(*req.GithubURL),
			RepositoryOwner:    extractRepoOwner(*req.GithubURL),
			DefaultBranch:      "main",
			IsConnected:        false,
			AutoUpgradeEnabled: true,
			UpgradeTrigger:     models.UpgradeTriggerTagRelease,
			BuildStatus:        models.BuildStatusPending,
		}
		_, err = s.chainRepo.CreateRepository(ctx, repo)
		if err != nil {
			// Log error but don't fail the chain creation
			fmt.Printf("Failed to create chain repository: %v\n", err)
		}
	}

	// Create social links if provided
	socialLinks := []models.ChainSocialLink{}
	displayOrder := 0

	if req.TwitterURL != nil && *req.TwitterURL != "" {
		socialLinks = append(socialLinks, models.ChainSocialLink{
			ChainID:      createdChain.ID,
			Platform:     models.PlatformTwitter,
			URL:          *req.TwitterURL,
			DisplayOrder: displayOrder,
			IsActive:     true,
		})
		displayOrder++
	}

	if req.TelegramURL != nil && *req.TelegramURL != "" {
		socialLinks = append(socialLinks, models.ChainSocialLink{
			ChainID:      createdChain.ID,
			Platform:     models.PlatformTelegram,
			URL:          *req.TelegramURL,
			DisplayOrder: displayOrder,
			IsActive:     true,
		})
		displayOrder++
	}

	if req.WebsiteURL != nil && *req.WebsiteURL != "" {
		socialLinks = append(socialLinks, models.ChainSocialLink{
			ChainID:      createdChain.ID,
			Platform:     models.PlatformWebsite,
			URL:          *req.WebsiteURL,
			DisplayOrder: displayOrder,
			IsActive:     true,
		})
		displayOrder++
	}

	if len(socialLinks) > 0 {
		err = s.chainRepo.CreateSocialLinks(ctx, createdChain.ID, socialLinks)
		if err != nil {
			// Log error but don't fail the chain creation
			fmt.Printf("Failed to create social links: %v\n", err)
		}
	}

	// Create assets if provided
	assets := []models.ChainAsset{}
	assetDisplayOrder := 0

	if req.WhitepaperURL != nil && *req.WhitepaperURL != "" {
		assets = append(assets, models.ChainAsset{
			ChainID:          createdChain.ID,
			AssetType:        models.AssetTypeWhitepaper,
			FileName:         "whitepaper.pdf",
			FileURL:          *req.WhitepaperURL,
			DisplayOrder:     assetDisplayOrder,
			IsPrimary:        false,
			IsFeatured:       false,
			IsActive:         true,
			ModerationStatus: "pending",
			UploadedBy:       createdBy,
		})
		assetDisplayOrder++
	}

	if req.TokenImageURL != nil && *req.TokenImageURL != "" {
		assets = append(assets, models.ChainAsset{
			ChainID:          createdChain.ID,
			AssetType:        models.AssetTypeLogo,
			FileName:         "logo",
			FileURL:          *req.TokenImageURL,
			DisplayOrder:     assetDisplayOrder,
			IsPrimary:        true,
			IsFeatured:       true,
			IsActive:         true,
			ModerationStatus: "pending",
			UploadedBy:       createdBy,
		})
		assetDisplayOrder++
	}

	if req.TokenVideoURL != nil && *req.TokenVideoURL != "" {
		assets = append(assets, models.ChainAsset{
			ChainID:          createdChain.ID,
			AssetType:        models.AssetTypeVideo,
			FileName:         "promo_video",
			FileURL:          *req.TokenVideoURL,
			DisplayOrder:     assetDisplayOrder,
			IsPrimary:        false,
			IsFeatured:       true,
			IsActive:         true,
			ModerationStatus: "pending",
			UploadedBy:       createdBy,
		})
		assetDisplayOrder++
	}

	if len(assets) > 0 {
		err = s.chainRepo.CreateAssets(ctx, createdChain.ID, assets)
		if err != nil {
			// Log error but don't fail the chain creation
			fmt.Printf("Failed to create assets: %v\n", err)
		}
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

// UpdateChainDescription updates the description of a chain
func (s *ChainService) UpdateChainDescription(ctx context.Context, chainID string, userID string, req *models.UpdateChainDescriptionRequest) (*models.Chain, error) {
	// Validate ownership
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}

	// Update description
	err = s.chainRepo.UpdateDescription(ctx, chain.ID, req.ChainDescription)
	if err != nil {
		if err.Error() == "chain not found" {
			return nil, ErrChainNotFound
		}
		return nil, fmt.Errorf("failed to update chain description: %w", err)
	}

	// Return updated chain
	return s.chainRepo.GetByID(ctx, chain.ID, nil)
}

// GetRepositoryByChainID retrieves a GitHub repository by chain ID
func (s *ChainService) GetRepositoryByChainID(ctx context.Context, chainID string, userID string) (*models.ChainRepository, error) {
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}

	repo, err := s.chainRepo.GetRepositoryByChainID(ctx, chain.ID)
	if err != nil {
		if err.Error() == "repository not found" {
			return nil, ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo, nil
}

// UpdateRepositoryByChainID updates a GitHub repository by chain ID (partial update)
func (s *ChainService) UpdateRepositoryByChainID(ctx context.Context, chainID string, userID string, req *models.UpdateChainRepositoryRequest) (*models.ChainRepository, error) {
	// Validate ownership
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}

	// Get existing repository
	existingRepo, err := s.chainRepo.GetRepositoryByChainID(ctx, chain.ID)
	if err != nil {
		if err.Error() == "repository not found" {
			return nil, ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Apply partial updates
	if req.GithubURL != nil {
		existingRepo.GithubURL = *req.GithubURL
	}
	if req.RepositoryName != nil {
		existingRepo.RepositoryName = *req.RepositoryName
	}
	if req.RepositoryOwner != nil {
		existingRepo.RepositoryOwner = *req.RepositoryOwner
	}
	if req.DefaultBranch != nil {
		existingRepo.DefaultBranch = *req.DefaultBranch
	}

	// Update in database
	updatedRepo, err := s.chainRepo.UpdateRepository(ctx, existingRepo)
	if err != nil {
		if err.Error() == "repository not found" {
			return nil, ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("failed to update repository: %w", err)
	}

	return updatedRepo, nil
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

// GetPriceHistory retrieves OHLC price history for a chain
func (s *ChainService) GetPriceHistory(ctx context.Context, chainID string, startTime, endTime *time.Time) ([]models.PriceHistoryCandle, error) {
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

	// Delegate to virtual pool service (reuse the logic)
	// Actually, let's call the repository directly since we already have access
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

// extractRepoName extracts repository name from GitHub URL
// Example: https://github.com/owner/repo -> "repo"
func extractRepoName(githubURL string) string {
	parts := strings.Split(strings.TrimSuffix(githubURL, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractRepoOwner extracts repository owner from GitHub URL
// Example: https://github.com/owner/repo -> "owner"
func extractRepoOwner(githubURL string) string {
	parts := strings.Split(strings.TrimSuffix(githubURL, "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}

// GetAssets retrieves all assets for a chain
func (s *ChainService) GetAssets(ctx context.Context, chainID string, userID string) ([]models.ChainAsset, error) {
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}

	assets, err := s.chainRepo.GetAssetsByChainID(ctx, chain.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	return assets, nil
}

// CreateAsset creates a new asset for a chain
func (s *ChainService) CreateAsset(ctx context.Context, chainID string, userID string, req *models.CreateChainAssetRequest) (*models.ChainAsset, error) {
	// Validate ownership
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}

	// Parse user ID for uploaded_by
	uploadedByUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// If no display order provided, get the next available order
	displayOrder := s.getIntValueOrDefault(req.DisplayOrder, 0)
	if req.DisplayOrder == nil {
		existingAssets, err := s.chainRepo.GetAssetsByChainID(ctx, chain.ID)
		if err == nil {
			displayOrder = len(existingAssets)
		}
	}

	// Create asset
	asset := models.ChainAsset{
		ChainID:          chain.ID,
		AssetType:        req.AssetType,
		FileName:         req.FileName,
		FileURL:          req.FileURL,
		FileSizeBytes:    req.FileSizeBytes,
		MimeType:         req.MimeType,
		Title:            req.Title,
		Description:      req.Description,
		AltText:          req.AltText,
		DisplayOrder:     displayOrder,
		IsPrimary:        s.getBoolValueOrDefault(req.IsPrimary, false),
		IsFeatured:       s.getBoolValueOrDefault(req.IsFeatured, false),
		IsActive:         true,
		ModerationStatus: "pending",
		UploadedBy:       uploadedByUUID,
	}

	// Create asset in database
	err = s.chainRepo.CreateAssets(ctx, chain.ID, []models.ChainAsset{asset})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	// Get the created asset by querying all assets and finding the matching one
	assets, err := s.chainRepo.GetAssetsByChainID(ctx, chain.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created asset: %w", err)
	}

	// Find the newly created asset (last one with matching properties)
	for i := len(assets) - 1; i >= 0; i-- {
		if assets[i].AssetType == req.AssetType &&
			assets[i].FileName == req.FileName &&
			assets[i].FileURL == req.FileURL {
			return &assets[i], nil
		}
	}

	return nil, fmt.Errorf("failed to retrieve created asset")
}

// UpdateAsset updates an existing asset for a chain
func (s *ChainService) UpdateAsset(ctx context.Context, chainID string, assetID string, userID string, req *models.UpdateChainAssetRequest) (*models.ChainAsset, error) {
	// Validate ownership
	chain, err := s.getChainAndValidateOwnership(ctx, chainID, userID)
	if err != nil {
		return nil, err
	}

	// Parse asset ID
	assetUUID, err := uuid.Parse(assetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset ID: %w", err)
	}

	// Get existing asset
	existingAsset, err := s.chainRepo.GetAssetByID(ctx, assetUUID)
	if err != nil {
		if err.Error() == "asset not found" {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	// Verify the asset belongs to this chain
	if existingAsset.ChainID != chain.ID {
		return nil, ErrAssetNotFound
	}

	// Apply partial updates
	if req.FileName != nil {
		existingAsset.FileName = *req.FileName
	}
	if req.FileURL != nil {
		existingAsset.FileURL = *req.FileURL
	}
	if req.FileSizeBytes != nil {
		existingAsset.FileSizeBytes = req.FileSizeBytes
	}
	if req.MimeType != nil {
		existingAsset.MimeType = req.MimeType
	}
	if req.Title != nil {
		existingAsset.Title = req.Title
	}
	if req.Description != nil {
		existingAsset.Description = req.Description
	}
	if req.AltText != nil {
		existingAsset.AltText = req.AltText
	}
	if req.DisplayOrder != nil {
		existingAsset.DisplayOrder = *req.DisplayOrder
	}
	if req.IsPrimary != nil {
		existingAsset.IsPrimary = *req.IsPrimary
	}
	if req.IsFeatured != nil {
		existingAsset.IsFeatured = *req.IsFeatured
	}
	if req.IsActive != nil {
		existingAsset.IsActive = *req.IsActive
	}
	if req.ModerationStatus != nil {
		existingAsset.ModerationStatus = *req.ModerationStatus
	}
	if req.ModerationNotes != nil {
		existingAsset.ModerationNotes = req.ModerationNotes
	}

	// Update in database
	err = s.chainRepo.UpdateAsset(ctx, existingAsset)
	if err != nil {
		if err.Error() == "asset not found" {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	return existingAsset, nil
}

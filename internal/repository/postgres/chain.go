package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type chainRepository struct {
	db           *sqlx.DB
	userRepo     interfaces.UserRepository
	templateRepo interfaces.ChainTemplateRepository
}

// NewChainRepository creates a new PostgreSQL chain repository
func NewChainRepository(db *sqlx.DB, userRepo interfaces.UserRepository, templateRepo interfaces.ChainTemplateRepository) interfaces.ChainRepository {
	return &chainRepository{
		db:           db,
		userRepo:     userRepo,
		templateRepo: templateRepo,
	}
}

// Create creates a new chain
func (r *chainRepository) Create(ctx context.Context, chain *models.Chain) (*models.Chain, error) {
	query := `
		INSERT INTO chains (
			chain_name, token_symbol, chain_description, template_id, consensus_mechanism,
			token_total_supply, graduation_threshold, creation_fee_cnpy, initial_cnpy_reserve,
			initial_token_supply, bonding_curve_slope, creator_initial_purchase_cnpy,
			validator_min_stake, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id, status, is_graduated, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		chain.ChainName,
		chain.TokenSymbol,
		database.NullString(chain.ChainDescription),
		database.NullUUID(chain.TemplateID),
		chain.ConsensusMechanism,
		chain.TokenTotalSupply,
		chain.GraduationThreshold,
		chain.CreationFeeCNPY,
		chain.InitialCNPYReserve,
		chain.InitialTokenSupply,
		chain.BondingCurveSlope,
		chain.CreatorInitialPurchaseCNPY,
		chain.ValidatorMinStake,
		chain.CreatedBy,
	).Scan(&chain.ID, &chain.Status, &chain.IsGraduated, &chain.CreatedAt, &chain.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("chain name already exists")
		}
		return nil, fmt.Errorf("failed to create chain: %w", err)
	}

	return chain, nil
}

// GetByID retrieves a chain by ID with optional relations
func (r *chainRepository) GetByID(ctx context.Context, id uuid.UUID, includeRelations []string) (*models.Chain, error) {
	chain, err := r.getChainByField(ctx, "id", id.String())
	if err != nil {
		return nil, err
	}

	// Load relations if requested
	if err := r.loadChainRelations(ctx, chain, includeRelations); err != nil {
		return nil, fmt.Errorf("failed to load relations: %w", err)
	}

	return chain, nil
}

// GetByName retrieves a chain by name
func (r *chainRepository) GetByName(ctx context.Context, name string) (*models.Chain, error) {
	return r.getChainByField(ctx, "chain_name", name)
}

// GetByAddress retrieves a chain by its key address
func (r *chainRepository) GetByAddress(ctx context.Context, address string) (*models.Chain, error) {
	query := `
		SELECT c.id, c.chain_name, c.token_name, c.token_symbol, c.chain_description, c.template_id,
			c.consensus_mechanism, c.token_total_supply, c.block_time_seconds, c.upgrade_block_height,
			c.block_reward_amount, c.graduation_threshold, c.creation_fee_cnpy, c.initial_cnpy_reserve,
			c.initial_token_supply, c.bonding_curve_slope, c.scheduled_launch_time, c.actual_launch_time,
			c.creator_initial_purchase_cnpy, c.status, c.is_graduated, c.graduation_time,
			c.chain_id, c.genesis_hash, c.validator_min_stake, c.created_by,
			c.created_at, c.updated_at
		FROM chains c
		INNER JOIN chain_keys ck ON c.id = ck.chain_id
		WHERE ck.address = $1 AND ck.is_active = true`

	var chain models.Chain
	err := r.db.GetContext(ctx, &chain, query, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain not found")
		}
		return nil, fmt.Errorf("failed to get chain by address: %w", err)
	}

	return &chain, nil
}

// Update updates a chain
func (r *chainRepository) Update(ctx context.Context, chain *models.Chain) (*models.Chain, error) {
	query := `
		UPDATE chains SET
			chain_name = $2, token_symbol = $3, chain_description = $4, template_id = $5,
			consensus_mechanism = $6, token_total_supply = $7, graduation_threshold = $8,
			creation_fee_cnpy = $9, initial_cnpy_reserve = $10, initial_token_supply = $11,
			bonding_curve_slope = $12, scheduled_launch_time = $13, actual_launch_time = $14,
			creator_initial_purchase_cnpy = $15, status = $16, is_graduated = $17,
			graduation_time = $18, chain_id = $19, genesis_hash = $20, validator_min_stake = $21,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		chain.ID,
		chain.ChainName,
		chain.TokenSymbol,
		database.NullString(chain.ChainDescription),
		database.NullUUID(chain.TemplateID),
		chain.ConsensusMechanism,
		chain.TokenTotalSupply,
		chain.GraduationThreshold,
		chain.CreationFeeCNPY,
		chain.InitialCNPYReserve,
		chain.InitialTokenSupply,
		chain.BondingCurveSlope,
		chain.ScheduledLaunchTime,
		chain.ActualLaunchTime,
		chain.CreatorInitialPurchaseCNPY,
		chain.Status,
		chain.IsGraduated,
		chain.GraduationTime,
		database.NullString(chain.ChainID),
		database.NullString(chain.GenesisHash),
		chain.ValidatorMinStake,
	).Scan(&chain.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain not found")
		}
		return nil, fmt.Errorf("failed to update chain: %w", err)
	}

	return chain, nil
}

// UpdateDescription updates only the chain description
func (r *chainRepository) UpdateDescription(ctx context.Context, id uuid.UUID, description string) error {
	query := `
		UPDATE chains SET
			chain_description = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, description)
	if err != nil {
		return fmt.Errorf("failed to update chain description: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chain not found")
	}

	return nil
}

// Delete deletes a chain
func (r *chainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM chains WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chain: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chain not found")
	}

	return nil
}

// List retrieves chains with filtering and pagination
func (r *chainRepository) List(ctx context.Context, filters interfaces.ChainFilters, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	whereClause, args := r.buildChainWhereClause(filters)

	// Count query
	countQuery := "SELECT COUNT(*) FROM chains c" + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chains: %w", err)
	}

	// Data query - explicitly list columns to ensure correct scan order
	dataQuery := fmt.Sprintf(`
		SELECT c.id, c.chain_name, c.token_name, c.token_symbol, c.chain_description,
			c.template_id, c.consensus_mechanism, c.token_total_supply, c.block_time_seconds,
			c.upgrade_block_height, c.block_reward_amount, c.graduation_threshold, c.creation_fee_cnpy,
			c.initial_cnpy_reserve, c.initial_token_supply, c.bonding_curve_slope,
			c.scheduled_launch_time, c.actual_launch_time, c.creator_initial_purchase_cnpy,
			c.status, c.is_graduated, c.graduation_time, c.chain_id, c.genesis_hash,
			c.validator_min_stake, c.created_by, c.created_at, c.updated_at,
			ct.template_name, ct.template_description, u.wallet_address, u.display_name
		FROM chains c
		LEFT JOIN chain_templates ct ON c.template_id = ct.id
		LEFT JOIN users u ON c.created_by = u.id
		%s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query chains: %w", err)
	}
	defer rows.Close()

	var chains []models.Chain
	for rows.Next() {
		var chain models.Chain
		var template models.ChainTemplate
		var user models.User
		var chainDescription, chainID, genesisHash, tokenName sql.NullString
		var templateID sql.NullString
		var scheduledLaunchTime, actualLaunchTime, graduationTime sql.NullTime
		var blockTimeSeconds sql.NullInt32
		var upgradeBlockHeight sql.NullInt64
		var blockRewardAmount sql.NullFloat64
		var templateName, templateDescription, walletAddress, displayName sql.NullString

		err := rows.Scan(
			&chain.ID, &chain.ChainName, &tokenName, &chain.TokenSymbol, &chainDescription,
			&templateID, &chain.ConsensusMechanism, &chain.TokenTotalSupply,
			&blockTimeSeconds, &upgradeBlockHeight, &blockRewardAmount,
			&chain.GraduationThreshold, &chain.CreationFeeCNPY, &chain.InitialCNPYReserve,
			&chain.InitialTokenSupply, &chain.BondingCurveSlope, &scheduledLaunchTime,
			&actualLaunchTime, &chain.CreatorInitialPurchaseCNPY, &chain.Status,
			&chain.IsGraduated, &graduationTime, &chainID, &genesisHash,
			&chain.ValidatorMinStake, &chain.CreatedBy, &chain.CreatedAt, &chain.UpdatedAt,
			&templateName, &templateDescription,
			&walletAddress, &displayName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chain: %w", err)
		}

		// Handle nullable fields
		chain.TokenName = database.StringPtr(tokenName)
		chain.ChainDescription = database.StringPtr(chainDescription)
		chain.TemplateID = database.UUIDPtr(templateID)
		if blockTimeSeconds.Valid {
			val := int(blockTimeSeconds.Int32)
			chain.BlockTimeSeconds = &val
		}
		if upgradeBlockHeight.Valid {
			chain.UpgradeBlockHeight = &upgradeBlockHeight.Int64
		}
		if blockRewardAmount.Valid {
			chain.BlockRewardAmount = &blockRewardAmount.Float64
		}
		if scheduledLaunchTime.Valid {
			chain.ScheduledLaunchTime = &scheduledLaunchTime.Time
		}
		if actualLaunchTime.Valid {
			chain.ActualLaunchTime = &actualLaunchTime.Time
		}
		if graduationTime.Valid {
			chain.GraduationTime = &graduationTime.Time
		}
		chain.ChainID = database.StringPtr(chainID)
		chain.GenesisHash = database.StringPtr(genesisHash)

		// Set template and user if available
		if chain.TemplateID != nil {
			template.ID = *chain.TemplateID
			template.TemplateName = database.StringValue(templateName)
			template.TemplateDescription = database.StringValue(templateDescription)
			chain.Template = &template
		}
		user.ID = chain.CreatedBy
		user.WalletAddress = database.StringValue(walletAddress)
		user.DisplayName = database.StringPtr(displayName)
		chain.Creator = &user

		// Load additional relations if requested
		if err := r.loadChainRelations(ctx, &chain, filters.Include); err != nil {
			return nil, 0, fmt.Errorf("failed to load relations: %w", err)
		}

		chains = append(chains, chain)
	}

	return chains, total, nil
}

// ListByCreator retrieves chains by creator
func (r *chainRepository) ListByCreator(ctx context.Context, creatorID uuid.UUID, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	filters := interfaces.ChainFilters{CreatedBy: &creatorID}
	return r.List(ctx, filters, pagination)
}

// ListByTemplate retrieves chains by template
func (r *chainRepository) ListByTemplate(ctx context.Context, templateID uuid.UUID, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	filters := interfaces.ChainFilters{TemplateID: &templateID}
	return r.List(ctx, filters, pagination)
}

// ListByStatus retrieves chains by status
func (r *chainRepository) ListByStatus(ctx context.Context, status string, pagination interfaces.Pagination) ([]models.Chain, int, error) {
	filters := interfaces.ChainFilters{Status: status}
	return r.List(ctx, filters, pagination)
}

// Helper methods
func (r *chainRepository) getChainByField(ctx context.Context, field, value string) (*models.Chain, error) {
	query := fmt.Sprintf(`
		SELECT id, chain_name, token_name, token_symbol, chain_description, template_id, consensus_mechanism,
			   token_total_supply, block_time_seconds, upgrade_block_height, block_reward_amount,
			   graduation_threshold, creation_fee_cnpy, initial_cnpy_reserve,
			   initial_token_supply, bonding_curve_slope, scheduled_launch_time, actual_launch_time,
			   creator_initial_purchase_cnpy, status, is_graduated, graduation_time, chain_id,
			   genesis_hash, validator_min_stake, created_by, created_at, updated_at
		FROM chains WHERE %s = $1`, field)

	var chain models.Chain
	var chainDescription, tokenName sql.NullString
	var templateID sql.NullString
	var scheduledLaunchTime, actualLaunchTime, graduationTime sql.NullTime
	var blockTimeSeconds sql.NullInt32
	var upgradeBlockHeight sql.NullInt64
	var blockRewardAmount sql.NullFloat64
	var chainID, genesisHash sql.NullString

	err := r.db.QueryRowxContext(ctx, query, value).Scan(
		&chain.ID, &chain.ChainName, &tokenName, &chain.TokenSymbol, &chainDescription,
		&templateID, &chain.ConsensusMechanism, &chain.TokenTotalSupply,
		&blockTimeSeconds, &upgradeBlockHeight, &blockRewardAmount,
		&chain.GraduationThreshold, &chain.CreationFeeCNPY, &chain.InitialCNPYReserve,
		&chain.InitialTokenSupply, &chain.BondingCurveSlope, &scheduledLaunchTime,
		&actualLaunchTime, &chain.CreatorInitialPurchaseCNPY, &chain.Status,
		&chain.IsGraduated, &graduationTime, &chainID, &genesisHash,
		&chain.ValidatorMinStake, &chain.CreatedBy, &chain.CreatedAt, &chain.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain not found")
		}
		return nil, fmt.Errorf("failed to get chain: %w", err)
	}

	// Handle nullable fields
	chain.TokenName = database.StringPtr(tokenName)
	chain.ChainDescription = database.StringPtr(chainDescription)
	chain.TemplateID = database.UUIDPtr(templateID)
	if blockTimeSeconds.Valid {
		val := int(blockTimeSeconds.Int32)
		chain.BlockTimeSeconds = &val
	}
	if upgradeBlockHeight.Valid {
		chain.UpgradeBlockHeight = &upgradeBlockHeight.Int64
	}
	if blockRewardAmount.Valid {
		chain.BlockRewardAmount = &blockRewardAmount.Float64
	}
	if scheduledLaunchTime.Valid {
		chain.ScheduledLaunchTime = &scheduledLaunchTime.Time
	}
	if actualLaunchTime.Valid {
		chain.ActualLaunchTime = &actualLaunchTime.Time
	}
	if graduationTime.Valid {
		chain.GraduationTime = &graduationTime.Time
	}
	chain.ChainID = database.StringPtr(chainID)
	chain.GenesisHash = database.StringPtr(genesisHash)

	return &chain, nil
}

func (r *chainRepository) buildChainWhereClause(filters interfaces.ChainFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 0

	if filters.Status != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.status = $%d", argCount))
		args = append(args, filters.Status)
	}

	if filters.CreatedBy != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.created_by = $%d", argCount))
		args = append(args, *filters.CreatedBy)
	}

	if filters.TemplateID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("c.template_id = $%d", argCount))
		args = append(args, *filters.TemplateID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

func (r *chainRepository) loadChainRelations(ctx context.Context, chain *models.Chain, include []string) error {
	includeMap := make(map[string]bool)
	for _, rel := range include {
		includeMap[rel] = true
	}

	// Load template
	if includeMap["template"] || includeMap["templates"] {
		if chain.TemplateID != nil {
			template, err := r.getTemplateByID(ctx, *chain.TemplateID)
			if err != nil {
				return fmt.Errorf("failed to load template: %w", err)
			}
			chain.Template = template
		}
	}

	// Load creator
	if includeMap["creator"] || includeMap["user"] {
		user, err := r.getUserByID(ctx, chain.CreatedBy)
		if err != nil {
			return fmt.Errorf("failed to load creator: %w", err)
		}
		chain.Creator = user
	}

	// Load repository
	if includeMap["repository"] || includeMap["repositories"] {
		repo, err := r.GetRepositoryByChainID(ctx, chain.ID)
		if err != nil && err.Error() != "repository not found" {
			return fmt.Errorf("failed to load repository: %w", err)
		}
		chain.Repository = repo
	}

	// Load social links
	if includeMap["socials"] || includeMap["social_links"] {
		links, err := r.GetSocialLinksByChainID(ctx, chain.ID)
		if err != nil {
			return fmt.Errorf("failed to load social links: %w", err)
		}
		chain.SocialLinks = links
	}

	// Load assets
	if includeMap["assets"] {
		assets, err := r.GetAssetsByChainID(ctx, chain.ID)
		if err != nil {
			return fmt.Errorf("failed to load assets: %w", err)
		}
		chain.Assets = assets
	}

	// Note: Virtual pool loading removed - use VirtualPoolRepository directly instead
	// Virtual pools should be loaded separately through the VirtualPoolRepository

	return nil
}

// Additional helper methods that would be implemented...
func (r *chainRepository) getTemplateByID(ctx context.Context, id uuid.UUID) (*models.ChainTemplate, error) {
	if r.templateRepo == nil {
		return nil, fmt.Errorf("template repository not available")
	}
	return r.templateRepo.GetByID(ctx, id)
}

func (r *chainRepository) getUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if r.userRepo == nil {
		return nil, fmt.Errorf("user repository not available")
	}
	return r.userRepo.GetByID(ctx, id)
}

// Repository operations (simplified implementations)
func (r *chainRepository) CreateRepository(ctx context.Context, repo *models.ChainRepository) (*models.ChainRepository, error) {
	// Implementation for creating repository
	return nil, fmt.Errorf("not implemented")
}

func (r *chainRepository) UpdateRepository(ctx context.Context, repo *models.ChainRepository) (*models.ChainRepository, error) {
	query := `
		UPDATE chain_repositories SET
			github_url = $2,
			repository_name = $3,
			repository_owner = $4,
			default_branch = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE chain_id = $1
		RETURNING id, is_connected, oauth_token_hash, webhook_secret, auto_upgrade_enabled,
			upgrade_trigger, last_sync_commit_hash, last_sync_time, build_status,
			last_build_time, build_logs, created_at, updated_at`

	var oauthTokenHash, webhookSecret, lastSyncCommitHash, buildLogs sql.NullString
	var lastSyncTime, lastBuildTime sql.NullTime

	err := r.db.QueryRowxContext(ctx, query,
		repo.ChainID,
		repo.GithubURL,
		repo.RepositoryName,
		repo.RepositoryOwner,
		repo.DefaultBranch,
	).Scan(
		&repo.ID, &repo.IsConnected, &oauthTokenHash, &webhookSecret,
		&repo.AutoUpgradeEnabled, &repo.UpgradeTrigger, &lastSyncCommitHash,
		&lastSyncTime, &repo.BuildStatus, &lastBuildTime, &buildLogs,
		&repo.CreatedAt, &repo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to update repository: %w", err)
	}

	// Handle nullable fields
	repo.OAuthTokenHash = database.StringPtr(oauthTokenHash)
	repo.WebhookSecret = database.StringPtr(webhookSecret)
	repo.LastSyncCommitHash = database.StringPtr(lastSyncCommitHash)
	repo.BuildLogs = database.StringPtr(buildLogs)
	if lastSyncTime.Valid {
		repo.LastSyncTime = &lastSyncTime.Time
	}
	if lastBuildTime.Valid {
		repo.LastBuildTime = &lastBuildTime.Time
	}

	return repo, nil
}

func (r *chainRepository) GetRepositoryByChainID(ctx context.Context, chainID uuid.UUID) (*models.ChainRepository, error) {
	query := `
		SELECT id, chain_id, github_url, repository_name, repository_owner, default_branch,
			is_connected, oauth_token_hash, webhook_secret, auto_upgrade_enabled,
			upgrade_trigger, last_sync_commit_hash, last_sync_time, build_status,
			last_build_time, build_logs, created_at, updated_at
		FROM chain_repositories
		WHERE chain_id = $1`

	var repo models.ChainRepository
	var oauthTokenHash, webhookSecret, lastSyncCommitHash, buildLogs sql.NullString
	var lastSyncTime, lastBuildTime sql.NullTime

	err := r.db.QueryRowxContext(ctx, query, chainID).Scan(
		&repo.ID, &repo.ChainID, &repo.GithubURL, &repo.RepositoryName,
		&repo.RepositoryOwner, &repo.DefaultBranch, &repo.IsConnected,
		&oauthTokenHash, &webhookSecret, &repo.AutoUpgradeEnabled,
		&repo.UpgradeTrigger, &lastSyncCommitHash, &lastSyncTime,
		&repo.BuildStatus, &lastBuildTime, &buildLogs,
		&repo.CreatedAt, &repo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Handle nullable fields
	repo.OAuthTokenHash = database.StringPtr(oauthTokenHash)
	repo.WebhookSecret = database.StringPtr(webhookSecret)
	repo.LastSyncCommitHash = database.StringPtr(lastSyncCommitHash)
	repo.BuildLogs = database.StringPtr(buildLogs)
	if lastSyncTime.Valid {
		repo.LastSyncTime = &lastSyncTime.Time
	}
	if lastBuildTime.Valid {
		repo.LastBuildTime = &lastBuildTime.Time
	}

	return &repo, nil
}

func (r *chainRepository) DeleteRepository(ctx context.Context, chainID uuid.UUID) error {
	// Implementation for deleting repository
	return fmt.Errorf("not implemented")
}

// Social links operations (simplified implementations)
func (r *chainRepository) CreateSocialLinks(ctx context.Context, chainID uuid.UUID, links []models.ChainSocialLink) error {
	// Implementation for creating social links
	return fmt.Errorf("not implemented")
}

func (r *chainRepository) UpdateSocialLinks(ctx context.Context, chainID uuid.UUID, links []models.ChainSocialLink) error {
	// Implementation for updating social links
	return fmt.Errorf("not implemented")
}

func (r *chainRepository) GetSocialLinksByChainID(ctx context.Context, chainID uuid.UUID) ([]models.ChainSocialLink, error) {
	// Implementation for getting social links
	return nil, nil
}

func (r *chainRepository) DeleteSocialLinksByChainID(ctx context.Context, chainID uuid.UUID) error {
	// Implementation for deleting social links
	return fmt.Errorf("not implemented")
}

// Assets operations
func (r *chainRepository) CreateAssets(ctx context.Context, chainID uuid.UUID, assets []models.ChainAsset) error {
	if len(assets) == 0 {
		return nil
	}

	query := `
		INSERT INTO chain_assets (
			chain_id, asset_type, file_name, file_url, file_size_bytes, mime_type,
			title, description, alt_text, display_order, is_primary, is_featured,
			is_active, moderation_status, moderation_notes, uploaded_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	for _, asset := range assets {
		_, err := r.db.ExecContext(ctx, query,
			chainID,
			asset.AssetType,
			asset.FileName,
			asset.FileURL,
			database.NullInt64(asset.FileSizeBytes),
			database.NullString(asset.MimeType),
			database.NullString(asset.Title),
			database.NullString(asset.Description),
			database.NullString(asset.AltText),
			asset.DisplayOrder,
			asset.IsPrimary,
			asset.IsFeatured,
			asset.IsActive,
			asset.ModerationStatus,
			database.NullString(asset.ModerationNotes),
			asset.UploadedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to create asset: %w", err)
		}
	}

	return nil
}

func (r *chainRepository) UpdateAssets(ctx context.Context, chainID uuid.UUID, assets []models.ChainAsset) error {
	// Batch update implementation (currently not used by API but kept for interface compatibility)
	if len(assets) == 0 {
		return nil
	}

	for _, asset := range assets {
		err := r.UpdateAsset(ctx, &asset)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAssetByID gets a single asset by ID
func (r *chainRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (*models.ChainAsset, error) {
	query := `
		SELECT id, chain_id, asset_type, file_name, file_url, file_size_bytes, mime_type,
			title, description, alt_text, display_order, is_primary, is_featured, is_active,
			moderation_status, moderation_notes, uploaded_by, created_at, updated_at
		FROM chain_assets
		WHERE id = $1`

	var asset models.ChainAsset
	err := r.db.GetContext(ctx, &asset, query, assetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return &asset, nil
}

// UpdateAsset updates a single asset
func (r *chainRepository) UpdateAsset(ctx context.Context, asset *models.ChainAsset) error {
	query := `
		UPDATE chain_assets SET
			file_name = $2,
			file_url = $3,
			file_size_bytes = $4,
			mime_type = $5,
			title = $6,
			description = $7,
			alt_text = $8,
			display_order = $9,
			is_primary = $10,
			is_featured = $11,
			is_active = $12,
			moderation_status = $13,
			moderation_notes = $14,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		asset.ID,
		asset.FileName,
		asset.FileURL,
		database.NullInt64(asset.FileSizeBytes),
		database.NullString(asset.MimeType),
		database.NullString(asset.Title),
		database.NullString(asset.Description),
		database.NullString(asset.AltText),
		asset.DisplayOrder,
		asset.IsPrimary,
		asset.IsFeatured,
		asset.IsActive,
		asset.ModerationStatus,
		database.NullString(asset.ModerationNotes),
	).Scan(&asset.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("asset not found")
		}
		return fmt.Errorf("failed to update asset: %w", err)
	}

	return nil
}

func (r *chainRepository) GetAssetsByChainID(ctx context.Context, chainID uuid.UUID) ([]models.ChainAsset, error) {
	query := `
		SELECT id, chain_id, asset_type, file_name, file_url, file_size_bytes, mime_type,
			title, description, alt_text, display_order, is_primary, is_featured, is_active,
			moderation_status, moderation_notes, uploaded_by, created_at, updated_at
		FROM chain_assets
		WHERE chain_id = $1
		ORDER BY display_order ASC, created_at ASC`

	var assets []models.ChainAsset
	err := r.db.SelectContext(ctx, &assets, query, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	// Return empty slice if no assets found (not an error)
	if assets == nil {
		assets = []models.ChainAsset{}
	}

	return assets, nil
}

func (r *chainRepository) DeleteAssetsByChainID(ctx context.Context, chainID uuid.UUID) error {
	// Implementation for deleting assets
	return fmt.Errorf("not implemented")
}

// CreateChainKey creates a new encrypted key for a chain
func (r *chainRepository) CreateChainKey(ctx context.Context, key *models.ChainKey) (*models.ChainKey, error) {
	query := `
		INSERT INTO chain_keys (
			chain_id, address, public_key, encrypted_private_key, salt,
			key_nickname, key_purpose, is_active, rotation_count
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		key.ChainID,
		key.Address,
		key.PublicKey,
		key.EncryptedPrivateKey,
		key.Salt,
		database.NullString(key.KeyNickname),
		key.KeyPurpose,
		key.IsActive,
		key.RotationCount,
	).Scan(&key.ID, &key.CreatedAt, &key.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("chain key already exists for this chain and purpose")
		}
		return nil, fmt.Errorf("failed to create chain key: %w", err)
	}

	return key, nil
}

// GetChainKeyByChainID retrieves a chain key by chain ID and purpose
func (r *chainRepository) GetChainKeyByChainID(ctx context.Context, chainID uuid.UUID, purpose string) (*models.ChainKey, error) {
	query := `
		SELECT id, chain_id, address, public_key, encrypted_private_key, salt,
			key_nickname, key_purpose, is_active, last_used_at, rotation_count,
			created_at, updated_at
		FROM chain_keys
		WHERE chain_id = $1 AND key_purpose = $2 AND is_active = true`

	var key models.ChainKey
	err := r.db.GetContext(ctx, &key, query, chainID, purpose)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain key not found")
		}
		return nil, fmt.Errorf("failed to get chain key: %w", err)
	}

	return &key, nil
}

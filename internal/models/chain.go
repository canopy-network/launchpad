package models

import (
	"time"

	"github.com/google/uuid"
)

// Chain represents the main chain entity
type Chain struct {
	ID                         uuid.UUID  `json:"id" db:"id"`
	ChainName                  string     `json:"chain_name" db:"chain_name"`
	TokenName                  *string    `json:"token_name" db:"token_name"`
	TokenSymbol                string     `json:"token_symbol" db:"token_symbol"`
	ChainDescription           *string    `json:"chain_description" db:"chain_description"`
	TemplateID                 *uuid.UUID `json:"template_id" db:"template_id"`
	ConsensusMechanism         string     `json:"consensus_mechanism" db:"consensus_mechanism"`
	TokenTotalSupply           int64      `json:"token_total_supply" db:"token_total_supply"`
	BlockTimeSeconds           *int       `json:"block_time_seconds" db:"block_time_seconds"`
	UpgradeBlockHeight         *int64     `json:"upgrade_block_height" db:"upgrade_block_height"`
	BlockRewardAmount          *float64   `json:"block_reward_amount" db:"block_reward_amount"`
	GraduationThreshold        float64    `json:"graduation_threshold" db:"graduation_threshold"` // Amount in CNPY required for graduation
	CreationFeeCNPY            float64    `json:"creation_fee_cnpy" db:"creation_fee_cnpy"`
	InitialCNPYReserve         float64    `json:"initial_cnpy_reserve" db:"initial_cnpy_reserve"`
	InitialTokenSupply         int64      `json:"initial_token_supply" db:"initial_token_supply"`
	BondingCurveSlope          float64    `json:"bonding_curve_slope" db:"bonding_curve_slope"`
	ScheduledLaunchTime        *time.Time `json:"scheduled_launch_time" db:"scheduled_launch_time"`
	ActualLaunchTime           *time.Time `json:"actual_launch_time" db:"actual_launch_time"`
	CreatorInitialPurchaseCNPY float64    `json:"creator_initial_purchase_cnpy" db:"creator_initial_purchase_cnpy"`
	Status                     string     `json:"status" db:"status"`
	IsGraduated                bool       `json:"is_graduated" db:"is_graduated"`
	GraduationTime             *time.Time `json:"graduation_time" db:"graduation_time"`
	ChainID                    *string    `json:"chain_id" db:"chain_id"`
	GenesisHash                *string    `json:"genesis_hash" db:"genesis_hash"`
	ValidatorMinStake          float64    `json:"validator_min_stake" db:"validator_min_stake"`
	CreatedBy                  uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt                  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at" db:"updated_at"`

	// Relationships (populated when requested)
	Template    *ChainTemplate    `json:"template,omitempty"`
	Creator     *User             `json:"creator,omitempty"`
	Repository  *ChainRepository  `json:"repository,omitempty"`
	SocialLinks []ChainSocialLink `json:"social_links,omitempty"`
	Assets      []ChainAsset      `json:"assets,omitempty"`
	VirtualPool *VirtualPool      `json:"virtual_pool,omitempty"`
}

// ChainTemplate represents pre-built blockchain templates
type ChainTemplate struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	TemplateName        string    `json:"template_name" db:"template_name"`
	TemplateDescription string    `json:"template_description" db:"template_description"`
	TemplateCategory    string    `json:"template_category" db:"template_category"`
	SupportedLanguage   string    `json:"supported_language" db:"supported_language"`
	DefaultTokenSupply  int64     `json:"default_token_supply" db:"default_token_supply"`
	DocumentationURL    *string   `json:"documentation_url" db:"documentation_url"`
	ExampleChains       []string  `json:"example_chains" db:"example_chains"`
	Version             string    `json:"version" db:"version"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// ChainRepository represents GitHub repository connections
type ChainRepository struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	ChainID            uuid.UUID  `json:"chain_id" db:"chain_id"`
	GithubURL          string     `json:"github_url" db:"github_url"`
	RepositoryName     string     `json:"repository_name" db:"repository_name"`
	RepositoryOwner    string     `json:"repository_owner" db:"repository_owner"`
	DefaultBranch      string     `json:"default_branch" db:"default_branch"`
	IsConnected        bool       `json:"is_connected" db:"is_connected"`
	OAuthTokenHash     *string    `json:"-" db:"oauth_token_hash"` // Hidden from JSON
	WebhookSecret      *string    `json:"-" db:"webhook_secret"`   // Hidden from JSON
	AutoUpgradeEnabled bool       `json:"auto_upgrade_enabled" db:"auto_upgrade_enabled"`
	UpgradeTrigger     string     `json:"upgrade_trigger" db:"upgrade_trigger"`
	LastSyncCommitHash *string    `json:"last_sync_commit_hash" db:"last_sync_commit_hash"`
	LastSyncTime       *time.Time `json:"last_sync_time" db:"last_sync_time"`
	BuildStatus        string     `json:"build_status" db:"build_status"`
	LastBuildTime      *time.Time `json:"last_build_time" db:"last_build_time"`
	BuildLogs          *string    `json:"build_logs,omitempty" db:"build_logs"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// ChainSocialLink represents social media and external links
type ChainSocialLink struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	ChainID           uuid.UUID  `json:"chain_id" db:"chain_id"`
	Platform          string     `json:"platform" db:"platform"`
	URL               string     `json:"url" db:"url"`
	DisplayName       *string    `json:"display_name" db:"display_name"`
	IsVerified        bool       `json:"is_verified" db:"is_verified"`
	FollowerCount     int        `json:"follower_count" db:"follower_count"`
	LastMetricsUpdate *time.Time `json:"last_metrics_update" db:"last_metrics_update"`
	DisplayOrder      int        `json:"display_order" db:"display_order"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// ChainAsset represents media assets and files
type ChainAsset struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ChainID          uuid.UUID `json:"chain_id" db:"chain_id"`
	AssetType        string    `json:"asset_type" db:"asset_type"`
	FileName         string    `json:"file_name" db:"file_name"`
	FileURL          string    `json:"file_url" db:"file_url"`
	FileSizeBytes    *int64    `json:"file_size_bytes" db:"file_size_bytes"`
	MimeType         *string   `json:"mime_type" db:"mime_type"`
	Title            *string   `json:"title" db:"title"`
	Description      *string   `json:"description" db:"description"`
	AltText          *string   `json:"alt_text" db:"alt_text"`
	DisplayOrder     int       `json:"display_order" db:"display_order"`
	IsPrimary        bool      `json:"is_primary" db:"is_primary"`
	IsFeatured       bool      `json:"is_featured" db:"is_featured"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	ModerationStatus string    `json:"moderation_status" db:"moderation_status"`
	ModerationNotes  *string   `json:"moderation_notes" db:"moderation_notes"`
	UploadedBy       uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// VirtualPool represents virtual liquidity pool state
type VirtualPool struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	ChainID               uuid.UUID `json:"chain_id" db:"chain_id"`
	CNPYReserve           float64   `json:"cnpy_reserve" db:"cnpy_reserve"`
	TokenReserve          int64     `json:"token_reserve" db:"token_reserve"`
	CurrentPriceCNPY      float64   `json:"current_price_cnpy" db:"current_price_cnpy"`
	MarketCapUSD          float64   `json:"market_cap_usd" db:"market_cap_usd"`
	TotalVolumeCNPY       float64   `json:"total_volume_cnpy" db:"total_volume_cnpy"`
	TotalTransactions     int       `json:"total_transactions" db:"total_transactions"`
	UniqueTraders         int       `json:"unique_traders" db:"unique_traders"`
	IsActive              bool      `json:"is_active" db:"is_active"`
	Price24hChangePercent float64   `json:"price_24h_change_percent" db:"price_24h_change_percent"`
	Volume24hCNPY         float64   `json:"volume_24h_cnpy" db:"volume_24h_cnpy"`
	High24hCNPY           float64   `json:"high_24h_cnpy" db:"high_24h_cnpy"`
	Low24hCNPY            float64   `json:"low_24h_cnpy" db:"low_24h_cnpy"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// VirtualPoolTransaction represents individual trading transactions
type VirtualPoolTransaction struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	VirtualPoolID         uuid.UUID `json:"virtual_pool_id" db:"virtual_pool_id"`
	ChainID               uuid.UUID `json:"chain_id" db:"chain_id"`
	UserID                uuid.UUID `json:"user_id" db:"user_id"`
	TransactionType       string    `json:"transaction_type" db:"transaction_type"`
	CNPYAmount            float64   `json:"cnpy_amount" db:"cnpy_amount"`
	TokenAmount           int64     `json:"token_amount" db:"token_amount"`
	PricePerTokenCNPY     float64   `json:"price_per_token_cnpy" db:"price_per_token_cnpy"`
	TradingFeeCNPY        float64   `json:"trading_fee_cnpy" db:"trading_fee_cnpy"`
	SlippagePercent       float64   `json:"slippage_percent" db:"slippage_percent"`
	TransactionHash       *string   `json:"transaction_hash" db:"transaction_hash"`
	BlockHeight           *int64    `json:"block_height" db:"block_height"`
	GasUsed               *int      `json:"gas_used" db:"gas_used"`
	PoolCNPYReserveAfter  float64   `json:"pool_cnpy_reserve_after" db:"pool_cnpy_reserve_after"`
	PoolTokenReserveAfter int64     `json:"pool_token_reserve_after" db:"pool_token_reserve_after"`
	MarketCapAfterUSD     float64   `json:"market_cap_after_usd" db:"market_cap_after_usd"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}

// Chain status constants
const (
	ChainStatusDraft         = "draft"
	ChainStatusPendingLaunch = "pending_launch"
	ChainStatusVirtualActive = "virtual_active"
	ChainStatusGraduated     = "graduated"
	ChainStatusFailed        = "failed"
)

// Asset type constants
const (
	AssetTypeLogo          = "logo"
	AssetTypeBanner        = "banner"
	AssetTypeScreenshot    = "screenshot"
	AssetTypeVideo         = "video"
	AssetTypeWhitepaper    = "whitepaper"
	AssetTypeDocumentation = "documentation"
)

// Social platform constants
const (
	PlatformTwitter       = "twitter"
	PlatformTelegram      = "telegram"
	PlatformDiscord       = "discord"
	PlatformWebsite       = "website"
	PlatformWhitepaper    = "whitepaper"
	PlatformDocumentation = "documentation"
	PlatformMedium        = "medium"
	PlatformYoutube       = "youtube"
)

// Build status constants
const (
	BuildStatusPending  = "pending"
	BuildStatusBuilding = "building"
	BuildStatusSuccess  = "success"
	BuildStatusFailed   = "failed"
)

// Upgrade trigger constants
const (
	UpgradeTriggerTagRelease = "tag_release"
	UpgradeTriggerMainPush   = "main_push"
	UpgradeTriggerManual     = "manual"
)

// ChainKey represents an encrypted cryptographic key for a chain
type ChainKey struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	ChainID             uuid.UUID  `json:"chain_id" db:"chain_id"`
	Address             string     `json:"address" db:"address"`
	PublicKey           []byte     `json:"public_key" db:"public_key"`
	EncryptedPrivateKey string     `json:"encrypted_private_key" db:"encrypted_private_key"`
	Salt                []byte     `json:"salt" db:"salt"`
	KeyNickname         *string    `json:"key_nickname" db:"key_nickname"`
	KeyPurpose          string     `json:"key_purpose" db:"key_purpose"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	LastUsedAt          *time.Time `json:"last_used_at" db:"last_used_at"`
	RotationCount       int        `json:"rotation_count" db:"rotation_count"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// Key purpose constants
const (
	KeyPurposeChainOperation = "chain_operation"
	KeyPurposeGovernance     = "governance"
	KeyPurposeTreasury       = "treasury"
	KeyPurposeBackup         = "backup"
)

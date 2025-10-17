package models

// CreateChainRequest represents the request payload for creating a new chain
type CreateChainRequest struct {
	// Basic chain information
	ChainName        string  `json:"chain_name" validate:"required,min=1,max=100"`
	TokenName        *string `json:"token_name" validate:"omitempty,min=1,max=100"`
	TokenSymbol      string  `json:"token_symbol" validate:"required,min=1,max=20,uppercase"`
	ChainDescription *string `json:"chain_description" validate:"omitempty,max=5000"`

	// Template and configuration
	TemplateID         *string `json:"template_id" validate:"omitempty,uuid"`
	ConsensusMechanism string  `json:"consensus_mechanism" validate:"omitempty,max=50"`

	// Economic parameters
	TokenTotalSupply    *int64   `json:"token_total_supply" validate:"omitempty,min=1000000,max=1000000000000"`
	BlockTimeSeconds    *int     `json:"block_time_seconds" validate:"omitempty,oneof=5 10 20 30 60 120 300 600 1800"`
	UpgradeBlockHeight  *int64   `json:"upgrade_block_height" validate:"omitempty,min=1"`
	BlockRewardAmount   *float64 `json:"block_reward_amount" validate:"omitempty,min=0"`
	GraduationThreshold *float64 `json:"graduation_threshold" validate:"omitempty,min=1000,max=10000000"`
	CreationFeeCNPY     *float64 `json:"creation_fee_cnpy" validate:"omitempty,min=0"`

	// Bonding curve parameters
	InitialCNPYReserve         *float64 `json:"initial_cnpy_reserve" validate:"omitempty,min=1000"`
	InitialTokenSupply         *int64   `json:"initial_token_supply" validate:"omitempty,min=100000"`
	BondingCurveSlope          *float64 `json:"bonding_curve_slope" validate:"omitempty,min=0.000000001"`
	ValidatorMinStake          *float64 `json:"validator_min_stake" validate:"omitempty,min=100"`
	CreatorInitialPurchaseCNPY *float64 `json:"creator_initial_purchase_cnpy" validate:"omitempty,min=0"`

	// GitHub repository
	GithubURL *string `json:"github_url" validate:"omitempty,url"`

	// Social links
	TwitterURL  *string `json:"twitter_url" validate:"omitempty,url"`
	TelegramURL *string `json:"telegram_url" validate:"omitempty,url"`
	WebsiteURL  *string `json:"website_url" validate:"omitempty,url"`

	// Assets
	WhitepaperURL *string `json:"whitepaper_url" validate:"omitempty,url"`
	TokenImageURL *string `json:"token_image_url" validate:"omitempty,url"`
	TokenVideoURL *string `json:"token_video_url" validate:"omitempty,url"`
}

// Query parameter structures
type ChainsQueryParams struct {
	Status     string `form:"status" validate:"omitempty,oneof=draft pending_launch virtual_active graduated failed"`
	CreatedBy  string `form:"created_by" validate:"omitempty,uuid"`
	TemplateID string `form:"template_id" validate:"omitempty,uuid"`
	Page       int    `form:"page" validate:"omitempty,min=1"`
	Limit      int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Include    string `form:"include" validate:"omitempty"`
}

type TemplatesQueryParams struct {
	Category        string `form:"category" validate:"omitempty,max=50"`
	ComplexityLevel string `form:"complexity_level" validate:"omitempty,oneof=beginner intermediate advanced expert"`
	IsActive        *bool  `form:"is_active"`
	Page            int    `form:"page" validate:"omitempty,min=1"`
	Limit           int    `form:"limit" validate:"omitempty,min=1,max=100"`
}

type TransactionsQueryParams struct {
	UserID          string `form:"user_id" validate:"omitempty,uuid"`
	TransactionType string `form:"transaction_type" validate:"omitempty,oneof=buy sell"`
	Page            int    `form:"page" validate:"omitempty,min=1"`
	Limit           int    `form:"limit" validate:"omitempty,min=1,max=100"`
}

type VirtualPoolsQueryParams struct {
	Page  int `form:"page" validate:"omitempty,min=1"`
	Limit int `form:"limit" validate:"omitempty,min=1,max=100"`
}

// EmailAuthRequest represents the request payload for email authentication
type EmailAuthRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyEmailCodeRequest represents the request payload for verifying email code
type VerifyEmailCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}

// CreateWalletRequest represents the request payload for creating a new wallet
type CreateWalletRequest struct {
	Password          string  `json:"password" validate:"required,min=8"`
	UserID            *string `json:"user_id" validate:"omitempty,uuid"`
	ChainID           *string `json:"chain_id" validate:"omitempty,uuid"`
	WalletName        *string `json:"wallet_name" validate:"omitempty,min=1,max=100"`
	WalletDescription *string `json:"wallet_description" validate:"omitempty,max=500"`
}

// DecryptWalletRequest represents the request payload for decrypting a wallet
type DecryptWalletRequest struct {
	Password string `json:"password" validate:"required"`
}

// UpdateWalletRequest represents the request payload for updating wallet metadata
type UpdateWalletRequest struct {
	WalletName        *string `json:"wallet_name" validate:"omitempty,min=1,max=100"`
	WalletDescription *string `json:"wallet_description" validate:"omitempty,max=500"`
	IsActive          *bool   `json:"is_active"`
}

// WalletsQueryParams represents query parameters for listing wallets
type WalletsQueryParams struct {
	UserID    string `form:"user_id" validate:"omitempty,uuid"`
	ChainID   string `form:"chain_id" validate:"omitempty,uuid"`
	IsActive  *bool  `form:"is_active"`
	IsLocked  *bool  `form:"is_locked"`
	CreatedBy string `form:"created_by" validate:"omitempty,uuid"`
	Page      int    `form:"page" validate:"omitempty,min=1"`
	Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`
}

// UpdateChainRepositoryRequest represents the request payload for updating a chain repository
// All fields are optional to support partial updates
type UpdateChainRepositoryRequest struct {
	GithubURL       *string `json:"github_url" validate:"omitempty,url"`
	RepositoryName  *string `json:"repository_name" validate:"omitempty,min=1,max=100"`
	RepositoryOwner *string `json:"repository_owner" validate:"omitempty,min=1,max=100"`
	DefaultBranch   *string `json:"default_branch" validate:"omitempty,min=1,max=100"`
}

// CreateChainAssetRequest represents the request payload for creating a new chain asset
type CreateChainAssetRequest struct {
	AssetType     string  `json:"asset_type" validate:"required,oneof=logo banner screenshot video whitepaper documentation"`
	FileName      string  `json:"file_name" validate:"required,min=1,max=255"`
	FileURL       string  `json:"file_url" validate:"required,url"`
	FileSizeBytes *int64  `json:"file_size_bytes" validate:"omitempty,min=1"`
	MimeType      *string `json:"mime_type" validate:"omitempty,max=100"`
	Title         *string `json:"title" validate:"omitempty,max=200"`
	Description   *string `json:"description" validate:"omitempty,max=1000"`
	AltText       *string `json:"alt_text" validate:"omitempty,max=500"`
	DisplayOrder  *int    `json:"display_order" validate:"omitempty,min=0"`
	IsPrimary     *bool   `json:"is_primary"`
	IsFeatured    *bool   `json:"is_featured"`
}

// UpdateChainAssetRequest represents the request payload for updating a chain asset
// All fields are optional to support partial updates
type UpdateChainAssetRequest struct {
	FileName         *string `json:"file_name" validate:"omitempty,min=1,max=255"`
	FileURL          *string `json:"file_url" validate:"omitempty,url"`
	FileSizeBytes    *int64  `json:"file_size_bytes" validate:"omitempty,min=1"`
	MimeType         *string `json:"mime_type" validate:"omitempty,max=100"`
	Title            *string `json:"title" validate:"omitempty,max=200"`
	Description      *string `json:"description" validate:"omitempty,max=1000"`
	AltText          *string `json:"alt_text" validate:"omitempty,max=500"`
	DisplayOrder     *int    `json:"display_order" validate:"omitempty,min=0"`
	IsPrimary        *bool   `json:"is_primary"`
	IsFeatured       *bool   `json:"is_featured"`
	IsActive         *bool   `json:"is_active"`
	ModerationStatus *string `json:"moderation_status" validate:"omitempty,oneof=pending approved rejected"`
	ModerationNotes  *string `json:"moderation_notes" validate:"omitempty,max=1000"`
}

// UpdateChainDescriptionRequest represents the request payload for updating a chain's description
type UpdateChainDescriptionRequest struct {
	ChainDescription string `json:"chain_description" validate:"required,max=5000"`
}

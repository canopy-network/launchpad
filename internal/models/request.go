package models

// CreateChainRequest represents the request payload for creating a new chain
type CreateChainRequest struct {
	ChainName                  string   `json:"chain_name" validate:"required,min=1,max=100"`
	TokenSymbol                string   `json:"token_symbol" validate:"required,min=1,max=20,uppercase"`
	ChainDescription           *string  `json:"chain_description" validate:"omitempty,max=5000"`
	TemplateID                 *string  `json:"template_id" validate:"omitempty,uuid"`
	ConsensusMechanism         string   `json:"consensus_mechanism" validate:"omitempty,max=50"`
	TokenTotalSupply           *int64   `json:"token_total_supply" validate:"omitempty,min=1000000,max=1000000000000"`
	GraduationThreshold        *float64 `json:"graduation_threshold" validate:"omitempty,min=1000,max=10000000"` // Amount in CNPY required for graduation
	CreationFeeCNPY            *float64 `json:"creation_fee_cnpy" validate:"omitempty,min=0"`
	InitialCNPYReserve         *float64 `json:"initial_cnpy_reserve" validate:"omitempty,min=1000"`
	InitialTokenSupply         *int64   `json:"initial_token_supply" validate:"omitempty,min=100000"`
	BondingCurveSlope          *float64 `json:"bonding_curve_slope" validate:"omitempty,min=0.000000001"`
	ValidatorMinStake          *float64 `json:"validator_min_stake" validate:"omitempty,min=100"`
	CreatorInitialPurchaseCNPY *float64 `json:"creator_initial_purchase_cnpy" validate:"omitempty,min=0"`
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

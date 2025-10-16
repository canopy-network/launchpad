package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents user accounts and authentication data
type User struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	WalletAddress           string                 `json:"wallet_address" db:"wallet_address"`
	Email                   *string                `json:"email" db:"email"`
	Username                *string                `json:"username" db:"username"`
	DisplayName             *string                `json:"display_name" db:"display_name"`
	Bio                     *string                `json:"bio" db:"bio"`
	AvatarURL               *string                `json:"avatar_url" db:"avatar_url"`
	WebsiteURL              *string                `json:"website_url" db:"website_url"`
	TwitterHandle           *string                `json:"twitter_handle" db:"twitter_handle"`
	GithubUsername          *string                `json:"github_username" db:"github_username"`
	TelegramHandle          *string                `json:"telegram_handle" db:"telegram_handle"`
	IsVerified              bool                   `json:"is_verified" db:"is_verified"`
	VerificationTier        string                 `json:"verification_tier" db:"verification_tier"`
	EmailVerifiedAt         *time.Time             `json:"email_verified_at" db:"email_verified_at"`
	JWTVersion              int                    `json:"-" db:"jwt_version"` // Don't expose in JSON
	TotalChainsCreated      int                    `json:"total_chains_created" db:"total_chains_created"`
	TotalCNPYInvested       float64                `json:"total_cnpy_invested" db:"total_cnpy_invested"`
	ReputationScore         int                    `json:"reputation_score" db:"reputation_score"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time              `json:"updated_at" db:"updated_at"`
	LastActiveAt            *time.Time             `json:"last_active_at" db:"last_active_at"`
}

// UserVirtualPosition represents user positions and holdings in virtual pools
type UserVirtualPosition struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserID                uuid.UUID  `json:"user_id" db:"user_id"`
	ChainID               uuid.UUID  `json:"chain_id" db:"chain_id"`
	VirtualPoolID         uuid.UUID  `json:"virtual_pool_id" db:"virtual_pool_id"`
	TokenBalance          int64      `json:"token_balance" db:"token_balance"`
	TotalCNPYInvested     float64    `json:"total_cnpy_invested" db:"total_cnpy_invested"`
	TotalCNPYWithdrawn    float64    `json:"total_cnpy_withdrawn" db:"total_cnpy_withdrawn"`
	AverageEntryPriceCNPY float64    `json:"average_entry_price_cnpy" db:"average_entry_price_cnpy"`
	UnrealizedPnlCNPY     float64    `json:"unrealized_pnl_cnpy" db:"unrealized_pnl_cnpy"`
	RealizedPnlCNPY       float64    `json:"realized_pnl_cnpy" db:"realized_pnl_cnpy"`
	TotalReturnPercent    float64    `json:"total_return_percent" db:"total_return_percent"`
	IsActive              bool       `json:"is_active" db:"is_active"`
	FirstPurchaseAt       *time.Time `json:"first_purchase_at" db:"first_purchase_at"`
	LastActivityAt        *time.Time `json:"last_activity_at" db:"last_activity_at"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// Verification tier constants
const (
	VerificationTierBasic    = "basic"
	VerificationTierVerified = "verified"
	VerificationTierPremium  = "premium"
)

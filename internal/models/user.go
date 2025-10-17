package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents user accounts and authentication data
type User struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	WalletAddress      string     `json:"wallet_address" db:"wallet_address"`
	Email              *string    `json:"email" db:"email"`
	Username           *string    `json:"username" db:"username"`
	DisplayName        *string    `json:"display_name" db:"display_name"`
	Bio                *string    `json:"bio" db:"bio"`
	AvatarURL          *string    `json:"avatar_url" db:"avatar_url"`
	WebsiteURL         *string    `json:"website_url" db:"website_url"`
	TwitterHandle      *string    `json:"twitter_handle" db:"twitter_handle"`
	GithubUsername     *string    `json:"github_username" db:"github_username"`
	TelegramHandle     *string    `json:"telegram_handle" db:"telegram_handle"`
	IsVerified         bool       `json:"is_verified" db:"is_verified"`
	VerificationTier   string     `json:"verification_tier" db:"verification_tier"`
	EmailVerifiedAt    *time.Time `json:"email_verified_at" db:"email_verified_at"`
	JWTVersion         int        `json:"-" db:"jwt_version"` // Don't expose in JSON
	TotalChainsCreated int        `json:"total_chains_created" db:"total_chains_created"`
	TotalCNPYInvested  float64    `json:"total_cnpy_invested" db:"total_cnpy_invested"`
	ReputationScore    int        `json:"reputation_score" db:"reputation_score"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	LastActiveAt       *time.Time `json:"last_active_at" db:"last_active_at"`
}

// UpdateProfileRequest represents a request to update user profile
// All fields are optional pointers to support partial updates
type UpdateProfileRequest struct {
	Username       *string `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	DisplayName    *string `json:"display_name,omitempty" validate:"omitempty,max=100"`
	Bio            *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	AvatarURL      *string `json:"avatar_url,omitempty" validate:"omitempty,url,max=500"`
	WebsiteURL     *string `json:"website_url,omitempty" validate:"omitempty,url,max=500"`
	TwitterHandle  *string `json:"twitter_handle,omitempty" validate:"omitempty,max=50"`
	GithubUsername *string `json:"github_username,omitempty" validate:"omitempty,max=100,alphanum"`
	TelegramHandle *string `json:"telegram_handle,omitempty" validate:"omitempty,max=50"`
}

// Verification tier constants
const (
	VerificationTierBasic    = "basic"
	VerificationTierVerified = "verified"
	VerificationTierPremium  = "premium"
)

package models

import (
	"github.com/google/uuid"
	"time"
)

// GraduatedPool represents graduated liquidity pool state
type GraduatedPool struct {
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

// GraduatedPoolTransaction represents graduated trading transactions (swaps, liquidity deposits and withdrawals)
type GraduatedPoolTransaction struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	GraduatedPoolID       uuid.UUID `json:"graduated_pool_id" db:"graduated_pool_id"`
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
	GasUsed               *int      `json:"gas_used" db:"gas_used"` // TODO 'gas' is wrong - this isn't ethereum
	PoolCNPYReserveAfter  float64   `json:"pool_cnpy_reserve_after" db:"pool_cnpy_reserve_after"`
	PoolTokenReserveAfter int64     `json:"pool_token_reserve_after" db:"pool_token_reserve_after"`
	MarketCapAfterUSD     float64   `json:"market_cap_after_usd" db:"market_cap_after_usd"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}

// UserGraduatedLPPosition represents user liquidity position in a graduated pools
type UserGraduatedLPPosition struct {
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

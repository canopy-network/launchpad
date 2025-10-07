package bondingcurve

import (
	"errors"
	"math/big"
)

// VirtualPool represents the state of a bonding curve virtual pool
type VirtualPool struct {
	CNPYReserve  *big.Float `json:"cnpy_reserve"`  // x - CNPY reserve in the pool
	TokenReserve *big.Float `json:"token_reserve"` // y - token reserve in the pool
	TotalSupply  *big.Float `json:"total_supply"`  // total tokens minted
}

// TradeResult represents the result of a buy or sell operation
type TradeResult struct {
	AmountOut       *big.Float `json:"amount_out"`        // tokens received (buy) or CNPY received (sell)
	NewCNPYReserve  *big.Float `json:"new_cnpy_reserve"`  // updated CNPY reserve after trade
	NewTokenReserve *big.Float `json:"new_token_reserve"` // updated token reserve after trade
	NewTotalSupply  *big.Float `json:"new_total_supply"`  // updated total supply after trade
	Price           *big.Float `json:"price"`             // effective price for this trade
	PriceImpact     *big.Float `json:"price_impact"`      // price impact percentage
}

// BondingCurveConfig contains configuration for the bonding curve
type BondingCurveConfig struct {
	FeeRateBasisPoints uint64     `json:"fee_rate_basis_points"` // Fee rate in basis points (e.g., 100 = 1%)
	InitialPrice       *big.Float `json:"initial_price"`         // Initial price when pool starts at 0 (CNPY per token)
}

// Constants for the bonding curve
const (
	// DefaultFeeRateBasisPoints is the default fee rate (1%)
	DefaultFeeRateBasisPoints = 100

	// DefaultInitialPrice is the starting price when pool is at 0 (0.01 CNPY per token)
	DefaultInitialPrice = 0.01

	// BasisPointsDivisor for converting basis points to decimal
	BasisPointsDivisor = 10000

	// CurveDivisor used in the bonding curve formula (1000)
	CurveDivisor = 1000

	// Precision for big.Float operations
	Precision = 256
)

// Common errors
var (
	ErrInsufficientReserve = errors.New("insufficient reserve for trade")
	ErrInvalidAmount       = errors.New("invalid trade amount")
	ErrZeroAmount          = errors.New("trade amount must be greater than zero")
	ErrInsufficientTokens  = errors.New("insufficient tokens for sell")
	ErrPoolNotInitialized  = errors.New("virtual pool not properly initialized")
)

// NewVirtualPool creates a new virtual pool with initial reserves
func NewVirtualPool(cnpyReserve, tokenReserve, totalSupply *big.Float) *VirtualPool {
	return &VirtualPool{
		CNPYReserve:  new(big.Float).Copy(cnpyReserve),
		TokenReserve: new(big.Float).Copy(tokenReserve),
		TotalSupply:  new(big.Float).Copy(totalSupply),
	}
}

// Copy creates a deep copy of the virtual pool
func (vp *VirtualPool) Copy() *VirtualPool {
	return &VirtualPool{
		CNPYReserve:  new(big.Float).Copy(vp.CNPYReserve),
		TokenReserve: new(big.Float).Copy(vp.TokenReserve),
		TotalSupply:  new(big.Float).Copy(vp.TotalSupply),
	}
}

// Validate checks if the virtual pool is in a valid state
func (vp *VirtualPool) Validate() error {
	if vp.CNPYReserve == nil || vp.TokenReserve == nil || vp.TotalSupply == nil {
		return ErrPoolNotInitialized
	}

	// For trading, we need positive reserves (unless both are 0 for initial state)
	bothZero := vp.CNPYReserve.Sign() == 0 && vp.TokenReserve.Sign() == 0
	if !bothZero && (vp.CNPYReserve.Sign() <= 0 || vp.TokenReserve.Sign() <= 0) {
		return ErrInsufficientReserve
	}

	if vp.TotalSupply.Sign() < 0 {
		return ErrPoolNotInitialized
	}

	return nil
}

// CurrentPrice calculates the current price of tokens in terms of CNPY
func (vp *VirtualPool) CurrentPrice() *big.Float {
	if vp.TokenReserve.Sign() == 0 {
		return big.NewFloat(0)
	}

	// Price = CNPY reserve / Token reserve
	price := new(big.Float).Quo(vp.CNPYReserve, vp.TokenReserve)
	return price
}

// NewBondingCurveConfig creates a new bonding curve configuration with default values
func NewBondingCurveConfig() *BondingCurveConfig {
	return &BondingCurveConfig{
		FeeRateBasisPoints: DefaultFeeRateBasisPoints,
		InitialPrice:       big.NewFloat(DefaultInitialPrice),
	}
}

// ApplyFee calculates the amount after applying fees
func (config *BondingCurveConfig) ApplyFee(amount *big.Float) *big.Float {
	if config.FeeRateBasisPoints == 0 {
		return new(big.Float).Copy(amount)
	}

	// Calculate fee: amount * (BasisPointsDivisor - FeeRateBasisPoints) / BasisPointsDivisor
	feeMultiplier := new(big.Float).SetUint64(BasisPointsDivisor - config.FeeRateBasisPoints)
	divisor := new(big.Float).SetUint64(BasisPointsDivisor)

	amountAfterFee := new(big.Float).Mul(amount, feeMultiplier)
	amountAfterFee.Quo(amountAfterFee, divisor)

	return amountAfterFee
}

// CalculateFee calculates the fee amount for a given input
func (config *BondingCurveConfig) CalculateFee(amount *big.Float) *big.Float {
	if config.FeeRateBasisPoints == 0 {
		return big.NewFloat(0)
	}

	// Calculate fee: amount * FeeRateBasisPoints / BasisPointsDivisor
	feeRate := new(big.Float).SetUint64(config.FeeRateBasisPoints)
	divisor := new(big.Float).SetUint64(BasisPointsDivisor)

	fee := new(big.Float).Mul(amount, feeRate)
	fee.Quo(fee, divisor)

	return fee
}

package bondingcurve

import (
	"math/big"
)

// BondingCurve implements the sum-product bonding curve mechanics
type BondingCurve struct {
	config *BondingCurveConfig
}

// NewBondingCurve creates a new bonding curve instance
func NewBondingCurve(config *BondingCurveConfig) *BondingCurve {
	if config == nil {
		config = NewBondingCurveConfig()
	}
	return &BondingCurve{
		config: config,
	}
}

// Buy executes a buy transaction (minting tokens)
// Pump.fun sum-style bonding curve formula: dY = (amountInWithFee * y) / (x + amountInWithFee)
// Where x = CNPY reserve, y = token reserve, dY = tokens to mint
// Special case: When both reserves are 0, uses fixed initial price
func (bc *BondingCurve) Buy(pool *VirtualPool, cnpyAmountIn *big.Float) (*TradeResult, error) {
	if err := pool.Validate(); err != nil {
		return nil, err
	}

	if cnpyAmountIn == nil || cnpyAmountIn.Sign() <= 0 {
		return nil, ErrZeroAmount
	}

	// Get current pool state
	x := pool.CNPYReserve  // CNPY reserve
	y := pool.TokenReserve // Token reserve

	var tokensBeforeFee *big.Float
	var priceBefore *big.Float

	// Special case: Pool starts at 0, use fixed initial price
	if x.Sign() == 0 && y.Sign() == 0 {
		// tokensOut = cnpyAmountIn / initialPrice
		tokensBeforeFee = new(big.Float).Quo(cnpyAmountIn, bc.config.InitialPrice)
		priceBefore = new(big.Float).Copy(bc.config.InitialPrice)
	} else {
		// Calculate price before trade for price impact
		priceBefore = pool.CurrentPrice()

		// Calculate tokens using full input: dY = (cnpyAmountIn * y) / (x + cnpyAmountIn)
		// Numerator: cnpyAmountIn * y
		numerator := new(big.Float).Mul(cnpyAmountIn, y)

		// Denominator: x + cnpyAmountIn
		denominator := new(big.Float).Add(x, cnpyAmountIn)

		// Prevent division by zero
		if denominator.Sign() == 0 {
			return nil, ErrInsufficientReserve
		}

		// Calculate tokens before fee
		tokensBeforeFee = new(big.Float).Quo(numerator, denominator)
	}

	// Apply fee to output (user receives fewer tokens)
	tokensOut := bc.config.ApplyFee(tokensBeforeFee)

	// Calculate new pool state
	newCNPYReserve := new(big.Float).Add(x, cnpyAmountIn)             // Add full CNPY to reserve
	newTokenReserve := new(big.Float).Sub(y, tokensOut)               // Remove tokens from virtual pool (sold to user)
	newTotalSupply := new(big.Float).Add(pool.TotalSupply, tokensOut) // Mint tokens to user

	// Calculate effective price for this trade
	effectivePrice := new(big.Float).Quo(cnpyAmountIn, tokensOut)

	// Calculate price impact
	priceImpact := bc.calculatePriceImpact(priceBefore, effectivePrice)

	return &TradeResult{
		AmountOut:       tokensOut,
		NewCNPYReserve:  newCNPYReserve,
		NewTokenReserve: newTokenReserve,
		NewTotalSupply:  newTotalSupply,
		Price:           effectivePrice,
		PriceImpact:     priceImpact,
	}, nil
}

// Sell executes a sell transaction (burning tokens)
// Pump.fun sum-style bonding curve formula: dX = (tokenAmountIn * x) / (y + tokenAmountIn)
// Where x = CNPY reserve, y = token reserve, dX = CNPY to receive
func (bc *BondingCurve) Sell(pool *VirtualPool, tokenAmountIn *big.Float) (*TradeResult, error) {
	if err := pool.Validate(); err != nil {
		return nil, err
	}

	if tokenAmountIn == nil || tokenAmountIn.Sign() <= 0 {
		return nil, ErrZeroAmount
	}

	// Check if there are enough tokens to burn
	if tokenAmountIn.Cmp(pool.TotalSupply) > 0 {
		return nil, ErrInsufficientTokens
	}

	// Get current pool state
	x := pool.CNPYReserve  // CNPY reserve
	y := pool.TokenReserve // Token reserve

	// Calculate price before trade for price impact
	priceBefore := pool.CurrentPrice()

	// Calculate CNPY to receive: dX = (tokenAmountIn * x) / (y + tokenAmountIn)

	// Numerator: tokenAmountIn * x
	numerator := new(big.Float).Mul(tokenAmountIn, x)

	// Denominator: y + tokenAmountIn
	denominator := new(big.Float).Add(y, tokenAmountIn)

	// Prevent division by zero
	if denominator.Sign() == 0 {
		return nil, ErrInsufficientReserve
	}

	// Calculate CNPY to receive (before fees)
	cnpyOut := new(big.Float).Quo(numerator, denominator)

	// Apply fees to the output
	cnpyOutAfterFee := bc.config.ApplyFee(cnpyOut)

	// Validate we have enough CNPY in reserve
	if cnpyOutAfterFee.Cmp(x) > 0 {
		return nil, ErrInsufficientReserve
	}

	// Calculate new pool state
	newCNPYReserve := new(big.Float).Sub(x, cnpyOut)                      // Remove CNPY (before fees) from pool
	newTokenReserve := new(big.Float).Add(y, tokenAmountIn)               // Add tokens back to virtual pool
	newTotalSupply := new(big.Float).Sub(pool.TotalSupply, tokenAmountIn) // Burn tokens

	// Validate the new state
	if newCNPYReserve.Sign() < 0 || newTotalSupply.Sign() < 0 {
		return nil, ErrInsufficientReserve
	}

	// Calculate effective price for this trade
	effectivePrice := new(big.Float).Quo(cnpyOutAfterFee, tokenAmountIn)

	// Calculate price impact
	priceImpact := bc.calculatePriceImpact(priceBefore, effectivePrice)

	return &TradeResult{
		AmountOut:       cnpyOutAfterFee,
		NewCNPYReserve:  newCNPYReserve,
		NewTokenReserve: newTokenReserve,
		NewTotalSupply:  newTotalSupply,
		Price:           effectivePrice,
		PriceImpact:     priceImpact,
	}, nil
}

// SimulateBuy simulates a buy without modifying the pool state
func (bc *BondingCurve) SimulateBuy(pool *VirtualPool, cnpyAmountIn *big.Float) (*TradeResult, error) {
	poolCopy := pool.Copy()
	return bc.Buy(poolCopy, cnpyAmountIn)
}

// SimulateSell simulates a sell without modifying the pool state
func (bc *BondingCurve) SimulateSell(pool *VirtualPool, tokenAmountIn *big.Float) (*TradeResult, error) {
	poolCopy := pool.Copy()
	return bc.Sell(poolCopy, tokenAmountIn)
}

// GetAmountOut calculates the output amount for a given input (simulation only)
func (bc *BondingCurve) GetAmountOut(pool *VirtualPool, amountIn *big.Float, isBuy bool) (*big.Float, error) {
	if isBuy {
		result, err := bc.SimulateBuy(pool, amountIn)
		if err != nil {
			return nil, err
		}
		return result.AmountOut, nil
	} else {
		result, err := bc.SimulateSell(pool, amountIn)
		if err != nil {
			return nil, err
		}
		return result.AmountOut, nil
	}
}

// CalculateSlippage calculates the slippage for a trade
func (bc *BondingCurve) CalculateSlippage(pool *VirtualPool, amountIn *big.Float, expectedOut *big.Float, isBuy bool) (*big.Float, error) {
	actualOut, err := bc.GetAmountOut(pool, amountIn, isBuy)
	if err != nil {
		return nil, err
	}

	if expectedOut.Sign() == 0 {
		return big.NewFloat(0), nil
	}

	// Slippage = (expectedOut - actualOut) / expectedOut * 100
	diff := new(big.Float).Sub(expectedOut, actualOut)
	slippage := new(big.Float).Quo(diff, expectedOut)
	slippage.Mul(slippage, big.NewFloat(100))

	return slippage, nil
}

// calculatePriceImpact calculates the price impact percentage
func (bc *BondingCurve) calculatePriceImpact(priceBefore, priceAfter *big.Float) *big.Float {
	if priceBefore.Sign() == 0 {
		return big.NewFloat(0)
	}

	// Price impact = |(priceAfter - priceBefore) / priceBefore| * 100
	diff := new(big.Float).Sub(priceAfter, priceBefore)
	impact := new(big.Float).Quo(diff, priceBefore)
	impact.Abs(impact)
	impact.Mul(impact, big.NewFloat(100))

	return impact
}

// EstimatePriceAfterTrade estimates the new pool price after a trade
func (bc *BondingCurve) EstimatePriceAfterTrade(pool *VirtualPool, amountIn *big.Float, isBuy bool) (*big.Float, error) {
	var result *TradeResult
	var err error

	if isBuy {
		result, err = bc.SimulateBuy(pool, amountIn)
	} else {
		result, err = bc.SimulateSell(pool, amountIn)
	}

	if err != nil {
		return nil, err
	}

	// Calculate new price after trade
	newPool := NewVirtualPool(result.NewCNPYReserve, result.NewTokenReserve, result.NewTotalSupply)
	return newPool.CurrentPrice(), nil
}

// GetConfig returns the bonding curve configuration
func (bc *BondingCurve) GetConfig() *BondingCurveConfig {
	return bc.config
}

// GetOptimalTradeSize finds the optimal trade size for a given maximum slippage
func (bc *BondingCurve) GetOptimalTradeSize(pool *VirtualPool, maxSlippage *big.Float, isBuy bool) (*big.Float, error) {
	// This is a simplified implementation
	// In practice, you might want to use binary search or more sophisticated optimization

	maxAmount := pool.CNPYReserve
	if !isBuy {
		maxAmount = pool.TotalSupply
	}

	// Start with 1% of max amount and iterate
	testAmount := new(big.Float).Mul(maxAmount, big.NewFloat(0.01))
	step := new(big.Float).Mul(maxAmount, big.NewFloat(0.001))

	for testAmount.Cmp(maxAmount) < 0 {
		var result *TradeResult
		var err error

		if isBuy {
			result, err = bc.SimulateBuy(pool, testAmount)
		} else {
			result, err = bc.SimulateSell(pool, testAmount)
		}

		if err != nil {
			break
		}

		if result.PriceImpact.Cmp(maxSlippage) > 0 {
			// Return the previous amount
			return new(big.Float).Sub(testAmount, step), nil
		}

		testAmount.Add(testAmount, step)
	}

	return testAmount, nil
}

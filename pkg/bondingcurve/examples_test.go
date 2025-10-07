package bondingcurve_test

import (
	"fmt"
	"math/big"

	"github.com/enielson/launchpad/pkg/bondingcurve"
)

// Example demonstrates basic bonding curve usage
func Example() {
	// Create a new bonding curve with default configuration (1% fee)
	bc := bondingcurve.NewBondingCurve(nil)

	// Initialize a virtual pool with starting reserves
	// 1000 CNPY reserve, 800,000 token reserve, 200,000 tokens already minted
	pool := bondingcurve.NewVirtualPool(
		big.NewFloat(1000),   // CNPY reserve
		big.NewFloat(800000), // Token reserve
		big.NewFloat(200000), // Total supply (tokens already in circulation)
	)

	fmt.Printf("Initial pool state:\n")
	fmt.Printf("CNPY Reserve: %s\n", pool.CNPYReserve.String())
	fmt.Printf("Token Reserve: %s\n", pool.TokenReserve.String())
	fmt.Printf("Total Supply: %s\n", pool.TotalSupply.String())
	fmt.Printf("Current Price: %s CNPY per token\n", pool.CurrentPrice().String())

	// Simulate a buy transaction: user wants to spend 100 CNPY
	buyAmount := big.NewFloat(100)
	buyResult, err := bc.SimulateBuy(pool, buyAmount)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nBuy simulation (100 CNPY):\n")
	fmt.Printf("Tokens received: %s\n", buyResult.AmountOut.String())
	fmt.Printf("Effective price: %s CNPY per token\n", buyResult.Price.String())
	fmt.Printf("Price impact: %s%%\n", buyResult.PriceImpact.String())

	// Execute the buy transaction
	actualBuyResult, err := bc.Buy(pool, buyAmount)
	if err != nil {
		panic(err)
	}

	// Update pool state
	pool.CNPYReserve = actualBuyResult.NewCNPYReserve
	pool.TokenReserve = actualBuyResult.NewTokenReserve
	pool.TotalSupply = actualBuyResult.NewTotalSupply

	fmt.Printf("\nAfter buy transaction:\n")
	fmt.Printf("New CNPY Reserve: %s\n", pool.CNPYReserve.String())
	fmt.Printf("New Token Reserve: %s\n", pool.TokenReserve.String())
	fmt.Printf("New Total Supply: %s\n", pool.TotalSupply.String())
	fmt.Printf("New Price: %s CNPY per token\n", pool.CurrentPrice().String())

	// Output:
	// Initial pool state:
	// CNPY Reserve: 1000
	// Token Reserve: 800000
	// Total Supply: 200000
	// Current Price: 0.00125 CNPY per token
	//
	// Buy simulation (100 CNPY):
	// Tokens received: 72000
	// Effective price: 0.001388888889 CNPY per token
	// Price impact: 11.11111111%
	//
	// After buy transaction:
	// New CNPY Reserve: 1100
	// New Token Reserve: 728000
	// New Total Supply: 272000
	// New Price: 0.001510989011 CNPY per token
}

// ExampleBondingCurve_Buy demonstrates the buy mechanism with fee calculation
func ExampleBondingCurve_Buy() {
	// Create bonding curve with custom fee (2%)
	config := &bondingcurve.BondingCurveConfig{
		FeeRateBasisPoints: 200, // 2%
	}
	bc := bondingcurve.NewBondingCurve(config)

	// Create pool
	pool := bondingcurve.NewVirtualPool(
		big.NewFloat(500),    // 500 CNPY reserve
		big.NewFloat(400000), // 400,000 token reserve
		big.NewFloat(100000), // 100,000 tokens already minted
	)

	// User buys with 50 CNPY
	cnpyAmount := big.NewFloat(50)
	result, err := bc.Buy(pool, cnpyAmount)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Input: %s CNPY\n", cnpyAmount.String())
	fmt.Printf("Fee: %s CNPY\n", config.CalculateFee(cnpyAmount).String())
	fmt.Printf("Amount after fee: %s CNPY\n", config.ApplyFee(cnpyAmount).String())
	fmt.Printf("Tokens received: %s\n", result.AmountOut.String())
	fmt.Printf("Effective price: %s CNPY per token\n", result.Price.String())

	// Output:
	// Input: 50 CNPY
	// Fee: 1 CNPY
	// Amount after fee: 49 CNPY
	// Tokens received: 35636.36364
	// Effective price: 0.001403061224 CNPY per token
}

// ExampleBondingCurve_Sell demonstrates the sell mechanism with token burning
func ExampleBondingCurve_Sell() {
	bc := bondingcurve.NewBondingCurve(bondingcurve.NewBondingCurveConfig())

	// Pool after some activity
	pool := bondingcurve.NewVirtualPool(
		big.NewFloat(1200),   // 1200 CNPY reserve
		big.NewFloat(750000), // 750,000 token reserve
		big.NewFloat(250000), // 250,000 tokens in circulation
	)

	// User sells 5000 tokens
	tokenAmount := big.NewFloat(5000)
	result, err := bc.Sell(pool, tokenAmount)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tokens sold: %s\n", tokenAmount.String())
	fmt.Printf("CNPY received (after fee): %s\n", result.AmountOut.String())
	fmt.Printf("Effective price: %s CNPY per token\n", result.Price.String())
	fmt.Printf("Tokens burned (removed from supply): %s\n", tokenAmount.String())
	fmt.Printf("New total supply: %s\n", result.NewTotalSupply.String())

	// Output:
	// Tokens sold: 5000
	// CNPY received (after fee): 7.867549669
	// Effective price: 0.001573509934 CNPY per token
	// Tokens burned (removed from supply): 5000
	// New total supply: 245000
}

// ExampleVirtualPool_CurrentPrice demonstrates price calculation
func ExampleVirtualPool_CurrentPrice() {
	// Create pools with different reserve ratios
	pools := []*bondingcurve.VirtualPool{
		bondingcurve.NewVirtualPool(big.NewFloat(1000), big.NewFloat(1000000), big.NewFloat(0)),
		bondingcurve.NewVirtualPool(big.NewFloat(2000), big.NewFloat(1000000), big.NewFloat(0)),
		bondingcurve.NewVirtualPool(big.NewFloat(1000), big.NewFloat(500000), big.NewFloat(0)),
	}

	fmt.Printf("Pool prices (CNPY per token):\n")
	for i, pool := range pools {
		price := pool.CurrentPrice()
		fmt.Printf("Pool %d: %s CNPY/token (CNPY: %s, Tokens: %s)\n",
			i+1, price.String(), pool.CNPYReserve.String(), pool.TokenReserve.String())
	}

	// Output:
	// Pool prices (CNPY per token):
	// Pool 1: 0.001 CNPY/token (CNPY: 1000, Tokens: 1000000)
	// Pool 2: 0.002 CNPY/token (CNPY: 2000, Tokens: 1000000)
	// Pool 3: 0.002 CNPY/token (CNPY: 1000, Tokens: 500000)
}

// ExampleBondingCurve_SimulateBuy demonstrates simulation without state changes
func ExampleBondingCurve_SimulateBuy() {
	bc := bondingcurve.NewBondingCurve(bondingcurve.NewBondingCurveConfig())

	pool := bondingcurve.NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	// Store original state
	originalCNPY := new(big.Float).Copy(pool.CNPYReserve)
	originalTokens := new(big.Float).Copy(pool.TokenReserve)

	// Simulate multiple buy sizes
	amounts := []*big.Float{
		big.NewFloat(10),
		big.NewFloat(50),
		big.NewFloat(100),
		big.NewFloat(500),
	}

	fmt.Printf("Buy simulations:\n")
	for _, amount := range amounts {
		result, err := bc.SimulateBuy(pool, amount)
		if err != nil {
			continue
		}

		fmt.Printf("Buy %s CNPY → %s tokens (price impact: %s%%)\n",
			amount.String(),
			result.AmountOut.String(),
			result.PriceImpact.String())
	}

	// Verify pool state unchanged
	fmt.Printf("\nPool state unchanged after simulations:\n")
	fmt.Printf("CNPY Reserve: %s (was %s)\n", pool.CNPYReserve.String(), originalCNPY.String())
	fmt.Printf("Token Reserve: %s (was %s)\n", pool.TokenReserve.String(), originalTokens.String())

	// Output:
	// Buy simulations:
	// Buy 10 CNPY → 7841.584158 tokens (price impact: 2.02020202%)
	// Buy 50 CNPY → 37714.28571 tokens (price impact: 6.060606061%)
	// Buy 100 CNPY → 72000 tokens (price impact: 11.11111111%)
	// Buy 500 CNPY → 264000 tokens (price impact: 51.51515152%)
	//
	// Pool state unchanged after simulations:
	// CNPY Reserve: 1000 (was 1000)
	// Token Reserve: 800000 (was 800000)
}

// ExampleBondingCurve_GetOptimalTradeSize demonstrates finding optimal trade sizes
func ExampleBondingCurve_GetOptimalTradeSize() {
	bc := bondingcurve.NewBondingCurve(bondingcurve.NewBondingCurveConfig())

	pool := bondingcurve.NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	// Find optimal trade sizes for different slippage tolerances
	slippages := []*big.Float{
		big.NewFloat(1),  // 1%
		big.NewFloat(5),  // 5%
		big.NewFloat(10), // 10%
		big.NewFloat(20), // 20%
	}

	fmt.Printf("Optimal buy sizes for different slippage tolerances:\n")
	for _, maxSlippage := range slippages {
		optimalSize, err := bc.GetOptimalTradeSize(pool, maxSlippage, true)
		if err != nil {
			continue
		}

		// Verify the slippage
		result, _ := bc.SimulateBuy(pool, optimalSize)
		fmt.Printf("Max %s%% → %s CNPY (actual impact: %s%%)\n",
			maxSlippage.String(),
			optimalSize.String(),
			result.PriceImpact.String())
	}

	// Output:
	// Optimal buy sizes for different slippage tolerances:
	// Max 1% → 9 CNPY (actual impact: 1.919191919%)
	// Max 5% → 39 CNPY (actual impact: 4.949494949%)
	// Max 10% → 89 CNPY (actual impact: 10%)
	// Max 20% → 187 CNPY (actual impact: 19.8989899%)
}

// ExampleBondingCurveConfig_ApplyFee demonstrates fee calculations
func ExampleBondingCurveConfig_ApplyFee() {
	// Different fee configurations
	configs := []*bondingcurve.BondingCurveConfig{
		{FeeRateBasisPoints: 0},   // 0% fee
		{FeeRateBasisPoints: 50},  // 0.5% fee
		{FeeRateBasisPoints: 100}, // 1% fee
		{FeeRateBasisPoints: 300}, // 3% fee
	}

	amount := big.NewFloat(1000)

	fmt.Printf("Fee calculations for %s CNPY:\n", amount.String())
	for _, config := range configs {
		fee := config.CalculateFee(amount)
		afterFee := config.ApplyFee(amount)

		feePercent := new(big.Float).SetUint64(config.FeeRateBasisPoints)
		feePercent.Quo(feePercent, big.NewFloat(100))

		fmt.Printf("Fee %s%% → Fee: %s, After fee: %s\n",
			feePercent.String(), fee.String(), afterFee.String())
	}

	// Output:
	// Fee calculations for 1000 CNPY:
	// Fee 0% → Fee: 0, After fee: 1000
	// Fee 0.5% → Fee: 5, After fee: 995
	// Fee 1% → Fee: 10, After fee: 990
	// Fee 3% → Fee: 30, After fee: 970
}

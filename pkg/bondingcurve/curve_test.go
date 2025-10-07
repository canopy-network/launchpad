package bondingcurve

import (
	"math/big"
	"testing"
)

func TestNewBondingCurve(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		config := &BondingCurveConfig{FeeRateBasisPoints: 200}
		bc := NewBondingCurve(config)

		if bc.config.FeeRateBasisPoints != 200 {
			t.Errorf("expected fee rate 200, got %d", bc.config.FeeRateBasisPoints)
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		bc := NewBondingCurve(nil)

		if bc.config.FeeRateBasisPoints != DefaultFeeRateBasisPoints {
			t.Errorf("expected default fee rate %d, got %d", DefaultFeeRateBasisPoints, bc.config.FeeRateBasisPoints)
		}
	})
}

func TestBondingCurve_Buy(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	// Test pool: 1000 CNPY, 800,000 tokens
	pool := NewVirtualPool(
		big.NewFloat(1000),   // CNPY reserve
		big.NewFloat(800000), // Token reserve
		big.NewFloat(200000), // Total supply (tokens already minted)
	)

	t.Run("valid buy", func(t *testing.T) {
		cnpyIn := big.NewFloat(100) // Buy with 100 CNPY

		result, err := bc.Buy(pool, cnpyIn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify we got tokens out
		if result.AmountOut.Sign() <= 0 {
			t.Errorf("expected tokens out > 0, got %s", result.AmountOut.String())
		}

		// Verify CNPY reserve increased
		if result.NewCNPYReserve.Cmp(pool.CNPYReserve) <= 0 {
			t.Errorf("expected CNPY reserve to increase")
		}

		// Verify token reserve decreased (tokens removed from virtual pool)
		if result.NewTokenReserve.Cmp(pool.TokenReserve) >= 0 {
			t.Errorf("expected token reserve to decrease")
		}

		// Verify total supply increased (tokens minted)
		if result.NewTotalSupply.Cmp(pool.TotalSupply) <= 0 {
			t.Errorf("expected total supply to increase")
		}

		// Verify price is positive
		if result.Price.Sign() <= 0 {
			t.Errorf("expected positive price, got %s", result.Price.String())
		}

		// Verify the bonding curve formula
		// Pump.fun: dY = (cnpyIn * y) / (x + cnpyIn), then apply fee to output
		expectedNumerator := new(big.Float).Mul(cnpyIn, pool.TokenReserve)
		expectedDenominator := new(big.Float).Add(pool.CNPYReserve, cnpyIn)
		tokensBeforeFee := new(big.Float).Quo(expectedNumerator, expectedDenominator)
		expectedTokens := bc.config.ApplyFee(tokensBeforeFee)

		if result.AmountOut.Cmp(expectedTokens) != 0 {
			t.Errorf("formula calculation mismatch: expected %s, got %s",
				expectedTokens.String(), result.AmountOut.String())
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		_, err := bc.Buy(pool, big.NewFloat(0))
		if err != ErrZeroAmount {
			t.Errorf("expected ErrZeroAmount, got %v", err)
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		_, err := bc.Buy(pool, big.NewFloat(-100))
		if err != ErrZeroAmount {
			t.Errorf("expected ErrZeroAmount, got %v", err)
		}
	})

	t.Run("nil amount", func(t *testing.T) {
		_, err := bc.Buy(pool, nil)
		if err != ErrZeroAmount {
			t.Errorf("expected ErrZeroAmount, got %v", err)
		}
	})

	t.Run("invalid pool", func(t *testing.T) {
		invalidPool := &VirtualPool{
			CNPYReserve:  big.NewFloat(0),
			TokenReserve: big.NewFloat(1000),
			TotalSupply:  big.NewFloat(100),
		}

		_, err := bc.Buy(invalidPool, big.NewFloat(100))
		if err != ErrInsufficientReserve {
			t.Errorf("expected ErrInsufficientReserve, got %v", err)
		}
	})
}

func TestBondingCurve_Sell(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	// Test pool: 1000 CNPY, 800,000 tokens, 200,000 total supply
	pool := NewVirtualPool(
		big.NewFloat(1000),   // CNPY reserve
		big.NewFloat(800000), // Token reserve
		big.NewFloat(200000), // Total supply
	)

	t.Run("valid sell", func(t *testing.T) {
		tokensIn := big.NewFloat(1000) // Sell 1000 tokens

		result, err := bc.Sell(pool, tokensIn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify we got CNPY out
		if result.AmountOut.Sign() <= 0 {
			t.Errorf("expected CNPY out > 0, got %s", result.AmountOut.String())
		}

		// Verify CNPY reserve decreased
		if result.NewCNPYReserve.Cmp(pool.CNPYReserve) >= 0 {
			t.Errorf("expected CNPY reserve to decrease")
		}

		// Verify token reserve increased (tokens added back to virtual pool)
		if result.NewTokenReserve.Cmp(pool.TokenReserve) <= 0 {
			t.Errorf("expected token reserve to increase")
		}

		// Verify total supply decreased (tokens burned)
		if result.NewTotalSupply.Cmp(pool.TotalSupply) >= 0 {
			t.Errorf("expected total supply to decrease")
		}

		// Verify price is positive
		if result.Price.Sign() <= 0 {
			t.Errorf("expected positive price, got %s", result.Price.String())
		}
	})

	t.Run("sell more than total supply", func(t *testing.T) {
		tokensIn := big.NewFloat(300000) // More than total supply

		_, err := bc.Sell(pool, tokensIn)
		if err != ErrInsufficientTokens {
			t.Errorf("expected ErrInsufficientTokens, got %v", err)
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		_, err := bc.Sell(pool, big.NewFloat(0))
		if err != ErrZeroAmount {
			t.Errorf("expected ErrZeroAmount, got %v", err)
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		_, err := bc.Sell(pool, big.NewFloat(-100))
		if err != ErrZeroAmount {
			t.Errorf("expected ErrZeroAmount, got %v", err)
		}
	})
}

func TestBondingCurve_SimulateBuy(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	originalCNPY := new(big.Float).Copy(pool.CNPYReserve)
	originalTokens := new(big.Float).Copy(pool.TokenReserve)
	originalSupply := new(big.Float).Copy(pool.TotalSupply)

	t.Run("simulation doesn't modify original pool", func(t *testing.T) {
		_, err := bc.SimulateBuy(pool, big.NewFloat(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify original pool is unchanged
		if pool.CNPYReserve.Cmp(originalCNPY) != 0 {
			t.Errorf("original CNPY reserve was modified")
		}
		if pool.TokenReserve.Cmp(originalTokens) != 0 {
			t.Errorf("original token reserve was modified")
		}
		if pool.TotalSupply.Cmp(originalSupply) != 0 {
			t.Errorf("original total supply was modified")
		}
	})
}

func TestBondingCurve_SimulateSell(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	originalCNPY := new(big.Float).Copy(pool.CNPYReserve)
	originalTokens := new(big.Float).Copy(pool.TokenReserve)
	originalSupply := new(big.Float).Copy(pool.TotalSupply)

	t.Run("simulation doesn't modify original pool", func(t *testing.T) {
		_, err := bc.SimulateSell(pool, big.NewFloat(1000))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify original pool is unchanged
		if pool.CNPYReserve.Cmp(originalCNPY) != 0 {
			t.Errorf("original CNPY reserve was modified")
		}
		if pool.TokenReserve.Cmp(originalTokens) != 0 {
			t.Errorf("original token reserve was modified")
		}
		if pool.TotalSupply.Cmp(originalSupply) != 0 {
			t.Errorf("original total supply was modified")
		}
	})
}

func TestBondingCurve_GetAmountOut(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	t.Run("buy amount out", func(t *testing.T) {
		amountOut, err := bc.GetAmountOut(pool, big.NewFloat(100), true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if amountOut.Sign() <= 0 {
			t.Errorf("expected positive amount out, got %s", amountOut.String())
		}
	})

	t.Run("sell amount out", func(t *testing.T) {
		amountOut, err := bc.GetAmountOut(pool, big.NewFloat(1000), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if amountOut.Sign() <= 0 {
			t.Errorf("expected positive amount out, got %s", amountOut.String())
		}
	})
}

func TestBondingCurve_CalculateSlippage(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	t.Run("calculate slippage", func(t *testing.T) {
		// With 100 CNPY, we get 72000 tokens (after 1% fee)
		// If we expect 80000 tokens, there's slippage
		expectedOut := big.NewFloat(80000)
		slippage, err := bc.CalculateSlippage(pool, big.NewFloat(100), expectedOut, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Slippage should be positive (we get less than expected)
		if slippage.Sign() < 0 {
			t.Errorf("expected non-negative slippage, got %s", slippage.String())
		}
	})

	t.Run("zero expected out", func(t *testing.T) {
		slippage, err := bc.CalculateSlippage(pool, big.NewFloat(100), big.NewFloat(0), true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return 0 when expected out is 0
		if slippage.Sign() != 0 {
			t.Errorf("expected zero slippage, got %s", slippage.String())
		}
	})
}

func TestBondingCurve_EstimatePriceAfterTrade(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	t.Run("price after buy", func(t *testing.T) {
		originalPrice := pool.CurrentPrice()
		newPrice, err := bc.EstimatePriceAfterTrade(pool, big.NewFloat(100), true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Price should increase after buy
		if newPrice.Cmp(originalPrice) <= 0 {
			t.Errorf("expected price to increase after buy")
		}
	})

	t.Run("price after sell", func(t *testing.T) {
		originalPrice := pool.CurrentPrice()
		newPrice, err := bc.EstimatePriceAfterTrade(pool, big.NewFloat(1000), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Price should decrease after sell
		if newPrice.Cmp(originalPrice) >= 0 {
			t.Errorf("expected price to decrease after sell")
		}
	})
}

func TestBondingCurve_GetOptimalTradeSize(t *testing.T) {
	bc := NewBondingCurve(NewBondingCurveConfig())

	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)

	t.Run("optimal buy size", func(t *testing.T) {
		maxSlippage := big.NewFloat(5) // 5% max slippage
		optimalSize, err := bc.GetOptimalTradeSize(pool, maxSlippage, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if optimalSize.Sign() <= 0 {
			t.Errorf("expected positive optimal size, got %s", optimalSize.String())
		}

		// Verify the optimal size doesn't exceed max slippage
		result, err := bc.SimulateBuy(pool, optimalSize)
		if err != nil {
			t.Fatalf("unexpected error simulating optimal buy: %v", err)
		}

		if result.PriceImpact.Cmp(maxSlippage) > 0 {
			t.Errorf("optimal size exceeds max slippage: %s > %s",
				result.PriceImpact.String(), maxSlippage.String())
		}
	})

	t.Run("optimal sell size", func(t *testing.T) {
		maxSlippage := big.NewFloat(5) // 5% max slippage
		optimalSize, err := bc.GetOptimalTradeSize(pool, maxSlippage, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if optimalSize.Sign() <= 0 {
			t.Errorf("expected positive optimal size, got %s", optimalSize.String())
		}
	})
}

// Benchmark tests
func BenchmarkBondingCurve_Buy(b *testing.B) {
	bc := NewBondingCurve(NewBondingCurveConfig())
	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)
	cnpyIn := big.NewFloat(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		poolCopy := pool.Copy()
		_, err := bc.Buy(poolCopy, cnpyIn)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkBondingCurve_Sell(b *testing.B) {
	bc := NewBondingCurve(NewBondingCurveConfig())
	pool := NewVirtualPool(
		big.NewFloat(1000),
		big.NewFloat(800000),
		big.NewFloat(200000),
	)
	tokensIn := big.NewFloat(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		poolCopy := pool.Copy()
		_, err := bc.Sell(poolCopy, tokensIn)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestBondingCurve_SequentialBuys(t *testing.T) {
	// Use initial price of 0.05 to achieve realistic token amounts
	config := &BondingCurveConfig{
		FeeRateBasisPoints: 100,
		InitialPrice:       big.NewFloat(0.05),
	}
	bc := NewBondingCurve(config)

	// Initialize pool with initial primary reserve
	// Starting with 100 CNPY means the first buy uses the bonding curve formula
	// instead of the fixed initial price
	initialCNPY := 100.0
	initialTokens := initialCNPY / 0.05 // tokens = CNPY / initial_price

	pool := NewVirtualPool(
		big.NewFloat(initialCNPY),   // Initial CNPY reserve: 100
		big.NewFloat(initialTokens), // Initial token reserve: 2000
		big.NewFloat(initialTokens), // Total supply: 2000 (tokens already minted)
	)

	t.Logf("\n=== Initial Pool State ===")
	t.Logf("CNPY Reserve:  %s", pool.CNPYReserve.Text('f', 2))
	t.Logf("Token Reserve: %s", pool.TokenReserve.Text('f', 2))
	t.Logf("Total Supply:  %s", pool.TotalSupply.Text('f', 2))
	t.Logf("Initial Price: %s CNPY per token", pool.CurrentPrice().Text('f', 8))
	t.Logf("Fee Rate: %d bps (%.2f%%)\n", bc.config.FeeRateBasisPoints, float64(bc.config.FeeRateBasisPoints)/100)

	// Execute sequential buys totaling 50,000 CNPY
	buyAmounts := []float64{
		100, 200, 300, 400, 500, 600, 700, 800, 900, 1000,
		1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800, 1900, 2000,
		2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000,
	}

	// Print table header
	t.Logf("%-4s | %-10s | %-12s | %-10s | %-12s | %-12s | %-12s | %-12s",
		"#", "CNPY In", "Tokens Out", "Fee", "Price", "Impact", "CNPY Rsv", "Total Supply")
	t.Logf("-----|------------|--------------|------------|--------------|--------------|--------------|--------------|")

	for i, amount := range buyAmounts {
		cnpyIn := big.NewFloat(amount)

		result, err := bc.Buy(pool, cnpyIn)
		if err != nil {
			t.Fatalf("Buy #%d failed: %v", i+1, err)
		}

		// Update pool state
		pool.CNPYReserve = result.NewCNPYReserve
		pool.TokenReserve = result.NewTokenReserve
		pool.TotalSupply = result.NewTotalSupply

		// Calculate fee amount
		tokensBeforeFee := new(big.Float).Add(result.AmountOut, new(big.Float).Mul(result.AmountOut, big.NewFloat(float64(bc.config.FeeRateBasisPoints)/float64(BasisPointsDivisor-bc.config.FeeRateBasisPoints))))
		feeAmount := bc.config.CalculateFee(tokensBeforeFee)

		// Format values for display
		tokensOut, _ := result.AmountOut.Float64()
		fee, _ := feeAmount.Float64()
		price, _ := result.Price.Float64()
		impact, _ := result.PriceImpact.Float64()
		cnpyRsv, _ := result.NewCNPYReserve.Float64()
		totalSup, _ := result.NewTotalSupply.Float64()

		t.Logf("%-4d | %10.2f | %12.2f | %10.2f | %12.8f | %11.2f%% | %12.2f | %12.2f",
			i+1, amount, tokensOut, fee, price, impact, cnpyRsv, totalSup)
	}

	t.Logf("\n=== Final Pool State ===")
	t.Logf("CNPY Reserve:  %s", pool.CNPYReserve.Text('f', 2))
	t.Logf("Token Reserve: %s", pool.TokenReserve.Text('f', 2))
	t.Logf("Total Supply:  %s", pool.TotalSupply.Text('f', 2))
	t.Logf("Final Price:   %s CNPY per token", pool.CurrentPrice().Text('f', 8))
}

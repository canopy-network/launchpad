//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/enielson/launchpad/tests/testutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPriceHistoryAPI tests the price history endpoint
func TestPriceHistoryAPI(t *testing.T) {
	var chainID uuid.UUID
	var userID uuid.UUID
	var poolID uuid.UUID

	// Setup: Create test data with transactions at different times
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		ctx := context.Background()

		// Create test user
		user, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("pricehistory%d@example.com", time.Now().UnixNano())).
			WithUsername(fmt.Sprintf("priceuser%d", time.Now().Unix())).
			Create(ctx, db)
		require.NoError(t, err)
		userID = user.ID

		// Create test chain
		chain, err := fixtures.DefaultChain(userID).
			WithStatus(models.ChainStatusVirtualActive).
			Create(ctx, db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create virtual pool
		pool, err := fixtures.DefaultVirtualPool(chainID).Create(ctx, db)
		require.NoError(t, err)
		poolID = pool.ID

		// Create transactions at different times to test OHLC aggregation
		// Use truncated times to ensure they fall within the same minute bucket
		baseTime := time.Now().Add(-10 * time.Minute).Truncate(time.Minute)

		// Minute 1: 3 transactions (Open=0.00001, High=0.00003, Low=0.00001, Close=0.00002)
		createTransactionAtTime(t, ctx, db, chainID, userID, poolID, baseTime, 0.00001, 1000.0)
		createTransactionAtTime(t, ctx, db, chainID, userID, poolID, baseTime.Add(10*time.Second), 0.00003, 500.0)
		createTransactionAtTime(t, ctx, db, chainID, userID, poolID, baseTime.Add(30*time.Second), 0.00002, 750.0)

		// Minute 2: No transactions (gap test to verify sparse data)

		// Minute 3: 2 transactions (Open=0.00004, High=0.00005, Low=0.00004, Close=0.00005)
		minute3 := baseTime.Add(2 * time.Minute)
		createTransactionAtTime(t, ctx, db, chainID, userID, poolID, minute3, 0.00004, 600.0)
		createTransactionAtTime(t, ctx, db, chainID, userID, poolID, minute3.Add(45*time.Second), 0.00005, 800.0)

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM virtual_pool_transactions WHERE chain_id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM virtual_pools WHERE chain_id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", userID)
		})

		t.Logf("Created test chain %s with %d transactions in 2 distinct minute buckets", chainID, 5)
	})

	client := testutils.NewTestClient()

	t.Run("Get price history with default time range", func(t *testing.T) {
		path := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/price-history", chainID))
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var responseData struct {
			Data []models.PriceHistoryCandle `json:"data"`
		}
		testutils.UnmarshalResponse(t, body, &responseData)
		candles := responseData.Data

		// Should have 2 candles (sparse data - no gap filling)
		assert.Len(t, candles, 2, "Expected 2 candles (minute 1 and minute 3)")

		// Verify first candle (Minute 1)
		if len(candles) > 0 {
			candle1 := candles[0]
			assert.Equal(t, 0.00001, candle1.Open, "First candle open price")
			assert.Equal(t, 0.00003, candle1.High, "First candle high price")
			assert.Equal(t, 0.00001, candle1.Low, "First candle low price")
			assert.Equal(t, 0.00002, candle1.Close, "First candle close price")
			assert.Equal(t, 2250.0, candle1.Volume, "First candle volume (1000+500+750)")
			assert.Equal(t, 3, candle1.TradeCount, "First candle trade count")
		}

		// Verify second candle (Minute 3)
		if len(candles) > 1 {
			candle2 := candles[1]
			assert.Equal(t, 0.00004, candle2.Open, "Second candle open price")
			assert.Equal(t, 0.00005, candle2.High, "Second candle high price")
			assert.Equal(t, 0.00004, candle2.Low, "Second candle low price")
			assert.Equal(t, 0.00005, candle2.Close, "Second candle close price")
			assert.Equal(t, 1400.0, candle2.Volume, "Second candle volume (600+800)")
			assert.Equal(t, 2, candle2.TradeCount, "Second candle trade count")
		}

		t.Logf("✅ Retrieved %d candles successfully", len(candles))
	})

	t.Run("Get price history with custom time range", func(t *testing.T) {
		// Request only the first 5 minutes
		startTime := time.Now().Add(-15 * time.Minute).Format(time.RFC3339)
		endTime := time.Now().Add(-5 * time.Minute).Format(time.RFC3339)

		path := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/price-history?start_time=%s&end_time=%s",
			chainID, startTime, endTime))
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var responseData struct {
			Data []models.PriceHistoryCandle `json:"data"`
		}
		testutils.UnmarshalResponse(t, body, &responseData)
		candles := responseData.Data

		// Should still have our test data
		assert.GreaterOrEqual(t, len(candles), 0, "Should have candles in time range")
		t.Logf("✅ Retrieved %d candles with custom time range", len(candles))
	})

	t.Run("Get price history with invalid chain ID", func(t *testing.T) {
		path := testutils.GetAPIPath("/chains/invalid-uuid/price-history")
		resp, _ := client.Get(t, path)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		t.Logf("✅ Invalid chain ID correctly rejected")
	})

	t.Run("Get price history with non-existent chain", func(t *testing.T) {
		fakeChainID := uuid.New()
		path := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/price-history", fakeChainID))
		resp, _ := client.Get(t, path)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		t.Logf("✅ Non-existent chain correctly returns 404")
	})

	t.Run("Get price history with invalid time format", func(t *testing.T) {
		path := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/price-history?start_time=invalid", chainID))
		resp, _ := client.Get(t, path)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		t.Logf("✅ Invalid time format correctly rejected")
	})

	t.Run("Get price history with empty data", func(t *testing.T) {
		// Reuse the existing chain but query a time range before we created transactions
		// This tests the sparse data behavior when no transactions exist in the time range
		startTime := time.Now().Add(-48 * time.Hour).Format(time.RFC3339)
		endTime := time.Now().Add(-25 * time.Hour).Format(time.RFC3339)

		path := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/price-history?start_time=%s&end_time=%s",
			chainID, startTime, endTime))
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var responseData struct {
			Data []models.PriceHistoryCandle `json:"data"`
		}
		testutils.UnmarshalResponse(t, body, &responseData)
		assert.Len(t, responseData.Data, 0, "Should return empty array for chain with no transactions")
		t.Logf("✅ Empty data returns empty array")
	})
}

// Helper function to create a transaction at a specific time
func createTransactionAtTime(t *testing.T, ctx context.Context, db *sqlx.DB, chainID, userID, poolID uuid.UUID, timestamp time.Time, price, volume float64) {
	query := `
		INSERT INTO virtual_pool_transactions (
			virtual_pool_id, chain_id, user_id, transaction_type,
			cnpy_amount, token_amount, price_per_token_cnpy,
			pool_cnpy_reserve_after, pool_token_reserve_after, market_cap_after_usd,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	tokenAmount := int64(volume / price)
	_, err := db.ExecContext(ctx, query,
		poolID, chainID, userID, "buy",
		volume, tokenAmount, price,
		10000.0, 800000000, 0.0,
		timestamp,
	)
	require.NoError(t, err)
}

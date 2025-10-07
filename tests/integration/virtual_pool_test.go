//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/enielson/launchpad/tests/testutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestVirtualPoolAPI tests virtual pool API endpoints using fixtures
func TestVirtualPoolAPI(t *testing.T) {
	var chainID uuid.UUID
	var userID uuid.UUID

	// Setup: Create test data using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		ctx := context.Background()

		// Create test user
		user, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("pooltest%d@example.com", time.Now().UnixNano())).
			WithUsername(fmt.Sprintf("pooluser%d", time.Now().Unix())).
			Create(ctx, db)
		require.NoError(t, err)
		userID = user.ID

		// Create test chain
		chain, err := fixtures.DefaultChain(userID).
			WithStatus(models.ChainStatusVirtualActive).
			Create(ctx, db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create virtual pool for the chain
		pool, err := fixtures.DefaultVirtualPool(chainID).Create(ctx, db)
		require.NoError(t, err)

		// Create sample transactions
		_, err = fixtures.DefaultVirtualPoolTransaction(chainID, userID).
			WithVirtualPoolID(pool.ID).
			WithTransactionType("buy").
			WithCNPYAmount(10.0).
			WithTokenAmount(500000).
			Create(ctx, db)
		require.NoError(t, err)

		_, err = fixtures.DefaultVirtualPoolTransaction(chainID, userID).
			WithVirtualPoolID(pool.ID).
			WithTransactionType("sell").
			WithCNPYAmount(5.0).
			WithTokenAmount(250000).
			Create(ctx, db)
		require.NoError(t, err)

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

		t.Logf("Created test chain %s with virtual pool (CNPY: %.2f, Tokens: %d)",
			chainID, pool.CNPYReserve, pool.TokenReserve)
	})

	client := testutils.NewTestClient()

	t.Run("get_virtual_pool", func(t *testing.T) {
		path := testutils.GetAPIPath("/chains/" + chainID.String() + "/virtual-pool")
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var poolResponse struct {
			Data map[string]interface{} `json:"data"`
		}
		testutils.UnmarshalResponse(t, body, &poolResponse)
		pool := poolResponse.Data

		// Verify expected fields exist
		requiredFields := []string{"cnpy_reserve", "token_reserve", "current_price_cnpy", "total_transactions"}
		for _, field := range requiredFields {
			if _, exists := pool[field]; !exists {
				t.Errorf("Missing required field: %s", field)
			}
		}

		t.Logf("Virtual pool: CNPY=%.2f, Tokens=%v, Price=%.8f, TxCount=%v",
			pool["cnpy_reserve"], pool["token_reserve"], pool["current_price_cnpy"], pool["total_transactions"])
	})

	t.Run("get_pool_transactions", func(t *testing.T) {
		path := testutils.GetAPIPath("/chains/" + chainID.String() + "/transactions")
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var transactionsResponse struct {
			Data       []map[string]interface{} `json:"data"`
			Pagination *struct {
				Page  int `json:"page"`
				Limit int `json:"limit"`
				Total int `json:"total"`
				Pages int `json:"pages"`
			} `json:"pagination,omitempty"`
		}
		testutils.UnmarshalResponse(t, body, &transactionsResponse)
		transactions := transactionsResponse.Data

		// Should have our 2 test transactions
		if len(transactions) != 2 {
			t.Errorf("Expected 2 transactions, got %d", len(transactions))
		}

		// Verify transaction structure
		if len(transactions) > 0 {
			tx := transactions[0]
			requiredFields := []string{"transaction_type", "cnpy_amount", "token_amount", "price_per_token_cnpy"}
			for _, field := range requiredFields {
				if _, exists := tx[field]; !exists {
					t.Errorf("Transaction missing required field: %s", field)
				}
			}

			t.Logf("Found %d transactions", len(transactions))
			t.Logf("Sample transaction: Type=%s, CNPY=%.2f, Tokens=%v",
				tx["transaction_type"], tx["cnpy_amount"], tx["token_amount"])
		}
	})

	t.Run("get_nonexistent_pool", func(t *testing.T) {
		fakeChainID := "00000000-0000-0000-0000-000000000000"
		path := testutils.GetAPIPath("/chains/" + fakeChainID + "/virtual-pool")
		resp, _ := client.Get(t, path)

		// Should return 404 for nonexistent chain
		if resp.StatusCode != 404 {
			t.Errorf("Expected status 404 for nonexistent chain, got %d", resp.StatusCode)
		}
	})
}

// TestGetVirtualPools tests the GET /api/v1/virtual-pools endpoint
func TestGetVirtualPools(t *testing.T) {
	var chainID1, chainID2 uuid.UUID
	var userID uuid.UUID

	timestamp := time.Now().UnixNano()

	// Setup: Create test data using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		ctx := context.Background()

		// Create test user
		user, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("poolstest%d@example.com", timestamp)).
			WithUsername(fmt.Sprintf("poolsuser%d", timestamp)).
			Create(ctx, db)
		require.NoError(t, err)
		userID = user.ID

		// Create first test chain with virtual pool
		chain1Fixture := fixtures.DefaultChain(userID)
		chain1Fixture.ChainName = fmt.Sprintf("Test Pool Chain 1 %d", timestamp)
		chain1, err := chain1Fixture.
			WithTokenSymbol("TPC1").
			WithStatus(models.ChainStatusVirtualActive).
			Create(ctx, db)
		require.NoError(t, err)
		chainID1 = chain1.ID

		pool1, err := fixtures.DefaultVirtualPool(chainID1).
			WithReserves(1000.0, 1000000).
			Create(ctx, db)
		require.NoError(t, err)

		// Create second test chain with virtual pool
		chain2Fixture := fixtures.DefaultChain(userID)
		chain2Fixture.ChainName = fmt.Sprintf("Test Pool Chain 2 %d", timestamp)
		chain2, err := chain2Fixture.
			WithTokenSymbol("TPC2").
			WithStatus(models.ChainStatusVirtualActive).
			Create(ctx, db)
		require.NoError(t, err)
		chainID2 = chain2.ID

		pool2, err := fixtures.DefaultVirtualPool(chainID2).
			WithReserves(2000.0, 2000000).
			Create(ctx, db)
		require.NoError(t, err)

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM virtual_pools WHERE id IN ($1, $2)", pool1.ID, pool2.ID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id IN ($1, $2)", chainID1, chainID2)
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", userID)
		})

		t.Logf("Created 2 test chains with virtual pools")
	})

	client := testutils.NewTestClient()

	t.Run("get_all_virtual_pools_success", func(t *testing.T) {
		path := testutils.GetAPIPath("/virtual-pools")
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var poolsResponse struct {
			Data       []models.VirtualPool `json:"data"`
			Pagination *models.Pagination   `json:"pagination"`
		}
		testutils.UnmarshalResponse(t, body, &poolsResponse)

		// Should have at least our 2 test pools
		require.GreaterOrEqual(t, len(poolsResponse.Data), 2, "Expected at least 2 virtual pools")

		// Verify pagination exists
		require.NotNil(t, poolsResponse.Pagination)
		require.Greater(t, poolsResponse.Pagination.Total, 0)
		require.Equal(t, 1, poolsResponse.Pagination.Page)
		require.Equal(t, 20, poolsResponse.Pagination.Limit) // Default limit

		// Verify pool structure
		pool := poolsResponse.Data[0]
		require.NotEqual(t, uuid.Nil, pool.ID)
		require.NotEqual(t, uuid.Nil, pool.ChainID)
		require.GreaterOrEqual(t, pool.CNPYReserve, 0.0)
		require.GreaterOrEqual(t, pool.TokenReserve, int64(0))
		require.GreaterOrEqual(t, pool.CurrentPriceCNPY, 0.0)
		require.False(t, pool.CreatedAt.IsZero())
		require.False(t, pool.UpdatedAt.IsZero())

		t.Logf("Retrieved %d virtual pools (total: %d)", len(poolsResponse.Data), poolsResponse.Pagination.Total)
	})

	t.Run("get_virtual_pools_with_pagination", func(t *testing.T) {
		path := testutils.GetAPIPath("/virtual-pools?page=1&limit=1")
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var poolsResponse struct {
			Data       []models.VirtualPool `json:"data"`
			Pagination *models.Pagination   `json:"pagination"`
		}
		testutils.UnmarshalResponse(t, body, &poolsResponse)

		// Should have exactly 1 pool due to limit
		require.Len(t, poolsResponse.Data, 1, "Expected exactly 1 virtual pool with limit=1")

		// Verify pagination parameters
		require.NotNil(t, poolsResponse.Pagination)
		require.Equal(t, 1, poolsResponse.Pagination.Page)
		require.Equal(t, 1, poolsResponse.Pagination.Limit)
		require.GreaterOrEqual(t, poolsResponse.Pagination.Total, 2) // At least our 2 test pools
		require.GreaterOrEqual(t, poolsResponse.Pagination.Pages, 2) // At least 2 pages

		t.Logf("Pagination: Page %d of %d, Total: %d",
			poolsResponse.Pagination.Page, poolsResponse.Pagination.Pages, poolsResponse.Pagination.Total)
	})

	t.Run("get_virtual_pools_page_zero_uses_default", func(t *testing.T) {
		// page=0 defaults to page=1 (not an error)
		path := testutils.GetAPIPath("/virtual-pools?page=0")
		resp, body := client.Get(t, path)

		testutils.AssertStatusOK(t, resp)

		var poolsResponse struct {
			Data       []models.VirtualPool `json:"data"`
			Pagination *models.Pagination   `json:"pagination"`
		}
		testutils.UnmarshalResponse(t, body, &poolsResponse)

		require.NotNil(t, poolsResponse.Pagination)
		require.Equal(t, 1, poolsResponse.Pagination.Page, "page=0 should default to page=1")
		t.Logf("page=0 correctly defaults to page=1")
	})

	t.Run("get_virtual_pools_invalid_limit", func(t *testing.T) {
		path := testutils.GetAPIPath("/virtual-pools?limit=101")
		resp, body := client.Get(t, path)

		testutils.AssertStatus(t, resp, 400)

		var errorResponse struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
				Details []struct {
					Field   string `json:"field"`
					Message string `json:"message"`
				} `json:"details"`
			} `json:"error"`
		}
		testutils.UnmarshalResponse(t, body, &errorResponse)

		require.NotEmpty(t, errorResponse.Error.Message)
		require.Equal(t, "BAD_REQUEST", errorResponse.Error.Code)
		require.NotEmpty(t, errorResponse.Error.Details)
		t.Logf("Validation error: %s", errorResponse.Error.Message)
	})
}

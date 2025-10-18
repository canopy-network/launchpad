//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/postgres"
	"github.com/enielson/launchpad/internal/workers/fakevolume"
	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/enielson/launchpad/tests/testutils"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFakeVolumeWorker tests the fake volume generation worker
func TestFakeVolumeWorker(t *testing.T) {
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		ctx := context.Background()

		// Create test user
		user, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("fakevolume%d@example.com", time.Now().UnixNano())).
			WithUsername(fmt.Sprintf("fakevolumeuser%d", time.Now().Unix())).
			Create(ctx, db)
		require.NoError(t, err)

		// Create test chain with virtual_active status
		chain, err := fixtures.DefaultChain(user.ID).
			WithStatus(models.ChainStatusVirtualActive).
			Create(ctx, db)
		require.NoError(t, err)

		// Create virtual pool
		pool, err := fixtures.DefaultVirtualPool(chain.ID).Create(ctx, db)
		require.NoError(t, err)

		// Initialize repositories
		userRepo := postgres.NewUserRepository(db)
		templateRepo := postgres.NewChainTemplateRepository(db)
		chainRepo := postgres.NewChainRepository(db, userRepo, templateRepo)
		poolRepo := postgres.NewVirtualPoolRepository(db)

		// Initialize fake volume worker with very short interval for testing
		config := fakevolume.Config{
			Interval: 1 * time.Second,
		}
		worker := fakevolume.NewWorker(chainRepo, poolRepo, userRepo, config)

		// Start worker
		err = worker.Start()
		require.NoError(t, err)

		t.Logf("Created test chain %s with initial pool state: CNPY=%.2f, Tokens=%d",
			chain.ChainName, pool.CNPYReserve, pool.TokenReserve)

		// Wait for worker to generate some transactions
		time.Sleep(3 * time.Second)

		// Stop worker
		err = worker.Stop()
		require.NoError(t, err)

		// Verify transactions were created
		var transactionCount int
		err = db.GetContext(ctx, &transactionCount,
			"SELECT COUNT(*) FROM virtual_pool_transactions WHERE chain_id = $1", chain.ID)
		require.NoError(t, err)

		assert.Greater(t, transactionCount, 0, "Should have created at least 1 transaction")
		t.Logf("✅ Generated %d transactions", transactionCount)

		// Verify pool state was updated
		updatedPool, err := poolRepo.GetPoolByChainID(ctx, chain.ID)
		require.NoError(t, err)

		assert.NotEqual(t, pool.CNPYReserve, updatedPool.CNPYReserve, "CNPY reserve should have changed")
		assert.NotEqual(t, pool.TokenReserve, updatedPool.TokenReserve, "Token reserve should have changed")
		assert.NotEqual(t, pool.CurrentPriceCNPY, updatedPool.CurrentPriceCNPY, "Current price should have changed")
		assert.Equal(t, transactionCount, updatedPool.TotalTransactions, "Transaction count should match")

		t.Logf("✅ Pool state updated: CNPY=%.2f, Tokens=%d, Price=%.8f, TxCount=%d",
			updatedPool.CNPYReserve, updatedPool.TokenReserve, updatedPool.CurrentPriceCNPY, updatedPool.TotalTransactions)

		// Verify transaction types include both buys and sells (with enough transactions)
		if transactionCount >= 5 {
			var buyCount, sellCount int
			err = db.GetContext(ctx, &buyCount,
				"SELECT COUNT(*) FROM virtual_pool_transactions WHERE chain_id = $1 AND transaction_type = 'buy'", chain.ID)
			require.NoError(t, err)

			err = db.GetContext(ctx, &sellCount,
				"SELECT COUNT(*) FROM virtual_pool_transactions WHERE chain_id = $1 AND transaction_type = 'sell'", chain.ID)
			require.NoError(t, err)

			assert.Greater(t, buyCount, 0, "Should have at least 1 buy transaction")
			t.Logf("✅ Transaction types: %d buys, %d sells", buyCount, sellCount)
		}

		// Verify user positions were created
		var positionCount int
		err = db.GetContext(ctx, &positionCount,
			"SELECT COUNT(*) FROM user_virtual_positions WHERE chain_id = $1", chain.ID)
		require.NoError(t, err)

		assert.Greater(t, positionCount, 0, "Should have created at least 1 user position")
		t.Logf("✅ Created %d user positions", positionCount)

		// Cleanup
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM user_virtual_positions WHERE chain_id = $1", chain.ID)
			db.ExecContext(context.Background(),
				"DELETE FROM virtual_pool_transactions WHERE chain_id = $1", chain.ID)
			db.ExecContext(context.Background(),
				"DELETE FROM virtual_pools WHERE chain_id = $1", chain.ID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chain.ID)
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", user.ID)
		})
	})
}

// TestFakeVolumeWorkerMultipleChains tests fake volume generation across multiple chains
func TestFakeVolumeWorkerMultipleChains(t *testing.T) {
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		ctx := context.Background()

		// Create test user
		user, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("multichain%d@example.com", time.Now().UnixNano())).
			WithUsername(fmt.Sprintf("multichainuser%d", time.Now().Unix())).
			Create(ctx, db)
		require.NoError(t, err)

		// Create 3 test chains with unique names
		var chainIDs []interface{}
		for i := 0; i < 3; i++ {
			fixture := fixtures.DefaultChain(user.ID).WithStatus(models.ChainStatusVirtualActive)
			// Override chain name to make it unique
			fixture.ChainName = fmt.Sprintf("Fake Volume Test Chain %d %d", i, time.Now().UnixNano())

			chain, err := fixture.Create(ctx, db)
			require.NoError(t, err)

			_, err = fixtures.DefaultVirtualPool(chain.ID).Create(ctx, db)
			require.NoError(t, err)

			chainIDs = append(chainIDs, chain.ID)
			// Small delay to ensure unique nano timestamps
			time.Sleep(1 * time.Millisecond)
		}

		// Initialize repositories
		userRepo := postgres.NewUserRepository(db)
		templateRepo := postgres.NewChainTemplateRepository(db)
		chainRepo := postgres.NewChainRepository(db, userRepo, templateRepo)
		poolRepo := postgres.NewVirtualPoolRepository(db)

		// Initialize fake volume worker
		config := fakevolume.Config{
			Interval: 1 * time.Second,
		}
		worker := fakevolume.NewWorker(chainRepo, poolRepo, userRepo, config)

		// Start worker
		err = worker.Start()
		require.NoError(t, err)

		// Wait for worker to generate transactions
		time.Sleep(3 * time.Second)

		// Stop worker
		err = worker.Stop()
		require.NoError(t, err)

		// Verify all chains got transactions
		for i, chainID := range chainIDs {
			var transactionCount int
			err = db.GetContext(ctx, &transactionCount,
				"SELECT COUNT(*) FROM virtual_pool_transactions WHERE chain_id = $1", chainID)
			require.NoError(t, err)

			assert.Greater(t, transactionCount, 0, "Chain %d should have transactions", i+1)
			t.Logf("✅ Chain %d: %d transactions", i+1, transactionCount)
		}

		// Cleanup
		t.Cleanup(func() {
			for _, chainID := range chainIDs {
				db.ExecContext(context.Background(),
					"DELETE FROM user_virtual_positions WHERE chain_id = $1", chainID)
				db.ExecContext(context.Background(),
					"DELETE FROM virtual_pool_transactions WHERE chain_id = $1", chainID)
				db.ExecContext(context.Background(),
					"DELETE FROM virtual_pools WHERE chain_id = $1", chainID)
				db.ExecContext(context.Background(),
					"DELETE FROM chains WHERE id = $1", chainID)
			}
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", user.ID)
		})
	})
}

// TestFakeVolumeWorkerNoChains tests that worker handles no active chains gracefully
func TestFakeVolumeWorkerNoChains(t *testing.T) {
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		// Initialize repositories
		userRepo := postgres.NewUserRepository(db)
		templateRepo := postgres.NewChainTemplateRepository(db)
		chainRepo := postgres.NewChainRepository(db, userRepo, templateRepo)
		poolRepo := postgres.NewVirtualPoolRepository(db)

		// Initialize fake volume worker
		config := fakevolume.Config{
			Interval: 1 * time.Second,
		}
		worker := fakevolume.NewWorker(chainRepo, poolRepo, userRepo, config)

		// Start worker
		err := worker.Start()
		require.NoError(t, err)

		// Wait briefly
		time.Sleep(2 * time.Second)

		// Stop worker - should not crash
		err = worker.Stop()
		require.NoError(t, err)

		t.Logf("✅ Worker handled no active chains gracefully")
	})
}

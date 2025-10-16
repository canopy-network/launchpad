package fixtures_test

import (
	"context"
	"testing"

	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/enielson/launchpad/tests/testutils"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// Example test showing how to use fixtures with transaction-based isolation
func TestUserFixture_Example(t *testing.T) {
	t.Skip("Skipping example test - for documentation purposes only")
	testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create a user with default values
		user, err := fixtures.DefaultUser().
			WithEmail("alice@example.com").
			WithWallet("0xABCDEF").
			Create(ctx, tx)

		assert.NoError(t, err)
		assert.NotNil(t, user.ID)
		assert.Equal(t, "alice@example.com", *user.Email)
		assert.Equal(t, "0xABCDEF", user.WalletAddress)

		// Transaction automatically rolls back - no cleanup needed!
	})
}

// Example test showing how to create related entities
func TestChainWithVirtualPool_Example(t *testing.T) {
	t.Skip("Skipping example test - for documentation purposes only")
	testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create a user first (chains require a creator)
		user, err := fixtures.DefaultUser().Create(ctx, tx)
		assert.NoError(t, err)

		// Create a chain
		chain, err := fixtures.DefaultChain(user.ID).
			WithTokenSymbol("DEMO").
			WithStatus("virtual_active").
			Create(ctx, tx)
		assert.NoError(t, err)
		assert.Equal(t, "DEMO", chain.TokenSymbol)

		// Create a virtual pool for the chain
		pool, err := fixtures.DefaultVirtualPool(chain.ID).
			WithReserves(5000.0, 400000000).
			Create(ctx, tx)
		assert.NoError(t, err)
		assert.Equal(t, 5000.0, pool.CNPYReserve)
		assert.Equal(t, int64(400000000), pool.TokenReserve)

		// Create a user position in the pool
		position, err := fixtures.DefaultUserPosition(user.ID, chain.ID, pool.ID).
			WithPosition(1000000, 100.0, 0.0001).
			Create(ctx, tx)
		assert.NoError(t, err)
		assert.Equal(t, int64(1000000), position.TokenBalance)

		// All data rolls back automatically!
	})
}

// Example showing how to use sample data IDs
func TestUsingSampleData_Example(t *testing.T) {
	t.Skip("Skipping example test - for documentation purposes only")
	testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Use existing Alice user from sample_data.sql
		aliceID := fixtures.SampleDataIDs.Users.Alice

		// Create a new chain for Alice
		chain, err := fixtures.DefaultChain(aliceID).
			WithTemplate(fixtures.SampleDataIDs.Templates.DeFiStandard).
			Create(ctx, tx)

		assert.NoError(t, err)
		assert.Equal(t, aliceID, chain.CreatedBy)

		// Transaction rolls back - sample data is unchanged
	})
}

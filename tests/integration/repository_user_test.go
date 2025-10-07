//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/enielson/launchpad/tests/testutils"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// TestUserFixture_CreateAndRetrieve demonstrates using fixtures with automatic rollback
func TestUserFixture_CreateAndRetrieve(t *testing.T) {
	testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create a user with custom fields
		createdUser, err := fixtures.DefaultUser().
			WithEmail("integration@test.com").
			WithWallet("0xDEADBEEF").
			WithUsername("testuser123").
			Create(ctx, tx)

		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.NotNil(t, createdUser.ID)

		// Verify fields
		assert.Equal(t, "integration@test.com", *createdUser.Email)
		assert.Equal(t, "0xDEADBEEF", createdUser.WalletAddress)
		assert.Equal(t, "testuser123", *createdUser.Username)
		assert.True(t, createdUser.IsVerified)

		t.Logf("Created user with ID: %s", createdUser.ID)

		// Transaction rolls back automatically - no cleanup needed!
	})
}

// TestMultipleUsersInSameTransaction shows creating multiple fixtures
func TestMultipleUsersInSameTransaction(t *testing.T) {
	testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Create first user
		user1, err := fixtures.DefaultUser().
			WithEmail("user1@test.com").
			Create(ctx, tx)
		assert.NoError(t, err)

		// Create second user
		user2, err := fixtures.DefaultUser().
			WithEmail("user2@test.com").
			Create(ctx, tx)
		assert.NoError(t, err)

		// Verify they have different IDs
		assert.NotEqual(t, user1.ID, user2.ID)

		t.Logf("Created users: %s and %s", user1.ID, user2.ID)

		// Both automatically rolled back
	})
}

// TestUsingSampleDataUser demonstrates using pre-loaded sample data
func TestUsingSampleDataUser(t *testing.T) {
	testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Alice exists in sample_data.sql
		aliceID := fixtures.SampleDataIDs.Users.Alice

		// You can use Alice's ID to create chains, etc.
		chain, err := fixtures.DefaultChain(aliceID).
			WithTokenSymbol("ALICE").
			Create(ctx, tx)

		assert.NoError(t, err)
		assert.Equal(t, aliceID, chain.CreatedBy)

		t.Logf("Created chain for Alice: %s", chain.ID)

		// Chain creation rolls back, Alice remains in sample data
	})
}

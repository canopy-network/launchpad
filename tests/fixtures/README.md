# Test Fixtures Guide

## Overview

The fixtures package provides a clean, fluent API for creating test data in integration tests. All fixtures use **transaction-based isolation** to ensure tests don't pollute the database.

## Key Features

✅ **Automatic Rollback** - All changes rollback after each test
✅ **Fluent API** - Chain methods for readable test setup
✅ **Sensible Defaults** - Minimal configuration required
✅ **Works with Transactions** - Uses `sqlx.ExtContext` for flexibility

## Basic Usage

### Simple User Creation

```go
func TestUserCreation(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx, tx) {
        user, err := fixtures.DefaultUser().
            WithEmail("test@example.com").
            WithWallet("0x1234").
            Create(ctx, tx)

        assert.NoError(t, err)
        assert.NotNil(t, user.ID)

        // Transaction auto-rolls back - no cleanup!
    })
}
```

### Creating Related Entities

```go
func TestChainWithPool(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx, tx) {
        // 1. Create user (chain creator)
        user, err := fixtures.DefaultUser().Create(ctx, tx)
        require.NoError(t, err)

        // 2. Create chain
        chain, err := fixtures.DefaultChain(user.ID).
            WithTokenSymbol("MYTOKEN").
            WithStatus("virtual_active").
            Create(ctx, tx)
        require.NoError(t, err)

        // 3. Create virtual pool
        pool, err := fixtures.DefaultVirtualPool(chain.ID).
            WithReserves(10000.0, 800000000).
            Create(ctx, tx)
        require.NoError(t, err)

        // 4. Create chain key (for worker tests)
        key, err := fixtures.DefaultChainKey(chain.ID).
            WithAddress("aabbccdd11223344").
            Create(ctx, tx)
        require.NoError(t, err)

        // All data rolls back automatically
    })
}
```

## Available Fixtures

### UserFixture

```go
DefaultUser()
    .WithEmail(string)
    .WithWallet(string)
    .WithUsername(string)
    .Create(ctx, tx)
```

**Defaults:**
- Random UUID
- Generated wallet address
- Generated email (test-{uuid}@example.com)
- Verified user with "verified" tier

### ChainFixture

```go
DefaultChain(creatorID uuid.UUID)
    .WithTemplate(templateID uuid.UUID)
    .WithStatus(models.ChainStatus)
    .WithTokenSymbol(string)
    .WithBondingCurve(cnpy float64, tokens int64, slope float64)
    .Create(ctx, tx)
```

**Defaults:**
- Timestamped chain name
- "TEST" token symbol
- nestbft consensus
- 1B total supply
- Draft status
- 1000 CNPY reserve, 800M token supply

### VirtualPoolFixture

```go
DefaultVirtualPool(chainID uuid.UUID)
    .WithReserves(cnpy float64, tokens int64)
    .Create(ctx, tx)
```

**Defaults:**
- 1000 CNPY reserve
- 800M token reserve
- Auto-calculated price
- Active pool

### ChainKeyFixture

```go
DefaultChainKey(chainID uuid.UUID)
    .WithAddress(string)
    .WithPurpose(string)
    .Create(ctx, tx)
```

**Defaults:**
- Hex address from chain ID
- Mock encrypted key
- "treasury" purpose
- Active key

### UserPositionFixture

```go
DefaultUserPosition(userID, chainID, poolID uuid.UUID)
    .WithPosition(tokenBalance int64, cnpyInvested, entryPrice float64)
    .Create(ctx, tx)
```

**Defaults:**
- Zero balance
- Zero investment
- Active position

## Real-World Examples

### Testing Bonding Curve Calculations

```go
func TestBondingCurveCalculations(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx, tx) {
        user, _ := fixtures.DefaultUser().Create(ctx, tx)

        chain, _ := fixtures.DefaultChain(user.ID).
            WithBondingCurve(1000.0, 800000000, 0.00000001).
            Create(ctx, tx)

        pool, _ := fixtures.DefaultVirtualPool(chain.ID).
            WithReserves(1000.0, 800000000).
            Create(ctx, tx)

        // Test buy calculations
        buyAmount := 100.0 // 100 CNPY
        tokensReceived := calculateBuy(pool, buyAmount)

        assert.Greater(t, tokensReceived, int64(0))
    })
}
```

### Testing Worker Processing

```go
func TestNewBlockWorker(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx, tx) {
        // Setup: Create chain with address
        user, _ := fixtures.DefaultUser().Create(ctx, tx)
        chain, _ := fixtures.DefaultChain(user.ID).
            WithStatus("virtual_active").
            Create(ctx, tx)

        // Create chain key with specific address
        key, _ := fixtures.DefaultChainKey(chain.ID).
            WithAddress("aabbccdd11223344").
            Create(ctx, tx)

        // Create pool
        pool, _ := fixtures.DefaultVirtualPool(chain.ID).Create(ctx, tx)

        // Simulate worker processing transaction to this address
        // ... worker logic ...

        // Verify pool updated
        // ... assertions ...
    })
}
```

### Testing Concurrent Operations

```go
func TestConcurrentBuys(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx, tx) {
        // Create base setup
        user1, _ := fixtures.DefaultUser().Create(ctx, tx)
        user2, _ := fixtures.DefaultUser().Create(ctx, tx)

        chain, _ := fixtures.DefaultChain(user1.ID).
            WithStatus("virtual_active").
            Create(ctx, tx)

        pool, _ := fixtures.DefaultVirtualPool(chain.ID).Create(ctx, tx)

        // Test concurrent operations
        // ... concurrency tests ...
    })
}
```

## Test Utilities

### WithTestTransaction

The recommended way to run tests with automatic rollback:

```go
testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
    // All database operations in here
    // Automatic rollback after test
})
```

### WithTestDB

For tests that need a regular DB connection (less common):

```go
testutils.WithTestDB(t, func(db *sqlx.DB) {
    // Use db connection
    // Manual cleanup required!
})
```

## Best Practices

### ✅ DO

- Use `WithTestTransaction` for all integration tests
- Chain fixture methods for readability
- Create minimal fixtures - only what you need
- Test one thing per test function

### ❌ DON'T

- Manually clean up data (transactions handle it)
- Create fixtures outside transactions
- Reuse fixture instances across tests
- Hardcode UUIDs (use fixtures)
- Mix fixture creation and business logic

## Migration from Old Tests

**Before (hardcoded UUIDs):**
```go
chainID := uuid.MustParse("550e8400-e29b-41d4-a716-446655442001")
```

**After (using fixtures):**
```go
testutils.WithTestTransaction(t, func(ctx, tx) {
    user, _ := fixtures.DefaultUser().Create(ctx, tx)
    chain, _ := fixtures.DefaultChain(user.ID).Create(ctx, tx)
    chainID := chain.ID
})
```

## Running Tests

```bash
# Run all integration tests
make test

# Run specific fixture tests
go test -v -tags=integration ./tests/fixtures/...

# Run with coverage
make test-coverage
```

## Troubleshooting

**Q: My test changes aren't rolling back**
- Make sure you're using `WithTestTransaction`, not `WithTestDB`
- Check that you're passing `tx` to fixture `.Create()` calls

**Q: Foreign key constraint errors**
- Create parent entities first (user before chain, chain before pool)

**Q: Fixture creation fails**
- Check database connection in test environment
- Verify schema is up to date (`make migrate-status`)
- Ensure required fields have values

**Q: Tests are slow**
- Each transaction adds overhead - keep tests focused
- Run tests in parallel where possible

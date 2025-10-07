# Test Fixtures - Usage Summary

## What We Built

A **hybrid approach** for integration test fixtures that combines:
1. **Transaction-based isolation** - All tests automatically rollback
2. **Fluent API builders** - Clean, readable test setup
3. **Direct SQL inserts** - Fast fixture creation without HTTP overhead
4. **Works with both `*sqlx.DB` and `*sqlx.Tx`** - Maximum flexibility

## Quick Start

### 1. Basic User Creation

```go
func TestMyFeature(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
        user, err := fixtures.DefaultUser().
            WithEmail("test@example.com").
            Create(ctx, tx)

        assert.NoError(t, err)
        // Test your feature...

        // Transaction rolls back automatically!
    })
}
```

### 2. Creating Related Entities

```go
func TestChainWithPool(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
        // Create user (chain creator)
        user, _ := fixtures.DefaultUser().Create(ctx, tx)

        // Create chain
        chain, _ := fixtures.DefaultChain(user.ID).
            WithTokenSymbol("MYTOKEN").
            WithStatus(models.ChainStatusVirtualActive).
            Create(ctx, tx)

        // Create virtual pool
        pool, _ := fixtures.DefaultVirtualPool(chain.ID).
            WithReserves(10000.0, 800000000).
            Create(ctx, tx)

        // Everything rolls back!
    })
}
```

### 3. Using Sample Data

```go
func TestWithSampleData(t *testing.T) {
    testutils.WithTestTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
        // Use Alice from sample_data.sql
        aliceID := fixtures.SampleDataIDs.Users.Alice

        chain, _ := fixtures.DefaultChain(aliceID).Create(ctx, tx)

        // Alice remains in sample data (tx rolls back)
    })
}
```

## Key Benefits

### ✅ Automatic Cleanup
No manual cleanup code needed - transactions roll back after each test

### ✅ Perfect Isolation
Tests never interfere with each other or pollute the database

### ✅ Fast Execution
Direct SQL inserts are faster than creating data through HTTP APIs

### ✅ Readable Tests
Fluent API makes test setup intention-clear

### ✅ Type Safe
Full IDE autocomplete and compile-time checking

## Available Fixtures

| Fixture | Builder Pattern | Example |
|---------|-----------------|---------|
| **User** | `DefaultUser()` | `.WithEmail("test@example.com")` |
| **Chain** | `DefaultChain(creatorID)` | `.WithTokenSymbol("TOKEN").WithStatus("virtual_active")` |
| **VirtualPool** | `DefaultVirtualPool(chainID)` | `.WithReserves(1000.0, 800000000)` |
| **ChainKey** | `DefaultChainKey(chainID)` | `.WithAddress("aabbccdd").WithPurpose("treasury")` |
| **UserPosition** | `DefaultUserPosition(userID, chainID, poolID)` | `.WithPosition(1000000, 100.0, 0.0001)` |

## File Structure

```
tests/
├── fixtures/
│   ├── fixtures.go          # Fixture builders with SQL inserts
│   ├── example_test.go      # Example usage patterns
│   └── README.md            # Comprehensive documentation
├── testutils/
│   ├── database.go          # Transaction helpers
│   └── http.go              # HTTP test client (existing)
└── integration/
    └── repository_user_test.go  # Example repository tests
```

## How It Works

### 1. Transaction Wrapper
```go
func WithTestTransaction(t *testing.T, fn func(context.Context, *sqlx.Tx)) {
    db, _ := database.Connect(TestDatabaseURL)
    defer db.Close()

    tx, _ := db.Beginx()
    defer tx.Rollback() // Always rollback for test isolation

    fn(context.Background(), tx)
}
```

### 2. Fixture Builders
Each fixture uses the builder pattern with fluent methods:
- Sensible defaults for all fields
- Optional `.WithX()` methods for customization
- `.Create(ctx, tx)` executes INSERT and returns the model

### 3. SQL Execution
Fixtures use `sqlx.ExtContext` interface to work with both `*sqlx.DB` and `*sqlx.Tx`:
```go
func (u *UserFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.User, error) {
    // Works with both DB and TX!
}
```

## Running Tests

```bash
# Run all integration tests
make test

# Run specific fixture tests
go test -v -tags=integration -run TestUserFixture ./tests/integration

# With coverage
make test-coverage
```

## Comparison with Other Approaches

### Before (Hardcoded UUIDs)
```go
chainID := uuid.MustParse("550e8400-e29b-41d4-a716-446655442001")
// Brittle, depends on sample data not changing
```

### After (Fixtures)
```go
testutils.WithTestTransaction(t, func(ctx, tx) {
    chain, _ := fixtures.DefaultChain(user.ID).Create(ctx, tx)
    chainID := chain.ID
    // Clean, isolated, self-contained
})
```

## Best Practices

### ✅ DO
- Always use `WithTestTransaction` for integration tests
- Create minimal fixtures - only what you need for the test
- Chain builder methods for readability
- Use `SampleDataIDs` for referencing pre-loaded data

### ❌ DON'T
- Manually clean up data (transactions handle it)
- Create fixtures outside transactions
- Hardcode UUIDs (use fixtures or sample data IDs)
- Reuse fixture instances across tests

## Examples in Codebase

See these files for real examples:
- `tests/fixtures/example_test.go` - Comprehensive examples
- `tests/integration/repository_user_test.go` - Real integration tests
- `tests/fixtures/README.md` - Full documentation

## Next Steps

To add fixtures for new entities:
1. Add a fixture struct in `tests/fixtures/fixtures.go`
2. Create a `Default{Entity}()` constructor with sensible defaults
3. Add `.WithX()` builder methods as needed
4. Implement `.Create(ctx, db)` with INSERT ... RETURNING
5. Write example tests

## Summary

This hybrid approach gives you:
- **Speed** of direct SQL
- **Safety** of transactions
- **Readability** of builder pattern
- **Flexibility** to work with DB or TX

Perfect for integration tests that need real database interaction without pollution!

# Centralized Mock Repositories

This package contains mock implementations of repository interfaces for testing purposes.

## Purpose

Consolidating mocks in a single location provides:
- **Single source of truth** - Update once, use everywhere
- **Reduced duplication** - Previously ~640 lines of duplicate mock code
- **Easier maintenance** - Interface changes only require one update
- **Consistent behavior** - All tests use the same mock implementations

## Current Mocks

### repositories.go
- `MockChainRepository` - Mock implementation of `interfaces.ChainRepository`
- `MockVirtualPoolRepository` - Mock implementation of `interfaces.VirtualPoolRepository`
- `MockUserRepository` - Mock implementation of `interfaces.UserRepository`
- `MockVirtualPoolTxRepository` - Mock with transaction support (embeds `MockVirtualPoolRepository`)

## Usage

```go
import (
    "testing"
    "github.com/enielson/launchpad/internal/testutil/mocks"
    "github.com/stretchr/testify/mock"
)

func TestMyFunction(t *testing.T) {
    // Create mock
    chainRepo := new(mocks.MockChainRepository)

    // Set expectations
    chainRepo.On("GetByID", mock.Anything, chainID, []string{}).Return(chain, nil)

    // Use in code
    result, err := myService.DoSomething(chainRepo)

    // Verify expectations
    chainRepo.AssertExpectations(t)
}
```

## Files Currently Using These Mocks

- ✅ `internal/graduator/graduator_test.go` - **Refactored** (645 lines → 413 lines, 36% reduction)
- `internal/workers/newblock/worker_test.go` - Still uses local mocks
- `internal/services/order_processor_test.go` - Still uses local mocks
- `internal/services/order_processor_tx_test.go` - Still uses local mocks

## Future Recommendation: Use Mockery

For long-term maintainability, consider using [mockery](https://github.com/vektra/mockery) to auto-generate mocks:

### Why Mockery?

- **Automated generation** - No manual mock maintenance
- **Always in sync** - Regenerate when interfaces change
- **Industry standard** - Used by Go projects worldwide
- **Zero drift** - Impossible for mocks to become out of date

### How to Set Up Mockery

1. **Install mockery:**
```bash
go install github.com/vektra/mockery/v2@latest
```

2. **Add generation comments to interface files:**
```go
// internal/repository/interfaces/chain.go

//go:generate mockery --name=ChainRepository --output=../../testutil/mocks --outpkg=mocks
type ChainRepository interface {
    // ... methods
}
```

3. **Generate mocks:**
```bash
go generate ./...
```

4. **Add to Makefile:**
```makefile
.PHONY: generate-mocks
generate-mocks:
\tgo generate ./internal/repository/interfaces/...
```

### Migration Path

If you decide to adopt mockery:

1. Add `//go:generate` comments to interface files
2. Run `go generate` to create mocks
3. Gradually migrate test files to use generated mocks
4. Remove this manually-maintained `repositories.go` file

## Maintenance

When interfaces change:

**Current approach:**
1. Update the interface in `internal/repository/interfaces/`
2. Update the mock in `internal/testutil/mocks/repositories.go`
3. Update any tests that break

**With mockery (recommended):**
1. Update the interface
2. Run `go generate ./...`
3. Update any tests that break (mocks auto-update)

## Notes

- These mocks use `github.com/stretchr/testify/mock`
- All mocks embed `mock.Mock` for expectation/assertion support
- Nil checks are included for pointer returns to prevent panics
- Transaction-aware mocks (like `MockVirtualPoolTxRepository`) support sqlx.Tx parameters

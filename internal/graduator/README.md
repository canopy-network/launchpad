# Graduator Package

The graduator package handles the virtual chain graduation process for the launchpad.

## Overview

The graduator initiates the virtual chain graduation process once the total CNPY of the virtual pool exceeds the graduation threshold (default: 50,000 CNPY).

## Responsibilities

### Virtual Pool Value Check
- Checks if the total CNPY value in the virtual pool exceeds the graduation threshold
- Compares `virtual_pools.cnpy_reserve` against `chains.graduation_threshold`
- Returns error if threshold is not met

### Genesis File Creation
- Processes `templates/genesis/genesis.json.template`
- Queries `user_virtual_positions` table for all positions with `token_balance > 0`
- Joins with `users` table to get wallet addresses
- Generates genesis.json with account balances
- Outputs to stdout

## Usage

```go
import (
    "context"
    "github.com/enielson/launchpad/internal/graduator"
    "github.com/enielson/launchpad/internal/repository/postgres"
    "github.com/enielson/launchpad/pkg/database"
    "github.com/google/uuid"
)

func main() {
    // Setup database and repositories
    db := database.Connect(databaseURL)
    chainRepo := postgres.NewChainRepository(db, userRepo, templateRepo)
    virtualPoolRepo := postgres.NewVirtualPoolRepository(db)

    // Create graduator
    grad := graduator.New(
        chainRepo,
        virtualPoolRepo,
        "templates/genesis/genesis.json.template",
    )

    // Check and graduate a chain
    chainID := uuid.MustParse("your-chain-id")
    err := grad.CheckAndGraduate(context.Background(), chainID)
    if err != nil {
        log.Fatalf("Failed to graduate chain: %v", err)
    }
}
```

## API

### `New(chainRepo, virtualPoolRepo, templatePath) *Graduator`
Creates a new Graduator instance.

### `CheckAndGraduate(ctx, chainID) error`
Checks if a chain is eligible for graduation and performs the graduation process:
1. Verifies chain exists and is not already graduated
2. Checks if virtual pool CNPY reserve meets graduation threshold
3. Generates genesis file if eligible

Returns error if:
- Chain not found
- Chain already graduated
- Virtual pool not found
- Graduation threshold not met
- Genesis file generation fails

### `GenerateGenesisFile(ctx, chainID) error`
Generates the genesis.json file for a chain:
1. Queries all user positions with token_balance > 0
2. Joins with users table to get wallet addresses
3. Processes the genesis template
4. Outputs to stdout

## Data Structures

### `GenesisAccount`
```go
type GenesisAccount struct {
    Address string  // User wallet address
    Amount  int64   // Token balance
}
```

### `GenesisData`
```go
type GenesisData struct {
    Accounts []GenesisAccount
}
```

## Template Format

The genesis template at `templates/genesis/genesis.json.template` uses Go's `text/template` package:

```json
{
    "accounts": [
        {{- range $index, $account := .Accounts }}
        {{- if $index }},{{ end }}
        {
            "address": "{{ $account.Address }}",
            "amount": {{ $account.Amount }}
        }
        {{- end }}
    ]
}
```

## Database Tables

- `chains` - graduation_threshold, is_graduated, graduation_time
- `virtual_pools` - cnpy_reserve
- `user_virtual_positions` - token_balance (JOIN with users for wallet_address)
- `users` - wallet_address

## Error Handling

The graduator returns descriptive errors for:
- Chain not found
- Already graduated chains
- Virtual pool not found
- Graduation threshold not met
- Database query failures
- Template parsing errors

## Future Enhancements

Potential additions not in current requirements:
- Automatic graduation status update in database
- Genesis file validation
- Notification system for graduations
- Graduation history tracking

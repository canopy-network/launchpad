# Canopy RPC Client

A standalone RPC client for interacting with Canopy blockchain nodes.

## Overview

This package provides a clean HTTP client for querying Canopy blockchain data. It was extracted from the main Canopy repository to provide a lightweight, reusable client without server dependencies.

## Installation

```go
import "github.com/enielson/launchpad/pkg/client/canopy"
```

## Usage

### Creating a Client

```go
client := canopy.NewClient("http://localhost:42069")
```

### Querying Blockchain Data

```go
// Get current height
height, err := client.Height()

// Get chain information
chainInfo, err := client.RootChainInfo(0, chainID)

// Get order book
orders, err := client.Orders(0, chainID)

// Get specific order
order, err := client.Order(0, orderID, chainID)

// Get account information
account, err := client.Account(height, address)

// Get block by height
block, err := client.BlockByHeight(height)
```

### Submitting Transactions

```go
// Submit a transaction
hash, err := client.Transaction(tx)

// Submit raw JSON transaction
hash, err := client.TransactionJSON(jsonTx)
```

### Pagination

Many query methods support pagination:

```go
pageParams := lib.PageParams{
    PageNumber: 1,
    PerPage:    100,
}

// Get paginated transactions
txs, err := client.TransactionsByHeight(height, pageParams)

// Get paginated accounts
accounts, err := client.Accounts(height, pageParams)

// Get paginated validators
validators, err := client.Validators(height, pageParams, lib.ValidatorFilters{})
```

## API Methods

### Blockchain Queries
- `Version()` - Get node version
- `Height()` - Get current blockchain height
- `BlockByHeight(height)` - Get block at height
- `BlockByHash(hash)` - Get block by hash
- `Blocks(params)` - Get paginated blocks

### Transaction Queries
- `TransactionByHash(hash)` - Get transaction by hash
- `TransactionsByHeight(height, params)` - Get transactions at height
- `TransactionsBySender(address, params)` - Get transactions from sender
- `TransactionsByRecipient(address, params)` - Get transactions to recipient
- `Pending(params)` - Get pending transactions

### State Queries
- `Account(height, address)` - Get account state
- `Accounts(height, params)` - Get all accounts
- `Validator(height, address)` - Get validator info
- `Validators(height, params, filters)` - Get validators
- `Pool(height, id)` - Get liquidity pool
- `Pools(height, params)` - Get all pools
- `State(height)` - Get full state snapshot
- `StateDiff(height, startHeight)` - Get state diff

### Committee Queries
- `Committee(height, id, params)` - Get committee members
- `CommitteeData(height, id)` - Get committee data
- `CommitteesData(height)` - Get all committees data
- `SubsidizedCommittees(height)` - Get subsidized committees
- `RetiredCommittees(height)` - Get retired committees

### Order Book Queries
- `Order(height, orderId, chainId)` - Get specific order
- `Orders(height, chainId)` - Get all orders for chain
- `RootChainInfo(height, chainId)` - Get root chain information

### Consensus Queries
- `CertByHeight(height)` - Get quorum certificate
- `ValidatorSet(height, id)` - Get validator set
- `LastProposers(height)` - Get recent proposers
- `DoubleSigners(height)` - Get double signers
- `MinimumEvidenceHeight(height)` - Get minimum evidence height
- `Checkpoint(height, id)` - Get checkpoint data

### Governance Queries
- `Proposals()` - Get governance proposals
- `Poll()` - Get active poll
- `Params(height)` - Get all parameters
- `FeeParams(height)` - Get fee parameters
- `GovParams(height)` - Get governance parameters
- `ConParams(height)` - Get consensus parameters
- `ValParams(height)` - Get validator parameters

### Other Queries
- `Supply(height)` - Get token supply
- `NonSigners(height)` - Get non-signing validators
- `Lottery(height, id)` - Get lottery winner

## Files

- `client.go` - Main client implementation with all query methods
- `routes.go` - Route path constants and HTTP method mappings
- `types.go` - Request/response type definitions
- `client_test.go` - Unit tests

## Dependencies

This client depends on the following Canopy packages:
- `github.com/canopy-network/canopy/lib` - Core types and utilities
- `github.com/canopy-network/canopy/fsm` - State machine types (Account, Validator, etc.)
- `github.com/canopy-network/canopy/lib/crypto` - Cryptographic types

These are available via the go.mod replace directive pointing to `../canopy`.

## Error Handling

All methods return `lib.ErrorI` interface for errors. Common error types:
- HTTP request/response errors
- JSON marshaling/unmarshaling errors
- Non-200 HTTP status codes (wrapped with body)

## Testing

Run the test suite:

```bash
go test ./pkg/client/canopy -v
```

## Example: Order Book Monitoring

```go
package main

import (
    "log"
    "time"

    "github.com/enielson/launchpad/pkg/client/canopy"
)

func main() {
    client := canopy.NewClient("http://localhost:42069")
    chainID := uint64(1)

    for {
        // Get current height
        height, err := client.Height()
        if err != nil {
            log.Printf("Error getting height: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        // Get orders for chain
        orders, err := client.Orders(*height, chainID)
        if err != nil {
            log.Printf("Error getting orders: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        log.Printf("Height %d: Found %d sell orders", *height, len(orders.SellOrders))

        time.Sleep(5 * time.Second)
    }
}
```

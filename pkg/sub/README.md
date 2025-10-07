# Root Chain Subscription Package

This package provides a Go library for subscribing to real-time root chain events via WebSocket connections. It's extracted from the Canopy blockchain project and designed to be reusable in other applications.

## Features

- **Real-time WebSocket subscriptions** to root chain events
- **Automatic reconnection** with exponential backoff retry logic
- **Thread-safe operations** with proper mutex protection
- **Multiple subscription management** via the Manager component
- **Customizable logging** through the Logger interface
- **Event handling** through callback functions
- **Connection status monitoring**

## Installation

```bash
go get github.com/gorilla/websocket
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"

    "your-project/sub"
)

func main() {
    // Create a manager
    manager := sub.NewManager(nil) // Uses default logger

    // Define event handler
    handler := func(info *sub.RootChainInfo) error {
        fmt.Printf("New block: Height=%d, ChainId=%d\n",
            info.Height, info.RootChainId)
        return nil
    }

    // Add subscription
    config := sub.Config{
        ChainId: 1,
        Url:     "ws://localhost:8081",
    }

    err := manager.AddSubscription(config, handler)
    if err != nil {
        log.Fatal(err)
    }

    // Let it run
    time.Sleep(30 * time.Second)

    // Clean shutdown
    manager.StopAll()
}
```

## Core Components

### RootChainInfo

Contains the information received from root chain subscriptions:

```go
type RootChainInfo struct {
    RootChainId      uint64 `json:"rootChainId"`
    Height           uint64 `json:"height"`
    ValidatorSet     []byte `json:"validatorSet,omitempty"`
    LastValidatorSet []byte `json:"lastValidatorSet,omitempty"`
    LotteryWinner    []byte `json:"lotteryWinner,omitempty"`
    Orders           []byte `json:"orders,omitempty"`
    Timestamp        uint64 `json:"timestamp,omitempty"`
}
```

### Subscription

Manages a single WebSocket connection to a root chain:

```go
// Create subscription
subscription := sub.NewSubscription(config, handler, logger)

// Start listening
subscription.Start()

// Check connection status
if subscription.IsConnected() {
    // Get latest cached info
    info := subscription.GetLatestInfo()
}

// Stop subscription
subscription.Stop()
```

### Manager

Handles multiple subscriptions:

```go
manager := sub.NewManager(logger)

// Add multiple subscriptions
manager.AddSubscription(config1, handler1)
manager.AddSubscription(config2, handler2)

// Get status of all subscriptions
status := manager.GetStatus()

// Get latest info for specific chain
info, found := manager.GetLatestInfo(chainId)

// Stop all subscriptions
manager.StopAll()
```

## Configuration

```go
type Config struct {
    ChainId uint64 `json:"chainId"` // Chain ID to subscribe to
    Url     string `json:"url"`     // WebSocket URL (e.g., "ws://localhost:8081")
}
```

## Event Handling

Define custom event handlers to process received data:

```go
type EventHandler func(info *RootChainInfo) error

handler := func(info *RootChainInfo) error {
    // Process the received root chain information
    // Examples:
    // - Update local database
    // - Trigger business logic
    // - Send notifications
    // - Update UI state

    if info.Height > lastKnownHeight {
        fmt.Printf("New block detected: %d\n", info.Height)
        // Handle new block
    }

    return nil // Return error if processing fails
}
```

## Custom Logging

Implement the Logger interface for custom logging:

```go
type Logger interface {
    Infof(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Error(msg string)
    Fatal(msg string)
    Warnf(format string, args ...interface{})
}

// Example with logrus
type LogrusLogger struct {
    logger *logrus.Logger
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
    l.logger.Infof(format, args...)
}
// ... implement other methods
```

## Error Handling

The package provides robust error handling:

- **Connection failures**: Automatic reconnection with exponential backoff
- **Message parsing errors**: Logged and skipped, connection continues
- **Handler errors**: Logged but don't interrupt the subscription
- **Graceful shutdown**: Proper cleanup of resources

## Thread Safety

All operations are thread-safe:

- Multiple goroutines can safely call manager methods
- Subscription state is protected with mutexes
- Event handlers are called sequentially per subscription

## WebSocket Protocol

The package connects to the `/v1/subscribe-rc-info` endpoint with:
- Query parameter: `chainId=<chain_id>`
- Protocol: WebSocket upgrade from HTTP
- Message format: JSON (can be extended to protobuf)

## Connection Management

- **Automatic reconnection** on connection loss
- **Exponential backoff** retry logic (starts at 1 second, doubles each retry)
- **Maximum retry attempts** (default: 25 attempts)
- **Connection status monitoring**
- **Graceful shutdown** with resource cleanup

## Examples

See `example.go` for complete usage examples:

- `ExampleUsage()`: Multiple subscriptions with manager
- `ExampleSingleSubscription()`: Single subscription usage

## Dependencies

- `github.com/gorilla/websocket`: WebSocket client implementation
- Standard Go libraries: `sync`, `time`, `net/url`, `encoding/json`

## License

This package is extracted from the Canopy blockchain project. Please refer to the original project's license terms.
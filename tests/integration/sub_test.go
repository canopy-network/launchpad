//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/canopy-network/canopy/lib"
	"github.com/enielson/launchpad/pkg/sub"
)

// testLogger implements sub.Logger interface using testing.T
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Infof(format string, args ...interface{}) {
	l.t.Logf("INFO: "+format, args...)
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	l.t.Logf("ERROR: "+format, args...)
}

func (l *testLogger) Error(msg string) {
	l.t.Logf("ERROR: %s", msg)
}

func (l *testLogger) Fatal(msg string) {
	l.t.Fatalf("FATAL: %s", msg)
}

func (l *testLogger) Warnf(format string, args ...interface{}) {
	l.t.Logf("WARN: "+format, args...)
}

// TestLiveCanopyConnection tests connection to a live Canopy node WebSocket server
// This test connects to a running Canopy instance and subscribes to root chain updates
// Run with: go test -v -tags=integration -run TestLiveCanopyConnection ./tests/integration/...
//
// Prerequisites:
//   - Running Canopy node at ws://localhost:50002
//   - Node must have chain ID 1 configured
//
// The test validates:
//   - WebSocket connection establishment
//   - Receiving RootChainInfo protobuf messages
//   - Message format and field validation
//   - Graceful connection shutdown
func TestLiveCanopyConnection(t *testing.T) {
	// Canopy node WebSocket endpoint
	// Note: The subscription.go code will append the path and chainId query param
	canopyURL := "ws://localhost:50002"
	chainID := uint64(1)

	t.Logf("Testing connection to Canopy node at %s for root chain ID %d", canopyURL, chainID)

	// Track received messages
	messageCount := 0
	receivedInfo := make(chan *lib.RootChainInfo, 5)

	// Create event handler to process RootChainInfo messages from Canopy
	handler := func(info *lib.RootChainInfo) error {
		messageCount++
		t.Logf("Received RootChainInfo from Canopy #%d: RootChainId=%d, Height=%d, Timestamp=%d",
			messageCount, info.RootChainId, info.Height, info.Timestamp)

		// Print OrderBook information if present
		if info.Orders != nil {
			t.Logf("  OrderBook: ChainId=%d, OrderCount=%d", info.Orders.ChainId, len(info.Orders.Orders))
			for i, order := range info.Orders.Orders {
				t.Logf("    Order[%d]: AmountForSale=%d, RequestedAmount=%d, Committee=%d",
					i, order.AmountForSale, order.RequestedAmount, order.Committee)
			}
		} else {
			t.Logf("  OrderBook: nil (no orders)")
		}

		select {
		case receivedInfo <- info:
		default:
		}

		return nil
	}

	// Create subscription config for Canopy node
	config := sub.Config{
		ChainId: chainID,
		Url:     canopyURL,
	}

	// Create custom logger for test
	logger := &testLogger{t: t}

	// Create and start subscription
	subscription := sub.NewSubscription(config, handler, logger)
	err := subscription.Start()
	if err != nil {
		t.Fatalf("Failed to start subscription: %v", err)
	}
	defer subscription.Stop()

	// Wait for connection to Canopy node with timeout
	connectionTimeout := time.After(10 * time.Second)
	connectionCheck := time.NewTicker(100 * time.Millisecond)
	defer connectionCheck.Stop()

	t.Log("Waiting for WebSocket connection to Canopy node...")
	connected := false
	for !connected {
		select {
		case <-connectionTimeout:
			t.Fatal("Timed out waiting for connection to Canopy node")
		case <-connectionCheck.C:
			if subscription.IsConnected() {
				connected = true
				t.Log("Successfully connected to Canopy node via WebSocket")
			}
		}
	}

	// Wait for at least one RootChainInfo protobuf message from Canopy
	messageTimeout := time.After(30 * time.Second)
	t.Log("Waiting for RootChainInfo protobuf messages from Canopy node...")

	select {
	case info := <-receivedInfo:
		t.Logf("Successfully received and decoded RootChainInfo protobuf: RootChainId=%d, Height=%d, Timestamp=%d",
			info.RootChainId, info.Height, info.Timestamp)

		// Validate the received RootChainInfo from Canopy
		if info.RootChainId != chainID {
			t.Errorf("Expected RootChainId %d from Canopy, got %d", chainID, info.RootChainId)
		}
		if info.Height == 0 {
			t.Error("Expected non-zero block height from Canopy")
		}
		if info.Timestamp == 0 {
			t.Error("Expected non-zero timestamp from Canopy")
		}

		// Verify cached latest info from subscription
		latestInfo := subscription.GetLatestInfo()
		if latestInfo.Height != info.Height {
			t.Errorf("Cached info height mismatch: expected %d, got %d", info.Height, latestInfo.Height)
		}

	case <-messageTimeout:
		t.Fatal("Timed out waiting for RootChainInfo messages from Canopy node")
	}

	// Continue collecting messages to verify continuous streaming from Canopy
	t.Log("Collecting additional RootChainInfo updates for 5 seconds...")
	additionalWait := time.After(5 * time.Second)
	additionalMessages := 0

waitLoop:
	for {
		select {
		case info := <-receivedInfo:
			additionalMessages++
			t.Logf("Additional RootChainInfo update #%d: Height=%d", additionalMessages, info.Height)
		case <-additionalWait:
			break waitLoop
		}
	}

	t.Logf("Canopy integration test completed. Total RootChainInfo messages received: %d", messageCount)

	// Test graceful shutdown of WebSocket connection
	t.Log("Testing graceful shutdown of Canopy subscription...")
	err = subscription.Stop()
	if err != nil {
		t.Errorf("Error during subscription shutdown: %v", err)
	}

	// Verify subscription is fully disconnected from Canopy node
	time.Sleep(100 * time.Millisecond)
	if subscription.IsConnected() {
		t.Error("Subscription should be disconnected from Canopy after Stop()")
	}

	t.Log("Canopy node integration test completed successfully")
}

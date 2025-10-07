package sub

import (
	"testing"
	"time"

	"github.com/canopy-network/canopy/lib"
)

// testLogger implements Logger interface using testing.T
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

func TestRetryMechanism(t *testing.T) {
	retry := NewRetry(100, 3) // 100ms initial wait, 3 max loops

	// First call should not wait
	if !retry.WaitAndDoRetry() {
		t.Error("First retry attempt should succeed")
	}

	// Second call should wait 100ms
	start := time.Now()
	if !retry.WaitAndDoRetry() {
		t.Error("Second retry attempt should succeed")
	}
	elapsed := time.Since(start)
	if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected ~100ms wait, got %v", elapsed)
	}

	// Third call should wait 200ms (doubled)
	start = time.Now()
	if !retry.WaitAndDoRetry() {
		t.Error("Third retry attempt should succeed")
	}
	elapsed = time.Since(start)
	if elapsed < 180*time.Millisecond || elapsed > 250*time.Millisecond {
		t.Errorf("Expected ~200ms wait, got %v", elapsed)
	}

	// Fourth call should fail (exceeded max loops)
	if retry.WaitAndDoRetry() {
		t.Error("Fourth retry attempt should fail")
	}
}

func TestSubscriptionCreation(t *testing.T) {
	config := Config{
		ChainId: 1,
		Url:     "ws://localhost:8081",
	}

	handler := func(info *lib.RootChainInfo) error {
		return nil
	}

	subscription := NewSubscription(config, handler, nil)

	if subscription.chainId != 1 {
		t.Errorf("Expected chainId 1, got %d", subscription.chainId)
	}

	if subscription.IsConnected() {
		t.Error("New subscription should not be connected initially")
	}

	// Test getting latest info (should be empty initially)
	info := subscription.GetLatestInfo()
	if info.Height != 0 {
		t.Errorf("Expected height 0, got %d", info.Height)
	}
}

func TestManagerOperations(t *testing.T) {
	manager := NewManager(nil)

	config := Config{
		ChainId: 1,
		Url:     "ws://localhost:8081",
	}

	handler := func(info *lib.RootChainInfo) error {
		return nil
	}

	// Add subscription
	err := manager.AddSubscription(config, handler)
	if err != nil {
		t.Errorf("Failed to add subscription: %v", err)
	}

	// Check if subscription exists
	sub, found := manager.GetSubscription(1)
	if !found {
		t.Error("Subscription should exist")
	}
	if sub.chainId != 1 {
		t.Errorf("Expected chainId 1, got %d", sub.chainId)
	}

	// Get status
	status := manager.GetStatus()
	if len(status) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(status))
	}

	// Remove subscription
	err = manager.RemoveSubscription(1)
	if err != nil {
		t.Errorf("Failed to remove subscription: %v", err)
	}

	// Check if subscription is gone
	_, found = manager.GetSubscription(1)
	if found {
		t.Error("Subscription should be removed")
	}
}

func TestConfigStruct(t *testing.T) {
	config := Config{
		ChainId: 42,
		Url:     "ws://example.com:8080",
	}

	if config.ChainId != 42 {
		t.Errorf("Expected ChainId 42, got %d", config.ChainId)
	}

	if config.Url != "ws://example.com:8080" {
		t.Errorf("Expected URL ws://example.com:8080, got %s", config.Url)
	}
}

func TestRootChainInfoStruct(t *testing.T) {
	info := lib.RootChainInfo{
		RootChainId: 1,
		Height:      100,
		Timestamp:   1234567890,
	}

	if info.RootChainId != 1 {
		t.Errorf("Expected RootChainId 1, got %d", info.RootChainId)
	}

	if info.Height != 100 {
		t.Errorf("Expected Height 100, got %d", info.Height)
	}

	if info.Timestamp != 1234567890 {
		t.Errorf("Expected Timestamp 1234567890, got %d", info.Timestamp)
	}
}

package sub

import (
	"fmt"
	"log"
	"time"

	"github.com/canopy-network/canopy/lib"
)

// ExampleUsage demonstrates how to use the root chain subscription package
func ExampleUsage() {
	// Create a logger (you can implement your own Logger interface)
	logger := &DefaultLogger{}

	// Create a subscription manager
	manager := NewManager(logger)

	// Define an event handler that will be called when new root chain info is received
	eventHandler := func(info *lib.RootChainInfo) error {
		fmt.Printf("Received update: ChainId=%d, Height=%d, Timestamp=%d\n",
			info.RootChainId, info.Height, info.Timestamp)

		// Process the received information
		// For example, you might:
		// - Update your local state
		// - Trigger business logic
		// - Store data to database
		// - Send notifications

		return nil
	}

	// Configure root chain connections
	configs := []Config{
		{
			ChainId: 1,
			Url:     "ws://localhost:8081", // Replace with actual root chain URL
		},
		{
			ChainId: 2,
			Url:     "ws://localhost:8082", // Replace with actual root chain URL
		},
	}

	// Add subscriptions for each root chain
	for _, config := range configs {
		err := manager.AddSubscription(config, eventHandler)
		if err != nil {
			log.Printf("Failed to add subscription for chainId=%d: %v", config.ChainId, err)
		}
	}

	// Let the subscriptions run for a while
	time.Sleep(30 * time.Second)

	// Check connection status
	status := manager.GetStatus()
	for chainId, connected := range status {
		fmt.Printf("ChainId %d: Connected=%v\n", chainId, connected)
	}

	// Get latest info for a specific chain
	if info, found := manager.GetLatestInfo(1); found {
		fmt.Printf("Latest info for chain 1: Height=%d\n", info.Height)
	}

	// Stop all subscriptions when done
	err := manager.StopAll()
	if err != nil {
		log.Printf("Error stopping subscriptions: %v", err)
	}
}

// ExampleSingleSubscription demonstrates usage with a single subscription
func ExampleSingleSubscription() {
	logger := &DefaultLogger{}

	// Create configuration
	config := Config{
		ChainId: 1,
		Url:     "ws://localhost:8081",
	}

	// Define event handler
	handler := func(info *lib.RootChainInfo) error {
		fmt.Printf("Chain %d updated to height %d\n", info.RootChainId, info.Height)
		return nil
	}

	// Create subscription
	subscription := NewSubscription(config, handler, logger)

	// Start subscription
	err := subscription.Start()
	if err != nil {
		log.Fatalf("Failed to start subscription: %v", err)
	}

	// Wait for some updates
	time.Sleep(10 * time.Second)

	// Check if connected
	if subscription.IsConnected() {
		fmt.Println("Subscription is connected")

		// Get latest cached info
		info := subscription.GetLatestInfo()
		fmt.Printf("Latest cached info: Height=%d\n", info.Height)
	}

	// Stop subscription
	err = subscription.Stop()
	if err != nil {
		log.Printf("Error stopping subscription: %v", err)
	}
}

package sub

import (
	"sync"

	"github.com/canopy-network/canopy/lib"
)

// Manager handles multiple root chain subscriptions
type Manager struct {
	subscriptions map[uint64]*Subscription // chainId -> subscription
	logger        Logger
	mu            sync.RWMutex
}

// NewManager creates a new subscription manager
func NewManager(logger Logger) *Manager {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return &Manager{
		subscriptions: make(map[uint64]*Subscription),
		logger:        logger,
	}
}

// AddSubscription adds a new root chain subscription
func (m *Manager) AddSubscription(config Config, handler EventHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if subscription already exists
	if _, exists := m.subscriptions[config.ChainId]; exists {
		m.logger.Warnf("Subscription for chainId=%d already exists, replacing...", config.ChainId)
		// Stop existing subscription
		m.subscriptions[config.ChainId].Stop()
	}

	// Create new subscription
	subscription := NewSubscription(config, handler, m.logger)
	m.subscriptions[config.ChainId] = subscription

	// Start the subscription
	return subscription.Start()
}

// RemoveSubscription removes a root chain subscription
func (m *Manager) RemoveSubscription(chainId uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	subscription, exists := m.subscriptions[chainId]
	if !exists {
		return nil // Already removed
	}

	// Stop the subscription
	err := subscription.Stop()
	delete(m.subscriptions, chainId)

	return err
}

// GetSubscription returns a subscription by chain ID
func (m *Manager) GetSubscription(chainId uint64) (*Subscription, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subscription, exists := m.subscriptions[chainId]
	return subscription, exists
}

// GetAllSubscriptions returns all active subscriptions
func (m *Manager) GetAllSubscriptions() map[uint64]*Subscription {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[uint64]*Subscription)
	for chainId, subscription := range m.subscriptions {
		result[chainId] = subscription
	}

	return result
}

// GetConnectedChainIds returns a list of chain IDs with active connections
func (m *Manager) GetConnectedChainIds() []uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var connectedChains []uint64
	for chainId, subscription := range m.subscriptions {
		if subscription.IsConnected() {
			connectedChains = append(connectedChains, chainId)
		}
	}

	return connectedChains
}

// GetLatestInfo returns the latest root chain info for a specific chain ID
func (m *Manager) GetLatestInfo(chainId uint64) (*lib.RootChainInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subscription, exists := m.subscriptions[chainId]
	if !exists {
		return nil, false
	}

	return subscription.GetLatestInfo(), true
}

// StopAll stops all subscriptions
func (m *Manager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for chainId, subscription := range m.subscriptions {
		if err := subscription.Stop(); err != nil {
			m.logger.Errorf("Error stopping subscription for chainId=%d: %s", chainId, err.Error())
			lastErr = err
		}
	}

	// Clear the subscriptions map
	m.subscriptions = make(map[uint64]*Subscription)

	return lastErr
}

// GetStatus returns status information for all subscriptions
func (m *Manager) GetStatus() map[uint64]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[uint64]bool)
	for chainId, subscription := range m.subscriptions {
		status[chainId] = subscription.IsConnected()
	}

	return status
}

package sub

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/canopy-network/canopy/lib"
	"github.com/gorilla/websocket"
)

const (
	// Default WebSocket endpoint path for root chain info subscriptions
	DefaultSubscribeRCInfoPath = "/v1/subscribe-rc-info"
	chainIdParamName           = "chainId"
)

// Subscription manages a WebSocket connection to a root chain for real-time updates
type Subscription struct {
	chainId     uint64               // the chain id of the subscription
	config      Config               // root chain configuration
	conn        *websocket.Conn      // the underlying websocket connection
	info        *lib.RootChainInfo   // cached root chain info from the publisher
	handler     EventHandler         // callback function for processing events
	logger      Logger               // logging interface
	stopCh      chan struct{}        // channel to signal shutdown
	mu          sync.RWMutex         // mutex for thread safety
	isConnected bool                 // connection status
}

// NewSubscription creates a new root chain subscription
func NewSubscription(config Config, handler EventHandler, logger Logger) *Subscription {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return &Subscription{
		chainId: config.ChainId,
		config:  config,
		handler: handler,
		logger:  logger,
		stopCh:  make(chan struct{}),
		info:    &lib.RootChainInfo{},
	}
}

// Start begins the subscription with automatic reconnection
func (s *Subscription) Start() error {
	go s.connectWithBackoff()
	return nil
}

// Stop gracefully shuts down the subscription
func (s *Subscription) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.stopCh)

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.isConnected = false
		return err
	}

	return nil
}

// IsConnected returns the current connection status
func (s *Subscription) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isConnected
}

// GetLatestInfo returns the most recent root chain info
func (s *Subscription) GetLatestInfo() *lib.RootChainInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent race conditions
	info := *s.info
	return &info
}

// connectWithBackoff establishes a websocket connection with exponential backoff retry
func (s *Subscription) connectWithBackoff() {
	// Parse the config URL
	parsedUrl, err := url.Parse(s.config.Url)
	if err != nil {
		s.logger.Fatal(fmt.Sprintf("Failed to parse URL: %s", err.Error()))
		return
	}

	// Get the host
	host := parsedUrl.Host
	if host == "" {
		// Fallback if url didn't have a scheme and was treated as a path
		host = parsedUrl.Path
	}

	// Create WebSocket URL
	wsURL := url.URL{
		Scheme:   "ws",
		Host:     host,
		Path:     DefaultSubscribeRCInfoPath,
		RawQuery: fmt.Sprintf("%s=%d", chainIdParamName, s.chainId),
	}

	// Create retry mechanism with exponential backoff
	retry := NewRetry(uint64(time.Second.Milliseconds()), 25)

	// Keep trying to connect until success or shutdown
	for retry.WaitAndDoRetry() {
		select {
		case <-s.stopCh:
			return
		default:
		}

		s.logger.Infof("Connecting to rootChainId=%d @ %s", s.config.ChainId, wsURL.String())

		// Attempt WebSocket connection
		conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
		if err != nil {
			s.logger.Errorf("Connection failed: %s", err.Error())
			continue
		}

		// Connection successful
		s.mu.Lock()
		s.conn = conn
		s.isConnected = true
		s.mu.Unlock()

		s.logger.Infof("Successfully connected to rootChainId=%d", s.config.ChainId)

		// Start listening for messages
		go s.listen()
		return
	}

	s.logger.Error("Failed to connect after maximum retry attempts")
}

// listen continuously reads messages from the WebSocket connection
func (s *Subscription) listen() {
	defer func() {
		s.mu.Lock()
		if s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
		s.isConnected = false
		s.mu.Unlock()

		// Attempt to reconnect unless we're shutting down
		select {
		case <-s.stopCh:
			return
		default:
			s.logger.Infof("Connection lost, attempting to reconnect...")
			go s.connectWithBackoff()
		}
	}()

	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		// Read message from WebSocket
		_, messageBytes, err := s.conn.ReadMessage()
		if err != nil {
			s.logger.Errorf("Failed to read message: %s", err.Error())
			return
		}
		// Unmarshal protobuf message into RootChainInfo
		newInfo := new(lib.RootChainInfo)
		if err := lib.Unmarshal(messageBytes, newInfo); err != nil {
			s.logger.Errorf("Failed to unmarshal message: %s", err.Error())
			continue
		}

		s.logger.Infof("Received info from RootChainId=%d and Height=%d",
			newInfo.RootChainId, newInfo.Height)

		// Update cached info with thread safety
		s.mu.Lock()
		s.info = newInfo
		s.mu.Unlock()

		// Call the event handler
		if s.handler != nil {
			if err := s.handler(newInfo); err != nil {
				s.logger.Errorf("Event handler error: %s", err.Error())
			}
		}
	}
}

package sessioncleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/enielson/launchpad/internal/repository/interfaces"
)

// Worker handles periodic cleanup of expired session tokens
type Worker struct {
	sessionRepo   interfaces.SessionTokenRepository
	interval      time.Duration
	retentionDays int
	stopChan      chan struct{}
	done          chan struct{}
}

// Config holds configuration for the session cleanup worker
type Config struct {
	// Interval is how often to run the cleanup job (default: 24 hours)
	Interval time.Duration

	// RetentionDays is how long to keep expired tokens before deleting (default: 7 days)
	RetentionDays int
}

// DefaultConfig returns default configuration for the worker
func DefaultConfig() Config {
	return Config{
		Interval:      24 * time.Hour,
		RetentionDays: 7,
	}
}

// NewWorker creates a new session cleanup worker
func NewWorker(sessionRepo interfaces.SessionTokenRepository, config Config) *Worker {
	if config.Interval == 0 {
		config.Interval = 24 * time.Hour
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 7
	}

	return &Worker{
		sessionRepo:   sessionRepo,
		interval:      config.Interval,
		retentionDays: config.RetentionDays,
		stopChan:      make(chan struct{}),
		done:          make(chan struct{}),
	}
}

// Start begins the cleanup worker
func (w *Worker) Start() error {
	log.Printf("Starting session cleanup worker (interval: %v, retention: %d days)", w.interval, w.retentionDays)

	go w.run()

	return nil
}

// Stop gracefully stops the cleanup worker
func (w *Worker) Stop() error {
	log.Println("Stopping session cleanup worker...")
	close(w.stopChan)

	// Wait for worker to finish current operation
	select {
	case <-w.done:
		log.Println("Session cleanup worker stopped")
	case <-time.After(10 * time.Second):
		log.Println("Session cleanup worker stop timeout")
	}

	return nil
}

// run is the main worker loop
func (w *Worker) run() {
	defer close(w.done)

	// Run cleanup immediately on start
	w.cleanup()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.cleanup()
		case <-w.stopChan:
			return
		}
	}
}

// cleanup performs the actual cleanup operation
func (w *Worker) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Running session token cleanup (retention: %d days)...", w.retentionDays)

	deleted, err := w.sessionRepo.DeleteExpiredTokens(ctx, w.retentionDays)
	if err != nil {
		log.Printf("Error during session cleanup: %v", err)
		return
	}

	if deleted > 0 {
		log.Printf("Deleted %d expired session tokens", deleted)
	} else {
		log.Println("No expired session tokens to clean up")
	}
}

// CleanupNow triggers an immediate cleanup (useful for testing)
func (w *Worker) CleanupNow() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	deleted, err := w.sessionRepo.DeleteExpiredTokens(ctx, w.retentionDays)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	log.Printf("Manual cleanup: deleted %d expired session tokens", deleted)
	return nil
}

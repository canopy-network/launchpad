package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/enielson/launchpad/pkg/response"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string]time.Time
	mu       sync.RWMutex
	interval time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified interval
func NewRateLimiter(interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]time.Time),
		interval: interval,
	}

	// Start cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// cleanup removes expired entries from the map every 30 seconds
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, timestamp := range rl.requests {
			if now.Sub(timestamp) > rl.interval {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given identifier should be allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check if identifier exists and is still within rate limit window
	if lastRequest, exists := rl.requests[identifier]; exists {
		if now.Sub(lastRequest) < rl.interval {
			return false
		}
	}

	// Update or add the identifier
	rl.requests[identifier] = now
	return true
}

// TimeUntilAllowed returns the duration until the next request is allowed
func (rl *RateLimiter) TimeUntilAllowed(identifier string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if lastRequest, exists := rl.requests[identifier]; exists {
		elapsed := time.Since(lastRequest)
		if elapsed < rl.interval {
			return rl.interval - elapsed
		}
	}

	return 0
}

// RateLimitMiddleware creates a middleware that rate limits requests
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP address from RemoteAddr (format is "IP:port")
			identifier := r.RemoteAddr
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				identifier = host
			}

			if !limiter.Allow(identifier) {
				waitTime := limiter.TimeUntilAllowed(identifier)
				response.TooManyRequests(w, "Rate limit exceeded. Please try again later.", waitTime)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

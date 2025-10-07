package sub

import (
	"time"

	"github.com/canopy-network/canopy/lib"
)

// Logger interface for customizable logging
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Error(msg string)
	Fatal(msg string)
	Warnf(format string, args ...interface{})
}

// DefaultLogger provides a simple console logger implementation
type DefaultLogger struct{}

func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	// Simple implementation - could be enhanced with proper logging
	println("INFO:", format)
}

func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	println("ERROR:", format)
}

func (l *DefaultLogger) Error(msg string) {
	println("ERROR:", msg)
}

func (l *DefaultLogger) Fatal(msg string) {
	println("FATAL:", msg)
	panic(msg)
}

func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	println("WARN:", format)
}

// Config represents the configuration for a root chain connection
type Config struct {
	ChainId uint64 `json:"chainId"`
	Url     string `json:"url"`
}

// EventHandler is called when new root chain info is received
type EventHandler func(info *lib.RootChainInfo) error

// Retry implements exponential backoff retry logic
type Retry struct {
	waitTimeMS uint64 // time to wait in milliseconds
	maxLoops   uint64 // the maximum number of loops before quitting
	loopCount  uint64 // the loop count itself
}

// NewRetry constructs a new Retry given parameters
func NewRetry(waitTimeMS, maxLoops uint64) *Retry {
	return &Retry{
		waitTimeMS: waitTimeMS,
		maxLoops:   maxLoops,
	}
}

// WaitAndDoRetry sleeps the appropriate time and returns false if maxed out retry
func (r *Retry) WaitAndDoRetry() bool {
	// if GTE max loops
	if r.maxLoops <= r.loopCount {
		// exit with 'try again'
		return false
	}
	// don't sleep or increment on the first iteration
	if r.loopCount != 0 {
		// sleep the allotted time
		time.Sleep(time.Duration(r.waitTimeMS) * time.Millisecond)
		// double the timeout
		r.waitTimeMS = r.waitTimeMS * 2
	}
	// increment the loop count
	r.loopCount++
	// exit with success
	return true
}

package newblock

import (
	"log"

	"github.com/enielson/launchpad/pkg/sub"
)

// loggerAdapter adapts Go's standard logger to the sub.Logger interface
type loggerAdapter struct{}

// NewLogger creates a new logger adapter for the subscription package
func NewLogger() sub.Logger {
	return &loggerAdapter{}
}

func (l *loggerAdapter) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *loggerAdapter) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func (l *loggerAdapter) Error(msg string) {
	log.Printf("[ERROR] %s", msg)
}

func (l *loggerAdapter) Fatal(msg string) {
	log.Fatalf("[FATAL] %s", msg)
}

func (l *loggerAdapter) Warnf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

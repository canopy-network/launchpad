package middleware

import (
	"context"
	"net/http"
)

// MockAuthMiddleware extracts the X-User-ID header and adds it to the context
// This is for development/testing purposes only
func MockAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract X-User-ID header
			userID := r.Header.Get("X-User-ID")

			// Add user ID to context if present
			if userID != "" {
				ctx := context.WithValue(r.Context(), "userID", userID)
				r = r.WithContext(ctx)
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

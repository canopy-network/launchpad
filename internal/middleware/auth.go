package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/pkg/response"
)

// AuthMiddleware creates middleware that validates session tokens
func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Missing authorization header")
				return
			}

			// Check for Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.Unauthorized(w, "Invalid authorization format. Use 'Bearer <token>'")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				response.Unauthorized(w, "Missing authorization token")
				return
			}

			// Validate token and get user
			user, session, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				// Handle specific error cases
				switch err {
				case services.ErrInvalidToken:
					response.Unauthorized(w, "Invalid session token")
				case services.ErrTokenExpired:
					response.Unauthorized(w, "Session token has expired")
				case services.ErrTokenRevoked:
					response.Unauthorized(w, "Session token has been revoked")
				case services.ErrSessionInvalid:
					response.Unauthorized(w, "Session is no longer valid")
				default:
					response.Unauthorized(w, "Authentication failed")
				}
				return
			}

			// Add user and session to context
			ctx := context.WithValue(r.Context(), "userID", user.ID.String())
			ctx = context.WithValue(ctx, "user", user)
			ctx = context.WithValue(ctx, "session", session)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware attempts to authenticate but doesn't fail if no token is provided
// Useful for endpoints that work for both authenticated and unauthenticated users
func OptionalAuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Check for Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				// Invalid format, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				// Empty token, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Validate token and get user
			user, session, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				// Token is invalid, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Add user and session to context
			ctx := context.WithValue(r.Context(), "userID", user.ID.String())
			ctx = context.WithValue(ctx, "user", user)
			ctx = context.WithValue(ctx, "session", session)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

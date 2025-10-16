package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *services.AuthService
	validator   *validators.Validator
}

func NewAuthHandler(authService *services.AuthService, validator *validators.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

// SendEmailCode handles POST /api/v1/auth/email
func (h *AuthHandler) SendEmailCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.EmailAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		validationErrors := h.validator.FormatErrors(err)
		response.ValidationError(w, validationErrors)
		return
	}

	// Generate and send verification code
	code, err := h.authService.SendEmailCode(ctx, req.Email)
	if err != nil {
		log.Printf("Failed to send email code to %s: %v", req.Email, err)
		response.InternalServerError(w, "Failed to send verification code")
		return
	}

	// In development, we might want to return the code for testing
	// In production, this should NOT be returned
	response.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Verification code sent successfully",
		"email":   req.Email,
		// Remove this in production:
		"code": code,
	})
}

// VerifyEmailCode handles POST /api/v1/auth/verify
func (h *AuthHandler) VerifyEmailCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.VerifyEmailCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		validationErrors := h.validator.FormatErrors(err)
		response.ValidationError(w, validationErrors)
		return
	}

	// Verify code
	if err := h.authService.VerifyCode(req.Email, req.Code); err != nil {
		if err == services.ErrInvalidCode {
			response.BadRequest(w, "Invalid verification code", nil)
			return
		}
		if err == services.ErrCodeExpired {
			response.BadRequest(w, "Verification code has expired", nil)
			return
		}
		log.Printf("Failed to verify code for %s: %v", req.Email, err)
		response.InternalServerError(w, "Failed to verify code")
		return
	}

	// Get user agent and IP address for session tracking
	userAgent := r.Header.Get("User-Agent")
	ipAddress := r.RemoteAddr
	// Strip port from IP address
	if colonIdx := len(ipAddress) - 1; colonIdx > 0 {
		for i := colonIdx; i >= 0; i-- {
			if ipAddress[i] == ':' {
				ipAddress = ipAddress[:i]
				break
			}
		}
	}

	// Complete login flow: create/get user, mark verified, create session
	loginResponse, err := h.authService.CompleteEmailLogin(ctx, req.Email, userAgent, ipAddress)
	if err != nil {
		log.Printf("Failed to complete login for %s: %v", req.Email, err)
		response.InternalServerError(w, "Failed to complete login")
		return
	}

	response.Success(w, http.StatusOK, loginResponse)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from Authorization header
	token := extractTokenFromHeader(r)
	if token == "" {
		response.BadRequest(w, "Missing authorization token", nil)
		return
	}

	// Revoke the token
	if err := h.authService.RevokeSession(ctx, token); err != nil {
		log.Printf("Failed to revoke session: %v", err)
		response.InternalServerError(w, "Failed to logout")
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

// GetSessions handles GET /api/v1/auth/sessions
func (h *AuthHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID", nil)
		return
	}

	// Get current session token to mark it in the response
	currentToken := extractTokenFromHeader(r)
	var currentSessionID string
	if currentToken != "" {
		_, currentSession, err := h.authService.ValidateToken(ctx, currentToken)
		if err == nil && currentSession != nil {
			currentSessionID = currentSession.ID.String()
		}
	}

	// Get all active sessions
	sessions, err := h.authService.GetActiveSessions(ctx, userUUID)
	if err != nil {
		log.Printf("Failed to get sessions for user %s: %v", userID, err)
		response.InternalServerError(w, "Failed to retrieve sessions")
		return
	}

	// Convert to SessionInfo for display
	sessionInfos := make([]models.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		isCurrent := session.ID.String() == currentSessionID
		sessionInfos = append(sessionInfos, *session.ToSessionInfo(isCurrent))
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"sessions": sessionInfos,
		"total":    len(sessionInfos),
	})
}

// RevokeSession handles DELETE /api/v1/auth/sessions/{id}
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session ID from URL
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid session ID", nil)
		return
	}

	// Get user ID from context
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID", nil)
		return
	}

	// Verify the session belongs to the user (security check)
	sessions, err := h.authService.GetActiveSessions(ctx, userUUID)
	if err != nil {
		log.Printf("Failed to get sessions: %v", err)
		response.InternalServerError(w, "Failed to revoke session")
		return
	}

	found := false
	for _, session := range sessions {
		if session.ID == sessionID {
			found = true
			break
		}
	}

	if !found {
		response.NotFound(w, "Session not found or already revoked")
		return
	}

	// Revoke the session
	if err := h.authService.RevokeSessionByID(ctx, sessionID); err != nil {
		log.Printf("Failed to revoke session %s: %v", sessionID, err)
		response.InternalServerError(w, "Failed to revoke session")
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Session revoked successfully",
	})
}

// Helper function to extract token from Authorization header
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Bearer token format: "Bearer <token>"
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}

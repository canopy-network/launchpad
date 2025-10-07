package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
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

	// Code verified successfully
	// In a real application, you would:
	// 1. Create or update user record
	// 2. Generate JWT token
	// 3. Return auth token to client

	response.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully",
		"email":   req.Email,
	})
}

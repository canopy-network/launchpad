package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *services.UserService
	validator   *validators.Validator
}

func NewUserHandler(userService *services.UserService, validator *validators.Validator) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
	}
}

// UpdateProfile handles PUT /api/v1/users/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateProfileRequest
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

	// Update profile
	user, err := h.userService.UpdateProfile(ctx, userUUID, &req)
	if err != nil {
		if err.Error() == "username already taken" {
			response.Conflict(w, "Username is already taken", nil)
			return
		}
		if err.Error() == "no fields to update" {
			response.BadRequest(w, "No fields provided to update", nil)
			return
		}
		if err.Error() == "user not found" {
			response.NotFound(w, "User not found")
			return
		}
		log.Printf("Failed to update profile for user %s: %v", userID, err)
		response.InternalServerError(w, "Failed to update profile")
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"user":    user,
		"message": "Profile updated successfully",
	})
}

package services

import (
	"context"
	"fmt"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo interfaces.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// UpdateProfile updates user profile information (partial update)
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	// Validate that at least one field is provided
	hasUpdate := req.Username != nil || req.DisplayName != nil || req.Bio != nil ||
		req.AvatarURL != nil || req.WebsiteURL != nil || req.TwitterHandle != nil ||
		req.GithubUsername != nil || req.TelegramHandle != nil

	if !hasUpdate {
		return nil, fmt.Errorf("no fields to update")
	}

	// Perform the update
	user, err := s.userRepo.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

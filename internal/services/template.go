package services

import (
	"context"
	"errors"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
)

type TemplateService struct {
	templateRepo interfaces.ChainTemplateRepository
}

func NewTemplateService(templateRepo interfaces.ChainTemplateRepository) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
	}
}

// GetTemplates retrieves all templates with optional filtering
func (s *TemplateService) GetTemplates(ctx context.Context, filters interfaces.TemplateFilters, pagination interfaces.Pagination) ([]models.ChainTemplate, int, error) {
	return s.templateRepo.List(ctx, filters, pagination)
}

// GetTemplateByID retrieves a template by ID
func (s *TemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.ChainTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

// GetActiveTemplates retrieves all active templates
func (s *TemplateService) GetActiveTemplates(ctx context.Context) ([]models.ChainTemplate, error) {
	return s.templateRepo.GetActive(ctx)
}

// GetTemplatesByCategory retrieves templates by category
func (s *TemplateService) GetTemplatesByCategory(ctx context.Context, category string) ([]models.ChainTemplate, error) {
	return s.templateRepo.GetByCategory(ctx, category)
}

package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
)

type TemplateHandler struct {
	templateService *services.TemplateService
	validator       *validators.Validator
}

func NewTemplateHandler(templateService *services.TemplateService, validator *validators.Validator) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		validator:       validator,
	}
}

// GetTemplates handles GET /api/v1/templates
func (h *TemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var params models.TemplatesQueryParams
	if err := h.parseQueryParams(r, &params); err != nil {
		response.BadRequest(w, "Invalid query parameters", err.Error())
		return
	}

	// Validate query parameters
	if err := h.validator.Validate(&params); err != nil {
		validationErrors := h.validator.FormatErrors(err)
		response.ValidationError(w, validationErrors)
		return
	}

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 20
	}

	// Convert to repository filters and pagination
	filters := interfaces.TemplateFilters{
		Category:        params.Category,
		ComplexityLevel: params.ComplexityLevel,
		IsActive:        params.IsActive,
	}

	pagination := interfaces.Pagination{
		Page:   params.Page,
		Limit:  params.Limit,
		Offset: (params.Page - 1) * params.Limit,
	}

	// Get templates
	templates, total, err := h.templateService.GetTemplates(ctx, filters, pagination)
	if err != nil {
		log.Printf("Failed to retrieve templates: %v", err)
		response.InternalServerError(w, "Failed to retrieve templates")
		return
	}

	// Build pagination response
	paginationResponse := &models.Pagination{
		Page:  params.Page,
		Limit: params.Limit,
		Total: total,
		Pages: (total + params.Limit - 1) / params.Limit,
	}

	response.SuccessWithPagination(w, http.StatusOK, templates, paginationResponse)
}

// Helper methods
func (h *TemplateHandler) parseQueryParams(r *http.Request, params interface{}) error {
	switch p := params.(type) {
	case *models.TemplatesQueryParams:
		p.Category = r.URL.Query().Get("category")
		p.ComplexityLevel = r.URL.Query().Get("complexity_level")

		if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
			if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
				p.IsActive = &isActive
			}
		}

		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil {
				p.Page = page
			}
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				p.Limit = limit
			}
		}
	}

	return nil
}

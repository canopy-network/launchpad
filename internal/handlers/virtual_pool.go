package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
)

type VirtualPoolHandler struct {
	virtualPoolService *services.VirtualPoolService
	validator          *validators.Validator
}

func NewVirtualPoolHandler(virtualPoolService *services.VirtualPoolService, validator *validators.Validator) *VirtualPoolHandler {
	return &VirtualPoolHandler{
		virtualPoolService: virtualPoolService,
		validator:          validator,
	}
}

// GetVirtualPools handles GET /api/v1/virtual-pools
func (h *VirtualPoolHandler) GetVirtualPools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var params models.VirtualPoolsQueryParams
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

	// Get virtual pools
	pools, pagination, err := h.virtualPoolService.GetAllPools(ctx, params.Page, params.Limit)
	if err != nil {
		log.Printf("Failed to retrieve virtual pools: %v", err)
		response.InternalServerError(w, "Failed to retrieve virtual pools")
		return
	}

	response.SuccessWithPagination(w, http.StatusOK, pools, pagination)
}

// Helper methods
func (h *VirtualPoolHandler) parseQueryParams(r *http.Request, params *models.VirtualPoolsQueryParams) error {
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	return nil
}

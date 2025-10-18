package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
	"github.com/go-chi/chi/v5"
)

type ChainHandler struct {
	chainService *services.ChainService
	validator    *validators.Validator
}

func NewChainHandler(chainService *services.ChainService, validator *validators.Validator) *ChainHandler {
	return &ChainHandler{
		chainService: chainService,
		validator:    validator,
	}
}

// GetChains handles GET /api/v1/chains
func (h *ChainHandler) GetChains(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var params models.ChainsQueryParams
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

	// Parse include relations
	include := []string{}
	if params.Include != "" {
		include = []string{params.Include}
	}

	// Get chains
	chains, pagination, err := h.chainService.GetChains(
		ctx,
		params.Status,
		params.CreatedBy,
		include,
		params.Page,
		params.Limit,
	)
	if err != nil {
		log.Printf("Failed to retrieve chains: %v", err)
		response.InternalServerError(w, "Failed to retrieve chains")
		return
	}

	response.SuccessWithPagination(w, http.StatusOK, chains, pagination)
}

// GetChain handles GET /api/v1/chains/{id}
func (h *ChainHandler) GetChain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	include := r.URL.Query().Get("include")

	chain, err := h.chainService.GetChainByID(ctx, chainID, include)
	if err != nil {
		if err == services.ErrChainNotFound {
			response.NotFound(w, "Chain not found")
			return
		}
		log.Printf("Failed to retrieve chain %s: %v", chainID, err)
		response.InternalServerError(w, "Failed to retrieve chain")
		return
	}

	response.Success(w, http.StatusOK, chain)
}

// CreateChain handles POST /api/v1/chains
func (h *ChainHandler) CreateChain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := h.getUserIDFromContext(ctx)

	var req models.CreateChainRequest
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

	// Create chain
	chain, err := h.chainService.CreateChain(ctx, &req, userID)
	if err != nil {
		if err == services.ErrChainAlreadyExists {
			log.Printf("Create chain failed - chain already exists: %v", err)
			response.Conflict(w, "Chain name already exists", nil)
			return
		}
		log.Printf("Create chain failed for user %s: %v", userID, err)
		response.InternalServerError(w, "Failed to create chain")
		return
	}

	response.Success(w, http.StatusCreated, chain)
}

// DeleteChain handles DELETE /api/v1/chains/{id}
func (h *ChainHandler) DeleteChain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	userID := h.getUserIDFromContext(ctx)

	err := h.chainService.DeleteChain(ctx, chainID, userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Chain deleted successfully",
	})
}

// UpdateChainDescription handles PUT /api/v1/chains/{id}/description
func (h *ChainHandler) UpdateChainDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	userID := h.getUserIDFromContext(ctx)

	var req models.UpdateChainDescriptionRequest
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

	// Update chain description
	chain, err := h.chainService.UpdateChainDescription(ctx, chainID, userID, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, chain)
}

// GetRepository handles GET /api/v1/chains/{id}/repository
func (h *ChainHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	userID := h.getUserIDFromContext(ctx)

	repo, err := h.chainService.GetRepositoryByChainID(ctx, chainID, userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, repo)
}

// UpdateRepository handles PUT /api/v1/chains/{id}/repository
func (h *ChainHandler) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	userID := h.getUserIDFromContext(ctx)

	var req models.UpdateChainRepositoryRequest
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

	// Update repository
	repo, err := h.chainService.UpdateRepositoryByChainID(ctx, chainID, userID, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, repo)
}

// GetTransactions handles GET /api/v1/chains/{id}/transactions
func (h *ChainHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")

	// Parse query parameters
	var params models.TransactionsQueryParams
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

	// Get transactions
	transactions, pagination, err := h.chainService.GetTransactions(
		ctx,
		chainID,
		params.UserID,
		params.TransactionType,
		params.Page,
		params.Limit,
	)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.SuccessWithPagination(w, http.StatusOK, transactions, pagination)
}

// GetPriceHistory handles GET /api/v1/chains/{id}/price-history
func (h *ChainHandler) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")

	// Parse query parameters
	var params models.PriceHistoryQueryParams
	if err := h.parsePriceHistoryParams(r, &params); err != nil {
		response.BadRequest(w, "Invalid query parameters", err.Error())
		return
	}

	// Get price history from service
	candles, err := h.chainService.GetPriceHistory(ctx, chainID, params.StartTime, params.EndTime)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, candles)
}

// GetAssets handles GET /api/v1/chains/{id}/assets
func (h *ChainHandler) GetAssets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	userID := h.getUserIDFromContext(ctx)

	assets, err := h.chainService.GetAssets(ctx, chainID, userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, assets)
}

// CreateAsset handles POST /api/v1/chains/{id}/assets
func (h *ChainHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	userID := h.getUserIDFromContext(ctx)

	var req models.CreateChainAssetRequest
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

	// Create asset
	asset, err := h.chainService.CreateAsset(ctx, chainID, userID, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, asset)
}

// UpdateAsset handles PUT /api/v1/chains/{id}/assets/{asset_id}
func (h *ChainHandler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chainID := chi.URLParam(r, "id")
	assetID := chi.URLParam(r, "asset_id")
	userID := h.getUserIDFromContext(ctx)

	var req models.UpdateChainAssetRequest
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

	// Update asset
	asset, err := h.chainService.UpdateAsset(ctx, chainID, assetID, userID, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Success(w, http.StatusOK, asset)
}

// Helper methods
func (h *ChainHandler) getUserIDFromContext(ctx context.Context) string {
	// This would be set by the auth middleware
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return ""
	}
	return userID
}

func (h *ChainHandler) parseQueryParams(r *http.Request, params interface{}) error {
	// This would typically use a library like gorilla/schema or similar
	// For now, we'll parse manually based on the request type
	switch p := params.(type) {
	case *models.ChainsQueryParams:
		p.Status = r.URL.Query().Get("status")
		p.CreatedBy = r.URL.Query().Get("created_by")
		p.TemplateID = r.URL.Query().Get("template_id")
		p.Include = r.URL.Query().Get("include")

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

	case *models.TransactionsQueryParams:
		p.UserID = r.URL.Query().Get("user_id")
		p.TransactionType = r.URL.Query().Get("transaction_type")

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

func (h *ChainHandler) parsePriceHistoryParams(r *http.Request, params *models.PriceHistoryQueryParams) error {
	// Parse start_time if provided
	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := parseTime(startTimeStr)
		if err != nil {
			return err
		}
		params.StartTime = &startTime
	}

	// Parse end_time if provided
	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := parseTime(endTimeStr)
		if err != nil {
			return err
		}
		params.EndTime = &endTime
	}

	return nil
}

// parseTime parses RFC3339/ISO 8601 time strings
func parseTime(timeStr string) (time.Time, error) {
	// Try RFC3339 format first (standard format)
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}

	// Try RFC3339Nano as fallback
	t, err = time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid time format, expected RFC3339/ISO 8601 (e.g., 2006-01-02T15:04:05Z)")
}

func (h *ChainHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch err {
	case services.ErrChainNotFound:
		response.NotFound(w, "Chain not found")
	case services.ErrChainAlreadyExists:
		response.Conflict(w, "Chain already exists", nil)
	case services.ErrChainNotInDraftStatus:
		response.UnprocessableEntity(w, "Chain is not in draft status", nil)
	case services.ErrUnauthorized:
		response.Forbidden(w, "Access denied")
	case services.ErrRepositoryNotFound:
		response.NotFound(w, "Repository not found")
	case services.ErrAssetNotFound:
		response.NotFound(w, "Asset not found")
	default:
		log.Printf("Unhandled service error: %v", err)
		response.InternalServerError(w, "Internal server error")
	}
}

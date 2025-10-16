package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/enielson/launchpad/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletService *services.WalletService
	validator     *validators.Validator
}

func NewWalletHandler(walletService *services.WalletService, validator *validators.Validator) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		validator:     validator,
	}
}

// CreateWallet handles POST /api/v1/wallets
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user ID from context
	userIDStr, ok := ctx.Value("userID").(string)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	authenticatedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid user ID", err.Error())
		return
	}

	var req models.CreateWalletRequest
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

	// Parse optional UUIDs
	var userID, chainID *uuid.UUID

	if req.UserID != nil {
		parsedUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			response.BadRequest(w, "Invalid user_id format", err.Error())
			return
		}
		userID = &parsedUserID
	}

	if req.ChainID != nil {
		parsedChainID, err := uuid.Parse(*req.ChainID)
		if err != nil {
			response.BadRequest(w, "Invalid chain_id format", err.Error())
			return
		}
		chainID = &parsedChainID
	}

	// Create wallet using authenticated user as creator
	wallet, err := h.walletService.CreateWallet(
		ctx,
		req.Password,
		userID,
		chainID,
		req.WalletName,
		req.WalletDescription,
		&authenticatedUserID,
	)

	if err != nil {
		log.Printf("Failed to create wallet: %v", err)
		response.InternalServerError(w, "Failed to create wallet")
		return
	}

	response.Success(w, http.StatusCreated, wallet)
}

// GetWallets handles GET /api/v1/wallets
func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var queryParams models.WalletsQueryParams
	if err := h.parseQueryParams(r, &queryParams); err != nil {
		response.BadRequest(w, "Invalid query parameters", err.Error())
		return
	}

	// Validate query parameters
	if err := h.validator.Validate(&queryParams); err != nil {
		validationErrors := h.validator.FormatErrors(err)
		response.ValidationError(w, validationErrors)
		return
	}

	// Set defaults
	if queryParams.Page == 0 {
		queryParams.Page = 1
	}
	if queryParams.Limit == 0 {
		queryParams.Limit = 20
	}

	// Build filters
	filters := interfaces.WalletFilters{
		IsActive: queryParams.IsActive,
		IsLocked: queryParams.IsLocked,
	}

	if queryParams.UserID != "" {
		userID, err := uuid.Parse(queryParams.UserID)
		if err != nil {
			response.BadRequest(w, "Invalid user_id format", err.Error())
			return
		}
		filters.UserID = &userID
	}

	if queryParams.ChainID != "" {
		chainID, err := uuid.Parse(queryParams.ChainID)
		if err != nil {
			response.BadRequest(w, "Invalid chain_id format", err.Error())
			return
		}
		filters.ChainID = &chainID
	}

	if queryParams.CreatedBy != "" {
		createdBy, err := uuid.Parse(queryParams.CreatedBy)
		if err != nil {
			response.BadRequest(w, "Invalid created_by format", err.Error())
			return
		}
		filters.CreatedBy = &createdBy
	}

	// Set up pagination
	pagination := interfaces.Pagination{
		Limit:  queryParams.Limit,
		Offset: (queryParams.Page - 1) * queryParams.Limit,
	}

	// Get wallets
	wallets, total, err := h.walletService.ListWallets(ctx, filters, pagination)
	if err != nil {
		log.Printf("Failed to get wallets: %v", err)
		response.InternalServerError(w, "Failed to retrieve wallets")
		return
	}

	// Calculate total pages
	totalPages := (total + queryParams.Limit - 1) / queryParams.Limit

	paginationResponse := &models.Pagination{
		Page:  queryParams.Page,
		Limit: queryParams.Limit,
		Total: total,
		Pages: totalPages,
	}

	response.SuccessWithPagination(w, http.StatusOK, wallets, paginationResponse)
}

// GetWallet handles GET /api/v1/wallets/{id}
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid wallet ID", err.Error())
		return
	}

	wallet, err := h.walletService.GetWallet(ctx, walletID)
	if err != nil {
		if err == services.ErrWalletNotFound {
			response.NotFound(w, "Wallet not found")
			return
		}
		log.Printf("Failed to get wallet: %v", err)
		response.InternalServerError(w, "Failed to retrieve wallet")
		return
	}

	response.Success(w, http.StatusOK, wallet)
}

// UpdateWallet handles PUT /api/v1/wallets/{id}
func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid wallet ID", err.Error())
		return
	}

	var req models.UpdateWalletRequest
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

	// Update wallet
	wallet, err := h.walletService.UpdateWallet(ctx, walletID, req.WalletName, req.WalletDescription, req.IsActive)
	if err != nil {
		if err == services.ErrWalletNotFound {
			response.NotFound(w, "Wallet not found")
			return
		}
		log.Printf("Failed to update wallet: %v", err)
		response.InternalServerError(w, "Failed to update wallet")
		return
	}

	response.Success(w, http.StatusOK, wallet)
}

// DeleteWallet handles DELETE /api/v1/wallets/{id}
func (h *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid wallet ID", err.Error())
		return
	}

	err = h.walletService.DeleteWallet(ctx, walletID)
	if err != nil {
		if err == services.ErrWalletNotFound {
			response.NotFound(w, "Wallet not found")
			return
		}
		log.Printf("Failed to delete wallet: %v", err)
		response.InternalServerError(w, "Failed to delete wallet")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Wallet deleted successfully",
	})
}

// DecryptWallet handles POST /api/v1/wallets/{id}/decrypt
func (h *WalletHandler) DecryptWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid wallet ID", err.Error())
		return
	}

	var req models.DecryptWalletRequest
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

	// Decrypt wallet
	walletWithKey, err := h.walletService.DecryptWallet(ctx, walletID, req.Password)
	if err != nil {
		if err == services.ErrWalletNotFound {
			response.NotFound(w, "Wallet not found")
			return
		}
		if err == services.ErrWalletLocked {
			response.Forbidden(w, "Wallet is locked due to too many failed attempts")
			return
		}
		if err == services.ErrInvalidPassword {
			response.Unauthorized(w, "Invalid password")
			return
		}
		log.Printf("Failed to decrypt wallet: %v", err)
		response.InternalServerError(w, "Failed to decrypt wallet")
		return
	}

	response.Success(w, http.StatusOK, walletWithKey)
}

// UnlockWallet handles POST /api/v1/wallets/{id}/unlock
func (h *WalletHandler) UnlockWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid wallet ID", err.Error())
		return
	}

	err = h.walletService.UnlockWallet(ctx, walletID)
	if err != nil {
		if err == services.ErrWalletNotFound {
			response.NotFound(w, "Wallet not found")
			return
		}
		log.Printf("Failed to unlock wallet: %v", err)
		response.InternalServerError(w, "Failed to unlock wallet")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{
		"message": "Wallet unlocked successfully",
	})
}

// Helper methods
func (h *WalletHandler) parseQueryParams(r *http.Request, params *models.WalletsQueryParams) error {
	params.UserID = r.URL.Query().Get("user_id")
	params.ChainID = r.URL.Query().Get("chain_id")
	params.CreatedBy = r.URL.Query().Get("created_by")

	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		params.IsActive = &isActive
	}

	if isLockedStr := r.URL.Query().Get("is_locked"); isLockedStr != "" {
		isLocked := isLockedStr == "true"
		params.IsLocked = &isLocked
	}

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

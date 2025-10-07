package handlers

import (
	"net/http"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/pkg/response"
)

// HealthCheck handles GET /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthResponse := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0", // This could come from build info
	}

	response.Success(w, http.StatusOK, healthResponse)
}

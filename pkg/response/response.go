package response

import (
	"encoding/json"
	"net/http"

	"github.com/enielson/launchpad/internal/models"
)

// Success writes a successful API response
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.APIResponse{
		Data: data,
	}

	json.NewEncoder(w).Encode(response)
}

// SuccessWithPagination writes a successful API response with pagination
func SuccessWithPagination(w http.ResponseWriter, statusCode int, data interface{}, pagination *models.Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.APIResponse{
		Data:       data,
		Pagination: pagination,
	}

	json.NewEncoder(w).Encode(response)
}

// Error writes an error API response
func Error(w http.ResponseWriter, statusCode int, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.APIError{
		Error: &models.ErrorDetails{
			Code:    getErrorCode(statusCode),
			Message: message,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// ValidationError writes a validation error response
func ValidationError(w http.ResponseWriter, errors []models.ValidationErrorDetail) {
	Error(w, http.StatusBadRequest, "Request validation failed", errors)
}

// BadRequest writes a 400 bad request response
func BadRequest(w http.ResponseWriter, message string, details interface{}) {
	Error(w, http.StatusBadRequest, message, details)
}

// Unauthorized writes a 401 unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, nil)
}

// Forbidden writes a 403 forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, nil)
}

// NotFound writes a 404 not found response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, nil)
}

// Conflict writes a 409 conflict response
func Conflict(w http.ResponseWriter, message string, details interface{}) {
	Error(w, http.StatusConflict, message, details)
}

// UnprocessableEntity writes a 422 unprocessable entity response
func UnprocessableEntity(w http.ResponseWriter, message string, details interface{}) {
	Error(w, http.StatusUnprocessableEntity, message, details)
}

// InternalServerError writes a 500 internal server error response
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message, nil)
}

// TooManyRequests writes a 429 too many requests response
func TooManyRequests(w http.ResponseWriter, message string, retryAfter interface{}) {
	Error(w, http.StatusTooManyRequests, message, retryAfter)
}

// getErrorCode maps HTTP status codes to error codes
func getErrorCode(statusCode int) string {
	codes := map[int]string{
		http.StatusBadRequest:          models.ErrorCodeBadRequest,
		http.StatusUnauthorized:        models.ErrorCodeUnauthorized,
		http.StatusForbidden:           models.ErrorCodeForbidden,
		http.StatusNotFound:            models.ErrorCodeNotFound,
		http.StatusConflict:            models.ErrorCodeConflict,
		http.StatusUnprocessableEntity: models.ErrorCodeUnprocessable,
		http.StatusTooManyRequests:     models.ErrorCodeRateLimitExceeded,
		http.StatusInternalServerError: models.ErrorCodeInternalError,
	}

	if code, exists := codes[statusCode]; exists {
		return code
	}
	return models.ErrorCodeInternalError
}

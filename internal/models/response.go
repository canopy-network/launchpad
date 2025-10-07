package models

// APIResponse represents the standard API response wrapper
type APIResponse struct {
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// APIError represents the standard error response
type APIError struct {
	Error *ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// ValidationErrorDetail represents validation error details
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrorCodeValidation        = "VALIDATION_ERROR"
	ErrorCodeUnauthorized      = "UNAUTHORIZED"
	ErrorCodeForbidden         = "FORBIDDEN"
	ErrorCodeNotFound          = "NOT_FOUND"
	ErrorCodeConflict          = "CONFLICT"
	ErrorCodeBusinessRule      = "BUSINESS_RULE_VIOLATION"
	ErrorCodeInternalError     = "INTERNAL_ERROR"
	ErrorCodeBadRequest        = "BAD_REQUEST"
	ErrorCodeUnprocessable     = "UNPROCESSABLE_ENTITY"
	ErrorCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

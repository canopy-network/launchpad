package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// EndpointCategory represents a category of API endpoints
type EndpointCategory string

const (
	CategoryHealth      EndpointCategory = "Health & Routes"
	CategoryAuth        EndpointCategory = "Authentication"
	CategoryTemplates   EndpointCategory = "Templates"
	CategoryChains      EndpointCategory = "Chains"
	CategoryVirtualPool EndpointCategory = "Virtual Pools"
)

// HTTPMethod represents an HTTP method
type HTTPMethod string

const (
	MethodGET    HTTPMethod = "GET"
	MethodPOST   HTTPMethod = "POST"
	MethodPUT    HTTPMethod = "PUT"
	MethodDELETE HTTPMethod = "DELETE"
)

// PathParam represents a path parameter with optional initial value function
type PathParam struct {
	Name         string
	InitialValue func(m *Model) string // Optional function to get initial value
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Name        string
	Method      HTTPMethod
	Path        string
	Category    EndpointCategory
	Description string
	// Parameters that need to be filled in the path (e.g., {id})
	PathParams []PathParam
	// Example request body for POST/PUT
	ExampleBody string
	// Query parameters
	QueryParams []QueryParam
}

// QueryParam represents a query parameter
type QueryParam struct {
	Name        string
	Description string
	Example     string
	Required    bool
}

// RequestResult represents the result of an HTTP request
type RequestResult struct {
	StatusCode   int
	Status       string
	Body         string
	Headers      http.Header
	Duration     time.Duration
	Error        error
	RequestTime  time.Time
	RequestURL   string
	RequestBody  string
	RequestUserID string
	Method       HTTPMethod
	EndpointName string
}

// GetAllEndpoints returns all available endpoints
func GetAllEndpoints() []Endpoint {
	return []Endpoint{
		// Health & Routes
		{
			Name:        "Health Check",
			Method:      MethodGET,
			Path:        "/health",
			Category:    CategoryHealth,
			Description: "Check if the API is running",
		},
		{
			Name:        "List Routes",
			Method:      MethodGET,
			Path:        "/api/v1/routes",
			Category:    CategoryHealth,
			Description: "List all registered API routes",
		},

		// Authentication
		{
			Name:        "Send Email Code",
			Method:      MethodPOST,
			Path:        "/api/v1/auth/email",
			Category:    CategoryAuth,
			Description: "Request email verification code",
			ExampleBody: `{
  "email": "ericnielson@fastmail.mx"
}`,
		},
		{
			Name:        "Verify Email Code",
			Method:      MethodPOST,
			Path:        "/api/v1/auth/verify",
			Category:    CategoryAuth,
			Description: "Verify email with code",
			ExampleBody: `{
  "email": "ericnielson@fastmail.mx",
  "code": "123456"
}`,
		},

		// Templates
		{
			Name:        "Get Templates",
			Method:      MethodGET,
			Path:        "/api/v1/templates",
			Category:    CategoryTemplates,
			Description: "List all chain templates",
			QueryParams: []QueryParam{
				{Name: "page", Description: "Page number", Example: "1"},
				{Name: "limit", Description: "Items per page", Example: "10"},
				{Name: "category", Description: "Filter by category", Example: "defi"},
				{Name: "complexity_level", Description: "Filter by complexity", Example: "beginner"},
			},
		},

		// Virtual Pools - List All
		{
			Name:        "Get All Virtual Pools",
			Method:      MethodGET,
			Path:        "/api/v1/virtual-pools",
			Category:    CategoryVirtualPool,
			Description: "List all virtual pools with pagination",
			QueryParams: []QueryParam{
				{Name: "page", Description: "Page number", Example: "1"},
				{Name: "limit", Description: "Items per page (max 100)", Example: "20"},
			},
		},

		// Chains
		{
			Name:        "Get Chains",
			Method:      MethodGET,
			Path:        "/api/v1/chains",
			Category:    CategoryChains,
			Description: "List all chains",
			QueryParams: []QueryParam{
				{Name: "page", Description: "Page number", Example: "1"},
				{Name: "limit", Description: "Items per page", Example: "10"},
				{Name: "status", Description: "Filter by status", Example: "draft"},
			},
		},
		{
			Name:        "Create Chain",
			Method:      MethodPOST,
			Path:        "/api/v1/chains",
			Category:    CategoryChains,
			Description: "Create a new chain",
			ExampleBody: `{
  "chain_name": "My Test Chain",
  "token_symbol": "TEST",
  "chain_description": "A test chain for development"
}`,
		},
		{
			Name:        "Get Chain",
			Method:      MethodGET,
			Path:        "/api/v1/chains/{id}",
			Category:    CategoryChains,
			Description: "Get chain by ID",
			PathParams: []PathParam{
				{
					Name:         "id",
					InitialValue: getFirstChainID,
				},
			},
		},
		{
			Name:        "Delete Chain",
			Method:      MethodDELETE,
			Path:        "/api/v1/chains/{id}",
			Category:    CategoryChains,
			Description: "Delete chain by ID",
			PathParams: []PathParam{
				{
					Name:         "id",
					InitialValue: getFirstChainID,
				},
			},
		},

		// Virtual Pools
		{
			Name:        "Get Virtual Pool",
			Method:      MethodGET,
			Path:        "/api/v1/chains/{id}/virtual-pool",
			Category:    CategoryVirtualPool,
			Description: "Get virtual pool for a chain",
			PathParams: []PathParam{
				{
					Name:         "id",
					InitialValue: getFirstChainID,
				},
			},
		},
		{
			Name:        "Get Transactions",
			Method:      MethodGET,
			Path:        "/api/v1/chains/{id}/transactions",
			Category:    CategoryVirtualPool,
			Description: "Get transactions for a chain",
			PathParams: []PathParam{
				{
					Name:         "id",
					InitialValue: getFirstChainID,
				},
			},
			QueryParams: []QueryParam{
				{Name: "page", Description: "Page number", Example: "1"},
				{Name: "limit", Description: "Items per page", Example: "10"},
				{Name: "user_id", Description: "Filter by user ID", Example: ""},
				{Name: "transaction_type", Description: "Filter by type", Example: "buy"},
			},
		},
	}
}

// GetEndpointsByCategory returns endpoints for a specific category
func GetEndpointsByCategory(category EndpointCategory) []Endpoint {
	all := GetAllEndpoints()
	var filtered []Endpoint
	for _, ep := range all {
		if ep.Category == category {
			filtered = append(filtered, ep)
		}
	}
	return filtered
}

// GetCategories returns all unique categories
func GetCategories() []EndpointCategory {
	return []EndpointCategory{
		CategoryHealth,
		CategoryAuth,
		CategoryTemplates,
		CategoryChains,
		CategoryVirtualPool,
	}
}

// ExecuteRequest executes an HTTP request and returns the result
func ExecuteRequest(baseURL, userID string, endpoint Endpoint, pathParamValues map[string]string, requestBody string, queryParams map[string]string) RequestResult {
	startTime := time.Now()
	result := RequestResult{
		RequestTime:  startTime,
		Method:       endpoint.Method,
		EndpointName: endpoint.Name,
		RequestBody:  requestBody,
	}

	// Build URL with path parameters
	url := baseURL + endpoint.Path
	for paramName, paramValue := range pathParamValues {
		url = replacePathParam(url, paramName, paramValue)
	}

	// Add query parameters
	if len(queryParams) > 0 {
		url += "?"
		first := true
		for key, value := range queryParams {
			if value == "" {
				continue
			}
			if !first {
				url += "&"
			}
			url += key + "=" + value
			first = false
		}
	}

	result.RequestURL = url
	result.RequestUserID = userID

	// Create request body reader
	var bodyReader io.Reader
	if requestBody != "" {
		bodyReader = bytes.NewBufferString(requestBody)
	}

	// Create HTTP request
	req, err := http.NewRequest(string(endpoint.Method), url, bodyReader)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result
	}

	// Format JSON response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
		result.Body = string(bodyBytes)
	} else {
		result.Body = prettyJSON.String()
	}

	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.Headers = resp.Header
	result.Duration = time.Since(startTime)

	return result
}

// replacePathParam replaces {paramName} in path with value
func replacePathParam(path, paramName, value string) string {
	placeholder := fmt.Sprintf("{%s}", paramName)
	return strings.ReplaceAll(path, placeholder, value)
}

// getFirstChainID returns the first chain ID from cached data
func getFirstChainID(m *Model) string {
	// Use cached chains from background fetch
	if len(m.cachedChains) > 0 {
		return m.cachedChains[0].ID
	}
	return ""
}

// findEndpointByName finds an endpoint by name
func findEndpointByName(name string, endpoints []Endpoint) Endpoint {
	for _, ep := range endpoints {
		if ep.Name == name {
			return ep
		}
	}
	return Endpoint{}
}

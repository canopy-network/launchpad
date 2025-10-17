package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const (
	// DefaultBaseURL is the default base URL for the launchpad API
	DefaultBaseURL = "http://localhost:3001"

	// TestUserID is a sample user ID from sample_data.sql (alice_dev)
	TestUserID = "550e8400-e29b-41d4-a716-446655440000"

	// APIPrefix is the API version prefix
	APIPrefix = "/api/v1"
)

// TestClient wraps http.Client with helper methods for testing
type TestClient struct {
	BaseURL string
	UserID  string
	Client  *http.Client
}

// NewTestClient creates a new test HTTP client with default settings
func NewTestClient() *TestClient {
	return &TestClient{
		BaseURL: DefaultBaseURL,
		UserID:  TestUserID,
		Client:  &http.Client{},
	}
}

// NewTestClientWithUser creates a test client with a specific user ID
func NewTestClientWithUser(userID string) *TestClient {
	return &TestClient{
		BaseURL: DefaultBaseURL,
		UserID:  userID,
		Client:  &http.Client{},
	}
}

// Get performs a GET request to the specified path
func (c *TestClient) Get(t *testing.T, path string) (*http.Response, []byte) {
	return c.DoRequest(t, "GET", path, nil)
}

// Post performs a POST request with JSON body
func (c *TestClient) Post(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return c.DoRequest(t, "POST", path, body)
}

// Put performs a PUT request with JSON body
func (c *TestClient) Put(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return c.DoRequest(t, "PUT", path, body)
}

// Delete performs a DELETE request
func (c *TestClient) Delete(t *testing.T, path string) (*http.Response, []byte) {
	return c.DoRequest(t, "DELETE", path, nil)
}

// DoRequest performs an HTTP request with the specified method, path, and optional body
func (c *TestClient) DoRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", c.UserID)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, responseBody
}

// UnmarshalResponse unmarshals response body into target struct
func UnmarshalResponse(t *testing.T, body []byte, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("Failed to unmarshal response: %v\nBody: %s", err, string(body))
	}
}

// AssertStatus checks that the response has the expected status code
func AssertStatus(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

// AssertStatusOK checks that the response is 200 OK
func AssertStatusOK(t *testing.T, resp *http.Response) {
	AssertStatus(t, resp, http.StatusOK)
}

// AssertStatusCreated checks that the response is 201 Created
func AssertStatusCreated(t *testing.T, resp *http.Response) {
	AssertStatus(t, resp, http.StatusCreated)
}

// PrintResponse is a helper to print response for debugging
func PrintResponse(t *testing.T, resp *http.Response, body []byte) {
	t.Helper()
	t.Logf("Response Status: %d", resp.StatusCode)
	t.Logf("Response Body: %s", string(body))
}

// StandardResponse represents the standard API response wrapper
type StandardResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *ErrorResponse  `json:"error,omitempty"`
}

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// GetAPIPath constructs a full API path with the version prefix
func GetAPIPath(path string) string {
	return fmt.Sprintf("%s%s", APIPrefix, path)
}

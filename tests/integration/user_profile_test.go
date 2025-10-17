//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/tests/testutils"
)

// TestUpdateUserProfile tests the basic user profile update flow
func TestUpdateUserProfile(t *testing.T) {
	client := testutils.NewTestClient()

	// First, authenticate to get a session token
	token := authenticateTestUser(t, client)

	// Create an authenticated client
	authClient := &AuthenticatedTestClient{
		TestClient: client,
		Token:      token,
	}

	// Update profile with username and bio
	t.Log("Updating user profile...")
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	bio := "This is my test bio"
	updateRequest := map[string]interface{}{
		"username": username,
		"bio":      bio,
	}

	profilePath := testutils.GetAPIPath("/users/profile")
	resp, body := authClient.Put(t, profilePath, updateRequest)

	testutils.AssertStatusOK(t, resp)

	// Validate the updated profile
	var profileResponse struct {
		Data struct {
			User    models.User `json:"user"`
			Message string      `json:"message"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &profileResponse)

	user := profileResponse.Data.User

	if user.Username == nil || *user.Username != username {
		t.Errorf("Expected username='%s', got '%v'", username, user.Username)
	}

	if user.Bio == nil || *user.Bio != bio {
		t.Errorf("Expected bio='%s', got '%v'", bio, user.Bio)
	}

	if profileResponse.Data.Message != "Profile updated successfully" {
		t.Errorf("Unexpected message: %s", profileResponse.Data.Message)
	}

	t.Logf("Successfully updated profile - Username: %s, Bio: %s", *user.Username, *user.Bio)
}

// TestUpdateUserProfilePartialUpdate tests partial updates (only some fields)
func TestUpdateUserProfilePartialUpdate(t *testing.T) {
	client := testutils.NewTestClient()
	token := authenticateTestUser(t, client)
	authClient := &AuthenticatedTestClient{
		TestClient: client,
		Token:      token,
	}

	tests := []struct {
		name           string
		updateFields   map[string]interface{}
		validateFields map[string]string
	}{
		{
			name: "Update display name only",
			updateFields: map[string]interface{}{
				"display_name": "Test Display Name",
			},
			validateFields: map[string]string{
				"display_name": "Test Display Name",
			},
		},
		{
			name: "Update avatar URL only",
			updateFields: map[string]interface{}{
				"avatar_url": "https://example.com/avatar.jpg",
			},
			validateFields: map[string]string{
				"avatar_url": "https://example.com/avatar.jpg",
			},
		},
		{
			name: "Update social handles",
			updateFields: map[string]interface{}{
				"twitter_handle":  "@testuser",
				"github_username": "testuser",
				"telegram_handle": "@testuser",
			},
			validateFields: map[string]string{
				"twitter_handle":  "@testuser",
				"github_username": "testuser",
				"telegram_handle": "@testuser",
			},
		},
		{
			name: "Update multiple fields",
			updateFields: map[string]interface{}{
				"display_name": "Updated Name",
				"bio":          "Updated bio text",
				"website_url":  "https://example.com",
			},
			validateFields: map[string]string{
				"display_name": "Updated Name",
				"bio":          "Updated bio text",
				"website_url":  "https://example.com",
			},
		},
	}

	profilePath := testutils.GetAPIPath("/users/profile")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := authClient.Put(t, profilePath, tt.updateFields)
			testutils.AssertStatusOK(t, resp)

			var profileResponse struct {
				Data struct {
					User models.User `json:"user"`
				} `json:"data"`
			}
			testutils.UnmarshalResponse(t, body, &profileResponse)

			user := profileResponse.Data.User

			// Validate each expected field
			for field, expectedValue := range tt.validateFields {
				var actualValue *string
				switch field {
				case "display_name":
					actualValue = user.DisplayName
				case "bio":
					actualValue = user.Bio
				case "avatar_url":
					actualValue = user.AvatarURL
				case "website_url":
					actualValue = user.WebsiteURL
				case "twitter_handle":
					actualValue = user.TwitterHandle
				case "github_username":
					actualValue = user.GithubUsername
				case "telegram_handle":
					actualValue = user.TelegramHandle
				}

				if actualValue == nil || *actualValue != expectedValue {
					t.Errorf("Expected %s='%s', got '%v'", field, expectedValue, actualValue)
				}
			}

			t.Logf("Successfully validated partial update: %s", tt.name)
		})
	}
}

// TestUpdateUserProfileValidation tests validation errors
func TestUpdateUserProfileValidation(t *testing.T) {
	client := testutils.NewTestClient()
	token := authenticateTestUser(t, client)
	authClient := &AuthenticatedTestClient{
		TestClient: client,
		Token:      token,
	}

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name:           "Empty request (no fields)",
			request:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when no fields are provided",
		},
		{
			name: "Invalid username (too short)",
			request: map[string]interface{}{
				"username": "ab",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when username is less than 3 characters",
		},
		{
			name: "Invalid username (non-alphanumeric)",
			request: map[string]interface{}{
				"username": "test@user",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when username contains non-alphanumeric characters",
		},
		{
			name: "Invalid avatar URL",
			request: map[string]interface{}{
				"avatar_url": "not-a-url",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when avatar_url is not a valid URL",
		},
		{
			name: "Invalid website URL",
			request: map[string]interface{}{
				"website_url": "invalid-url",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when website_url is not a valid URL",
		},
		{
			name: "Bio too long",
			request: map[string]interface{}{
				"bio": string(make([]byte, 501)), // 501 characters
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when bio exceeds 500 characters",
		},
	}

	profilePath := testutils.GetAPIPath("/users/profile")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := authClient.Put(t, profilePath, tt.request)

			testutils.AssertStatus(t, resp, tt.expectedStatus)

			var errorResponse struct {
				Error *testutils.ErrorResponse `json:"error"`
			}
			testutils.UnmarshalResponse(t, body, &errorResponse)

			if errorResponse.Error == nil {
				t.Error("Expected error response")
			} else {
				t.Logf("Validation error: %s - %s", errorResponse.Error.Code, errorResponse.Error.Message)
			}
		})
	}
}

// TestUpdateUserProfileDuplicateUsername tests username uniqueness constraint
func TestUpdateUserProfileDuplicateUsername(t *testing.T) {
	client := testutils.NewTestClient()

	// Create two users
	token1 := authenticateTestUser(t, client)
	email2 := fmt.Sprintf("user2_%d@example.com", time.Now().Unix())
	token2 := authenticateUser(t, client, email2)

	authClient1 := &AuthenticatedTestClient{
		TestClient: client,
		Token:      token1,
	}
	authClient2 := &AuthenticatedTestClient{
		TestClient: client,
		Token:      token2,
	}

	// Set username for user 1
	username := fmt.Sprintf("unique_user_%d", time.Now().Unix())
	profilePath := testutils.GetAPIPath("/users/profile")

	resp, _ := authClient1.Put(t, profilePath, map[string]interface{}{
		"username": username,
	})
	testutils.AssertStatusOK(t, resp)
	t.Logf("User 1 claimed username: %s", username)

	// Try to set the same username for user 2 - should fail
	t.Log("Attempting to use duplicate username...")
	resp, body := authClient2.Put(t, profilePath, map[string]interface{}{
		"username": username,
	})

	testutils.AssertStatus(t, resp, http.StatusConflict)

	var errorResponse struct {
		Error *testutils.ErrorResponse `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error == nil || errorResponse.Error.Message != "Username is already taken" {
		t.Errorf("Expected 'Username is already taken' error, got: %v", errorResponse.Error)
	}

	t.Log("Duplicate username correctly rejected")
}

// TestUpdateUserProfileUnauthenticated tests that unauthenticated requests are rejected
func TestUpdateUserProfileUnauthenticated(t *testing.T) {
	client := testutils.NewTestClient()

	updateRequest := map[string]interface{}{
		"username": "testuser",
	}

	profilePath := testutils.GetAPIPath("/users/profile")
	resp, body := client.DoRequest(t, "PUT", profilePath, updateRequest)

	testutils.AssertStatus(t, resp, http.StatusUnauthorized)

	var errorResponse struct {
		Error *testutils.ErrorResponse `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error == nil {
		t.Error("Expected error response for unauthenticated request")
	}

	t.Logf("Unauthenticated request correctly rejected: %s", errorResponse.Error.Message)
}

// Helper functions

// AuthenticatedTestClient wraps TestClient with authentication token
type AuthenticatedTestClient struct {
	*testutils.TestClient
	Token string
}

// Put performs a PUT request with authentication
func (c *AuthenticatedTestClient) Put(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return c.DoAuthenticatedRequest(t, "PUT", path, body)
}

// DoAuthenticatedRequest performs an HTTP request with Bearer token authentication
func (c *AuthenticatedTestClient) DoAuthenticatedRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
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

	// Set headers with Bearer token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

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

// authenticateTestUser authenticates the default test user and returns token
func authenticateTestUser(t *testing.T, client *testutils.TestClient) string {
	return authenticateUser(t, client, "test@example.com")
}

// authenticateUser authenticates a user by email and returns the session token
func authenticateUser(t *testing.T, client *testutils.TestClient, email string) string {
	t.Helper()

	// Step 1: Request verification code
	sendCodePath := testutils.GetAPIPath("/auth/email")
	resp, body := client.Post(t, sendCodePath, map[string]interface{}{
		"email": email,
	})
	testutils.AssertStatusOK(t, resp)

	var sendCodeResponse struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &sendCodeResponse)
	code := sendCodeResponse.Data.Code

	if code == "" {
		t.Fatal("No verification code returned")
	}

	// Step 2: Verify code and get token
	verifyPath := testutils.GetAPIPath("/auth/verify")
	resp, body = client.Post(t, verifyPath, map[string]interface{}{
		"email": email,
		"code":  code,
	})
	testutils.AssertStatusOK(t, resp)

	// Extract token from Authorization header
	authHeader := resp.Header.Get("Authorization")
	if authHeader == "" {
		t.Fatal("No Authorization header in verify response")
	}

	// Remove "Bearer " prefix
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token := authHeader[7:]
		t.Logf("Authenticated user %s with token: %s...", email, token[:8])
		return token
	}

	t.Fatal("Invalid Authorization header format")
	return ""
}

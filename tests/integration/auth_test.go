//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/enielson/launchpad/tests/testutils"
)

// TestEmailAuthFlow tests the complete email authentication flow:
// 1. Send email verification code
// 2. Verify the code
func TestEmailAuthFlow(t *testing.T) {
	client := testutils.NewTestClient()

	// Step 1: Request email verification code
	t.Log("Requesting email verification code...")
	email := "test@example.com"

	sendCodeRequest := map[string]interface{}{
		"email": email,
	}

	sendCodePath := testutils.GetAPIPath("/auth/email")
	resp, body := client.Post(t, sendCodePath, sendCodeRequest)

	testutils.AssertStatusOK(t, resp)

	// Parse response to get the code (in development mode, code is returned)
	var sendCodeResponse struct {
		Data struct {
			Message string `json:"message"`
			Email   string `json:"email"`
			Code    string `json:"code"` // Only returned in development
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &sendCodeResponse)

	if sendCodeResponse.Data.Email != email {
		t.Errorf("Expected email %s, got %s", email, sendCodeResponse.Data.Email)
	}

	code := sendCodeResponse.Data.Code
	if code == "" {
		t.Fatal("No verification code returned (expected in development mode)")
	}

	t.Logf("Received verification code: %s", code)

	// Step 2: Verify the code
	t.Log("Verifying email code...")
	verifyCodeRequest := map[string]interface{}{
		"email": email,
		"code":  code,
	}

	verifyPath := testutils.GetAPIPath("/auth/verify")
	resp, body = client.Post(t, verifyPath, verifyCodeRequest)

	testutils.AssertStatusOK(t, resp)

	// Parse verification response
	var verifyResponse struct {
		Data struct {
			Message string `json:"message"`
			Email   string `json:"email"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &verifyResponse)

	if verifyResponse.Data.Message != "Email verified successfully" {
		t.Errorf("Unexpected verification message: %s", verifyResponse.Data.Message)
	}

	t.Log("Email verification successful!")
}

// TestEmailAuthInvalidCode tests verification with invalid code
func TestEmailAuthInvalidCode(t *testing.T) {
	client := testutils.NewTestClient()

	// Step 1: Send verification code
	t.Log("Requesting email verification code...")
	email := "invalid_test@example.com"

	sendCodeRequest := map[string]interface{}{
		"email": email,
	}

	sendCodePath := testutils.GetAPIPath("/auth/email")
	resp, _ := client.Post(t, sendCodePath, sendCodeRequest)
	testutils.AssertStatusOK(t, resp)

	// Step 2: Try to verify with wrong code
	t.Log("Attempting verification with invalid code...")
	verifyCodeRequest := map[string]interface{}{
		"email": email,
		"code":  "000000", // Wrong code
	}

	verifyPath := testutils.GetAPIPath("/auth/verify")
	resp, body := client.Post(t, verifyPath, verifyCodeRequest)

	// Should get 400 Bad Request for invalid code
	testutils.AssertStatus(t, resp, http.StatusBadRequest)

	var errorResponse struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error.Message != "Invalid verification code" {
		t.Errorf("Expected 'Invalid verification code', got: %s", errorResponse.Error.Message)
	}

	t.Log("Invalid code correctly rejected")
}

// TestEmailAuthExpiredCode tests verification with expired code
func TestEmailAuthExpiredCode(t *testing.T) {
	// Note: This test would require manipulating time or waiting 10 minutes
	// For now, we'll just test the validation
	t.Skip("Code expiration test requires time manipulation - implement when needed")
}

// TestEmailAuthValidation tests input validation
func TestEmailAuthValidation(t *testing.T) {
	client := testutils.NewTestClient()

	tests := []struct {
		name           string
		email          string
		code           string
		endpoint       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing email on send",
			email:          "",
			endpoint:       "/auth/email",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request validation failed",
		},
		{
			name:           "Invalid email format",
			email:          "not-an-email",
			endpoint:       "/auth/email",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request validation failed",
		},
		{
			name:           "Missing code on verify",
			email:          "test@example.com",
			code:           "",
			endpoint:       "/auth/verify",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request validation failed",
		},
		{
			name:           "Invalid code length",
			email:          "test@example.com",
			code:           "123", // Too short
			endpoint:       "/auth/verify",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request validation failed",
		},
		{
			name:           "Non-numeric code",
			email:          "test@example.com",
			code:           "ABCDEF",
			endpoint:       "/auth/verify",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Request validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var resp *http.Response

			path := testutils.GetAPIPath(tt.endpoint)

			if tt.endpoint == "/auth/email" {
				req := map[string]interface{}{
					"email": tt.email,
				}
				resp, body = client.Post(t, path, req)
			} else if tt.endpoint == "/auth/verify" {
				req := map[string]interface{}{
					"email": tt.email,
					"code":  tt.code,
				}
				resp, body = client.Post(t, path, req)
			}

			testutils.AssertStatus(t, resp, tt.expectedStatus)

			var errorResponse struct {
				Error struct {
					Code    string      `json:"code"`
					Message string      `json:"message"`
					Details interface{} `json:"details"`
				} `json:"error"`
			}
			testutils.UnmarshalResponse(t, body, &errorResponse)

			if errorResponse.Error.Message != tt.expectedError {
				t.Errorf("Expected error '%s', got: %s", tt.expectedError, errorResponse.Error.Message)
			}

			t.Logf("Validation error correctly caught: %s", errorResponse.Error.Code)
		})
	}
}

// TestEmailAuthCodeReuse tests that codes can only be used once
func TestEmailAuthCodeReuse(t *testing.T) {
	client := testutils.NewTestClient()

	// Step 1: Request and verify code successfully
	t.Log("Requesting email verification code...")
	email := "reuse_test@example.com"

	sendCodeRequest := map[string]interface{}{
		"email": email,
	}

	sendCodePath := testutils.GetAPIPath("/auth/email")
	resp, body := client.Post(t, sendCodePath, sendCodeRequest)
	testutils.AssertStatusOK(t, resp)

	var sendCodeResponse struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &sendCodeResponse)
	code := sendCodeResponse.Data.Code

	// Verify code first time - should succeed
	t.Log("First verification attempt...")
	verifyPath := testutils.GetAPIPath("/auth/verify")
	verifyRequest := map[string]interface{}{
		"email": email,
		"code":  code,
	}

	resp, _ = client.Post(t, verifyPath, verifyRequest)
	testutils.AssertStatusOK(t, resp)
	t.Log("First verification successful")

	// Try to reuse the same code - should fail
	t.Log("Attempting to reuse code...")
	resp, body = client.Post(t, verifyPath, verifyRequest)

	testutils.AssertStatus(t, resp, http.StatusBadRequest)

	var errorResponse struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error.Message != "Invalid verification code" {
		t.Errorf("Expected 'Invalid verification code' on reuse, got: %s", errorResponse.Error.Message)
	}

	t.Log("Code reuse correctly prevented")
}

// TestEmailAuthMultipleUsers tests that codes are isolated per email
func TestEmailAuthMultipleUsers(t *testing.T) {
	client := testutils.NewTestClient()

	// Request codes for two different emails
	email1 := "user1@example.com"
	email2 := "user2@example.com"

	sendCodePath := testutils.GetAPIPath("/auth/email")

	// Send code to user 1
	t.Log("Sending code to user 1...")
	resp1, body1 := client.Post(t, sendCodePath, map[string]interface{}{"email": email1})
	testutils.AssertStatusOK(t, resp1)

	var resp1Data struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body1, &resp1Data)
	code1 := resp1Data.Data.Code

	// Send code to user 2
	t.Log("Sending code to user 2...")
	resp2, body2 := client.Post(t, sendCodePath, map[string]interface{}{"email": email2})
	testutils.AssertStatusOK(t, resp2)

	var resp2Data struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body2, &resp2Data)
	code2 := resp2Data.Data.Code

	// Ensure codes are different
	if code1 == code2 {
		t.Error("Generated codes should be different for different users")
	}

	// Verify user 1 with their code - should succeed
	t.Log("Verifying user 1 with correct code...")
	verifyPath := testutils.GetAPIPath("/auth/verify")
	resp, _ := client.Post(t, verifyPath, map[string]interface{}{
		"email": email1,
		"code":  code1,
	})
	testutils.AssertStatusOK(t, resp)

	// Try to verify user 1 with user 2's code - should fail
	t.Log("Attempting to verify user 1 with user 2's code...")
	resp, body := client.Post(t, verifyPath, map[string]interface{}{
		"email": email1,
		"code":  code2,
	})
	testutils.AssertStatus(t, resp, http.StatusBadRequest)

	var errorResponse struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error.Message != "Invalid verification code" {
		t.Errorf("Expected 'Invalid verification code', got: %s", errorResponse.Error.Message)
	}

	t.Log("Code isolation between users working correctly")
}

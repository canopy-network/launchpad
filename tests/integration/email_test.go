//go:build integration

package integration_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/enielson/launchpad/tests/testutils"
)

// TestSendEmailToSMTPUser tests sending an email to the configured SMTP user
// This is a real email test - it will actually send an email via SMTP
func TestSendEmailToSMTPUser(t *testing.T) {
	// Get the SMTP username from environment
	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		t.Skip("SMTP_USERNAME not configured, skipping real email test")
	}

	client := testutils.NewTestClient()

	t.Logf("Sending verification code to: %s", smtpUsername)

	// Request email verification code
	sendCodeRequest := map[string]interface{}{
		"email": smtpUsername,
	}

	sendCodePath := testutils.GetAPIPath("/auth/email")
	resp, body := client.Post(t, sendCodePath, sendCodeRequest)

	testutils.AssertStatusOK(t, resp)

	// Parse response
	var sendCodeResponse struct {
		Data struct {
			Message string `json:"message"`
			Email   string `json:"email"`
			Code    string `json:"code"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &sendCodeResponse)

	if sendCodeResponse.Data.Email != smtpUsername {
		t.Errorf("Expected email %s, got %s", smtpUsername, sendCodeResponse.Data.Email)
	}

	code := sendCodeResponse.Data.Code
	if code == "" {
		t.Fatal("No verification code returned")
	}

	t.Logf("✅ Email sent successfully!")
	t.Logf("Verification code: %s", code)
	t.Logf("Check your inbox at: %s", smtpUsername)

	// Verify the code works
	t.Log("Verifying the code...")
	verifyRequest := map[string]interface{}{
		"email": smtpUsername,
		"code":  code,
	}

	verifyPath := testutils.GetAPIPath("/auth/verify")
	resp, body = client.Post(t, verifyPath, verifyRequest)

	testutils.AssertStatusOK(t, resp)

	var verifyResponse struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &verifyResponse)

	if verifyResponse.Data.Message != "Email verified successfully" {
		t.Errorf("Expected success message, got: %s", verifyResponse.Data.Message)
	}

	t.Log("✅ Code verification successful!")
}

// TestSendEmailInvalidRecipient tests error handling for invalid email addresses
func TestSendEmailInvalidRecipient(t *testing.T) {
	client := testutils.NewTestClient()

	// Try to send to an invalid email
	sendCodeRequest := map[string]interface{}{
		"email": "not-a-valid-email",
	}

	sendCodePath := testutils.GetAPIPath("/auth/email")
	resp, body := client.Post(t, sendCodePath, sendCodeRequest)

	// Should get validation error
	testutils.AssertStatus(t, resp, http.StatusBadRequest)

	var errorResponse struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error.Message != "Request validation failed" {
		t.Errorf("Expected validation error, got: %s", errorResponse.Error.Message)
	}

	t.Log("✅ Invalid email correctly rejected")
}

// TestSendEmailMultipleTimes tests sending multiple emails to the same address
func TestSendEmailMultipleTimes(t *testing.T) {
	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		t.Skip("SMTP_USERNAME not configured, skipping real email test")
	}

	client := testutils.NewTestClient()
	sendCodePath := testutils.GetAPIPath("/auth/email")

	t.Logf("Sending 3 verification codes to: %s", smtpUsername)

	var lastCode string

	// Send 3 emails
	for i := 1; i <= 3; i++ {
		sendCodeRequest := map[string]interface{}{
			"email": smtpUsername,
		}

		resp, body := client.Post(t, sendCodePath, sendCodeRequest)
		testutils.AssertStatusOK(t, resp)

		var response struct {
			Data struct {
				Code string `json:"code"`
			} `json:"data"`
		}
		testutils.UnmarshalResponse(t, body, &response)

		lastCode = response.Data.Code
		t.Logf("Email %d sent - Code: %s", i, lastCode)
	}

	// Only the last code should work
	t.Log("Verifying the most recent code...")
	verifyRequest := map[string]interface{}{
		"email": smtpUsername,
		"code":  lastCode,
	}

	verifyPath := testutils.GetAPIPath("/auth/verify")
	resp, _ := client.Post(t, verifyPath, verifyRequest)

	testutils.AssertStatusOK(t, resp)

	t.Logf("✅ Successfully sent 3 emails and verified the latest code")
	t.Logf("Check your inbox at: %s (you should have 3 emails)", smtpUsername)
}

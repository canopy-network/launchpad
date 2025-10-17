//go:build integration

package integration_test

import (
	"testing"
	"time"

	"github.com/enielson/launchpad/tests/testutils"
)

// TestEmailRateLimit tests that email sending is rate limited to 1 per minute
// Disabled: Rate limiting is only enabled in production environment
func TestEmailRateLimit(t *testing.T) {
	t.Skip("Rate limiting is disabled in non-production environments")
	client := testutils.NewTestClient()
	emailPath := testutils.GetAPIPath("/auth/email")

	email := "ratelimit@test.com"
	requestBody := map[string]interface{}{
		"email": email,
	}

	// First request should succeed
	t.Run("first_request_succeeds", func(t *testing.T) {
		resp, _ := client.Post(t, emailPath, requestBody)
		testutils.AssertStatusOK(t, resp)
		t.Log("First request succeeded (200 OK)")
	})

	// Second request immediately after should be rate limited
	t.Run("second_request_rate_limited", func(t *testing.T) {
		resp, body := client.Post(t, emailPath, requestBody)

		if resp.StatusCode != 429 {
			t.Errorf("Expected status 429 Too Many Requests, got %d", resp.StatusCode)
		}

		var errorResponse struct {
			Error struct {
				Code    string      `json:"code"`
				Message string      `json:"message"`
				Details interface{} `json:"details"`
			} `json:"error"`
		}
		testutils.UnmarshalResponse(t, body, &errorResponse)

		if errorResponse.Error.Code != "RATE_LIMIT_EXCEEDED" {
			t.Errorf("Expected error code RATE_LIMIT_EXCEEDED, got %s", errorResponse.Error.Code)
		}

		t.Logf("Second request correctly rate limited (429): %s", errorResponse.Error.Message)
	})

	// Wait for rate limit to expire, then request should succeed again
	t.Run("request_after_cooldown_succeeds", func(t *testing.T) {
		t.Log("Waiting 11 seconds for rate limit to expire...")
		time.Sleep(11 * time.Second)

		resp, _ := client.Post(t, emailPath, requestBody)
		testutils.AssertStatusOK(t, resp)
		t.Log("Request after cooldown succeeded (200 OK)")
	})
}

// TestEmailRateLimitIsolation tests that rate limiting is per-IP
// Disabled: Rate limiting is only enabled in production environment
func TestEmailRateLimitIsolation(t *testing.T) {
	t.Skip("Rate limiting is disabled in non-production environments")
	// Note: This test would require running from different IPs
	// For now, we just test that the same IP gets rate limited
	client := testutils.NewTestClient()
	emailPath := testutils.GetAPIPath("/auth/email")

	// First email
	email1 := "user1@test.com"
	req1 := map[string]interface{}{"email": email1}

	resp1, _ := client.Post(t, emailPath, req1)
	testutils.AssertStatusOK(t, resp1)

	// Different email, same IP - should still be rate limited
	email2 := "user2@test.com"
	req2 := map[string]interface{}{"email": email2}

	resp2, _ := client.Post(t, emailPath, req2)
	if resp2.StatusCode != 429 {
		t.Errorf("Expected rate limit (429) for same IP different email, got %d", resp2.StatusCode)
	}

	t.Log("Rate limiting correctly applies per IP, not per email")
}

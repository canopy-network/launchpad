//go:build integration

package integration_test

import (
	"encoding/json"
	"testing"

	"github.com/enielson/launchpad/tests/testutils"
)

// TestListRoutes tests the routes listing endpoint
func TestListRoutes(t *testing.T) {
	client := testutils.NewTestClient()

	routesPath := testutils.GetAPIPath("/routes")
	resp, body := client.Get(t, routesPath)

	testutils.AssertStatusOK(t, resp)

	// Parse response
	var routesResponse struct {
		Data  []map[string]interface{} `json:"data"`
		Count int                      `json:"count"`
	}
	testutils.UnmarshalResponse(t, body, &routesResponse)

	// Verify we have routes
	if routesResponse.Count == 0 {
		t.Error("Expected routes to be listed, got 0")
	}

	if len(routesResponse.Data) == 0 {
		t.Error("Expected data array to contain routes")
	}

	t.Logf("Found %d routes", routesResponse.Count)

	// Verify structure of first route
	if len(routesResponse.Data) > 0 {
		route := routesResponse.Data[0]

		// Check required fields exist
		if _, ok := route["method"]; !ok {
			t.Error("Route missing 'method' field")
		}
		if _, ok := route["path"]; !ok {
			t.Error("Route missing 'path' field")
		}
		if _, ok := route["middleware_count"]; !ok {
			t.Error("Route missing 'middleware_count' field")
		}

		t.Logf("Example route: %s %s (middlewares: %.0f)",
			route["method"], route["path"], route["middleware_count"])
	}

	// Verify expected routes are present
	expectedRoutes := map[string]bool{
		"GET /health":              false,
		"GET /api/v1/routes":       false,
		"POST /api/v1/auth/email":  false,
		"POST /api/v1/auth/verify": false,
		"GET /api/v1/templates":    false,
		"GET /api/v1/chains":       false,
		"POST /api/v1/chains":      false,
	}

	for _, r := range routesResponse.Data {
		method := r["method"].(string)
		path := r["path"].(string)
		key := method + " " + path

		if _, exists := expectedRoutes[key]; exists {
			expectedRoutes[key] = true
		}
	}

	// Check that all expected routes were found
	for route, found := range expectedRoutes {
		if !found {
			t.Errorf("Expected route not found: %s", route)
		}
	}

	// Pretty print all routes for debugging
	t.Log("All registered routes:")
	prettyJSON, _ := json.MarshalIndent(routesResponse.Data, "", "  ")
	t.Logf("\n%s", string(prettyJSON))
}

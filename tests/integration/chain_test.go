//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/tests/fixtures"
	"github.com/enielson/launchpad/tests/testutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestCreateChain tests the full chain creation flow:
// 1. Create a test template using fixtures
// 2. Use the template ID to create a new chain
// 3. Validate the chain was created successfully
func TestCreateChain(t *testing.T) {
	var templateID uuid.UUID

	// Step 0: Create test template using fixture (persists for HTTP test)
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		templateName := fmt.Sprintf("Integration Test Template %d", time.Now().UnixNano())
		template, err := fixtures.DefaultChainTemplate().
			WithName(templateName).
			WithCategory("defi").
			Create(context.Background(), db)

		require.NoError(t, err)
		templateID = template.ID

		// Cleanup: Remove template after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_templates WHERE id = $1", templateID)
		})

		t.Logf("Created test template: %s (ID: %s)", template.TemplateName, template.ID)
	})

	client := testutils.NewTestClient()

	// Step 1: Create a new chain using the template
	t.Log("Creating new chain...")
	chainName := fmt.Sprintf("Integration Test Chain %d", time.Now().Unix())
	createChainRequest := map[string]interface{}{
		"chain_name":   chainName,
		"token_symbol": "INTTEST",
		"template_id":  templateID.String(),
	}

	chainsPath := testutils.GetAPIPath("/chains")
	resp, body := client.Post(t, chainsPath, createChainRequest)

	testutils.AssertStatusCreated(t, resp)

	// Step 3: Validate the created chain
	var chainResponse struct {
		Data models.Chain `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &chainResponse)

	chain := chainResponse.Data

	// Validate chain fields
	if chain.ChainName != chainName {
		t.Errorf("Expected chain_name='%s', got '%s'", chainName, chain.ChainName)
	}

	if chain.TokenSymbol != "INTTEST" {
		t.Errorf("Expected token_symbol='INTTEST', got '%s'", chain.TokenSymbol)
	}

	if chain.TemplateID == nil || *chain.TemplateID != templateID {
		t.Errorf("Expected template_id='%s', got '%v'", templateID, chain.TemplateID)
	}

	if chain.Status != models.ChainStatusDraft {
		t.Errorf("Expected status='draft', got '%s'", chain.Status)
	}

	if chain.ID.String() == "" {
		t.Error("Expected chain ID to be set")
	}

	if chain.CreatedBy.String() != testutils.TestUserID {
		t.Errorf("Expected created_by='%s', got '%s'", testutils.TestUserID, chain.CreatedBy)
	}

	if chain.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}

	t.Logf("Successfully created chain with ID: %s", chain.ID)

	// Step 4: Fetch the chain again to verify it was persisted
	t.Log("Fetching created chain to verify persistence...")
	getChainPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s", chain.ID))
	resp, body = client.Get(t, getChainPath)

	testutils.AssertStatusOK(t, resp)

	var fetchedChainResponse struct {
		Data models.Chain `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &fetchedChainResponse)

	fetchedChain := fetchedChainResponse.Data

	// Verify fetched chain matches created chain
	if fetchedChain.ID != chain.ID {
		t.Errorf("Fetched chain ID mismatch: expected '%s', got '%s'", chain.ID, fetchedChain.ID)
	}

	if fetchedChain.ChainName != chainName {
		t.Errorf("Fetched chain name mismatch: expected '%s', got '%s'", chainName, fetchedChain.ChainName)
	}

	if fetchedChain.TokenSymbol != "INTTEST" {
		t.Errorf("Fetched token symbol mismatch: expected 'INTTEST', got '%s'", fetchedChain.TokenSymbol)
	}

	t.Logf("Successfully verified chain persistence")
}

// TestCreateChainWithoutTemplate tests creating a chain without a template ID
func TestCreateChainWithoutTemplate(t *testing.T) {
	client := testutils.NewTestClient()

	// Use timestamp to ensure unique chain name
	chainName := fmt.Sprintf("No Template Chain %d", time.Now().Unix())

	createChainRequest := map[string]interface{}{
		"chain_name":   chainName,
		"token_symbol": "NOTMPL",
	}

	chainsPath := testutils.GetAPIPath("/chains")
	resp, body := client.Post(t, chainsPath, createChainRequest)

	testutils.AssertStatusCreated(t, resp)

	var chainResponse struct {
		Data models.Chain `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &chainResponse)

	chain := chainResponse.Data

	if chain.ChainName != chainName {
		t.Errorf("Expected chain_name='%s', got '%s'", chainName, chain.ChainName)
	}

	if chain.TemplateID != nil {
		t.Errorf("Expected template_id to be nil, got '%v'", chain.TemplateID)
	}

	if chain.Status != models.ChainStatusDraft {
		t.Errorf("Expected status='draft', got '%s'", chain.Status)
	}

	t.Logf("Successfully created chain without template, ID: %s", chain.ID)
}

// TestCreateChainValidation tests validation errors
func TestCreateChainValidation(t *testing.T) {
	client := testutils.NewTestClient()

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name:           "Missing chain name",
			request:        map[string]interface{}{"token_symbol": "TEST"},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when chain_name is missing",
		},
		{
			name:           "Missing token symbol",
			request:        map[string]interface{}{"chain_name": "Test Chain"},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when token_symbol is missing",
		},
		{
			name: "Lowercase token symbol",
			request: map[string]interface{}{
				"chain_name":   "Test Chain",
				"token_symbol": "lowercase",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when token_symbol is not uppercase",
		},
		{
			name: "Invalid template ID format",
			request: map[string]interface{}{
				"chain_name":   "Test Chain",
				"token_symbol": "TEST",
				"template_id":  "not-a-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when template_id is not a valid UUID",
		},
	}

	chainsPath := testutils.GetAPIPath("/chains")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := client.Post(t, chainsPath, tt.request)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d. Body: %s",
					tt.description, tt.expectedStatus, resp.StatusCode, string(body))
			}

			// Validate error response structure
			if resp.StatusCode >= 400 {
				var errorResponse struct {
					Error *testutils.ErrorResponse `json:"error"`
				}
				if err := json.Unmarshal(body, &errorResponse); err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				} else if errorResponse.Error == nil {
					t.Error("Expected error field in response")
				} else {
					t.Logf("Error code: %s, message: %s", errorResponse.Error.Code, errorResponse.Error.Message)
				}
			}
		})
	}
}

// TestGetChains tests fetching the list of chains
func TestGetChains(t *testing.T) {
	var chain1ID, chain2ID uuid.UUID

	// Setup: Create test chains using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		fixture1 := fixtures.DefaultChain(creatorID)
		fixture1.ChainName = fmt.Sprintf("Test Chain 1 %d", timestamp)
		chain1, err := fixture1.WithTokenSymbol("TST1").
			Create(context.Background(), db)
		require.NoError(t, err)
		chain1ID = chain1.ID

		fixture2 := fixtures.DefaultChain(creatorID)
		fixture2.ChainName = fmt.Sprintf("Test Chain 2 %d", timestamp)
		chain2, err := fixture2.WithTokenSymbol("TST2").
			Create(context.Background(), db)
		require.NoError(t, err)
		chain2ID = chain2.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id IN ($1, $2)",
				chain1ID, chain2ID)
		})

		t.Logf("Created test chains: %s, %s", chain1ID, chain2ID)
	})

	client := testutils.NewTestClient()

	chainsPath := testutils.GetAPIPath("/chains")
	resp, body := client.Get(t, chainsPath)

	testutils.AssertStatusOK(t, resp)

	var chainsResponse struct {
		Data       []models.Chain `json:"data"`
		Pagination *struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
			Total int `json:"total"`
			Pages int `json:"pages"`
		} `json:"pagination,omitempty"`
	}
	testutils.UnmarshalResponse(t, body, &chainsResponse)

	// Should have at least our 2 test chains
	if len(chainsResponse.Data) < 2 {
		t.Errorf("Expected at least 2 chains, got %d", len(chainsResponse.Data))
	}

	t.Logf("Found %d chains", len(chainsResponse.Data))

	// Verify our test chains are in the results
	foundChain1 := false
	foundChain2 := false
	for _, chain := range chainsResponse.Data {
		if chain.ID == chain1ID {
			foundChain1 = true
			t.Logf("Found Test Chain 1: %s (%s)", chain.ChainName, chain.ID)
		}
		if chain.ID == chain2ID {
			foundChain2 = true
			t.Logf("Found Test Chain 2: %s (%s)", chain.ChainName, chain.ID)
		}
	}

	if !foundChain1 {
		t.Error("Test Chain 1 not found in API response")
	}
	if !foundChain2 {
		t.Error("Test Chain 2 not found in API response")
	}

	if chainsResponse.Pagination != nil {
		t.Logf("Pagination: page=%d, limit=%d, total=%d, pages=%d",
			chainsResponse.Pagination.Page,
			chainsResponse.Pagination.Limit,
			chainsResponse.Pagination.Total,
			chainsResponse.Pagination.Pages)
	}
}

// TestGetTemplates tests fetching the list of templates
func TestGetTemplates(t *testing.T) {
	var template1ID, template2ID uuid.UUID

	// Setup: Create test templates using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()

		template1, err := fixtures.DefaultChainTemplate().
			WithName(fmt.Sprintf("Test Template 1 %d", timestamp)).
			WithCategory("defi").
			Create(context.Background(), db)
		require.NoError(t, err)
		template1ID = template1.ID

		template2, err := fixtures.DefaultChainTemplate().
			WithName(fmt.Sprintf("Test Template 2 %d", timestamp)).
			WithCategory("gaming").
			Create(context.Background(), db)
		require.NoError(t, err)
		template2ID = template2.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_templates WHERE id IN ($1, $2)",
				template1ID, template2ID)
		})

		t.Logf("Created test templates: %s, %s", template1ID, template2ID)
	})

	client := testutils.NewTestClient()

	templatesPath := testutils.GetAPIPath("/templates")
	resp, body := client.Get(t, templatesPath)

	testutils.AssertStatusOK(t, resp)

	var templatesResponse struct {
		Data       []models.ChainTemplate `json:"data"`
		Pagination *struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
			Total int `json:"total"`
			Pages int `json:"pages"`
		} `json:"pagination,omitempty"`
	}
	testutils.UnmarshalResponse(t, body, &templatesResponse)

	// Should have at least our 2 test templates
	if len(templatesResponse.Data) < 2 {
		t.Errorf("Expected at least 2 templates, got %d", len(templatesResponse.Data))
	}

	t.Logf("Found %d templates", len(templatesResponse.Data))

	// Verify our test templates are in the results
	foundTemplate1 := false
	foundTemplate2 := false
	for _, template := range templatesResponse.Data {
		if template.ID == template1ID {
			foundTemplate1 = true
			t.Logf("Found Test Template 1: %s (%s)", template.TemplateName, template.ID)
		}
		if template.ID == template2ID {
			foundTemplate2 = true
			t.Logf("Found Test Template 2: %s (%s)", template.TemplateName, template.ID)
		}
	}

	if !foundTemplate1 {
		t.Error("Test Template 1 not found in API response")
	}
	if !foundTemplate2 {
		t.Error("Test Template 2 not found in API response")
	}
}

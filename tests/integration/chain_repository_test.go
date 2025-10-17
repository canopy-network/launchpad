//go:build integration

package integration_test

import (
	"context"
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

// TestGetChainRepository tests fetching a GitHub repository for a chain
func TestGetChainRepository(t *testing.T) {
	var chainID, repoID uuid.UUID

	// Setup: Create test chain and repository using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Repo Test Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("REPO").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test repository
		repoURL := "https://github.com/test-org/test-chain-repo"
		repo, err := fixtures.DefaultChainRepository(chainID).
			WithGithubURL(repoURL).
			WithRepositoryName("test-chain-repo").
			WithRepositoryOwner("test-org").
			WithBuildStatus(models.BuildStatusSuccess).
			Create(context.Background(), db)
		require.NoError(t, err)
		repoID = repo.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_repositories WHERE id = $1", repoID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s with repository: %s", chainID, repoID)
	})

	client := testutils.NewTestClient()

	// Test: Fetch the repository
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))
	resp, body := client.Get(t, repoPath)

	testutils.AssertStatusOK(t, resp)

	var repoResponse struct {
		Data models.ChainRepository `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &repoResponse)

	repository := repoResponse.Data

	// Validate repository fields
	if repository.ID != repoID {
		t.Errorf("Expected repo ID=%s, got %s", repoID, repository.ID)
	}

	if repository.ChainID != chainID {
		t.Errorf("Expected chain_id=%s, got %s", chainID, repository.ChainID)
	}

	if repository.GithubURL != "https://github.com/test-org/test-chain-repo" {
		t.Errorf("Expected github_url='https://github.com/test-org/test-chain-repo', got '%s'", repository.GithubURL)
	}

	if repository.RepositoryName != "test-chain-repo" {
		t.Errorf("Expected repository_name='test-chain-repo', got '%s'", repository.RepositoryName)
	}

	if repository.RepositoryOwner != "test-org" {
		t.Errorf("Expected repository_owner='test-org', got '%s'", repository.RepositoryOwner)
	}

	if repository.DefaultBranch != "main" {
		t.Errorf("Expected default_branch='main', got '%s'", repository.DefaultBranch)
	}

	if repository.BuildStatus != models.BuildStatusSuccess {
		t.Errorf("Expected build_status='success', got '%s'", repository.BuildStatus)
	}

	if repository.AutoUpgradeEnabled != true {
		t.Error("Expected auto_upgrade_enabled=true")
	}

	if repository.UpgradeTrigger != models.UpgradeTriggerTagRelease {
		t.Errorf("Expected upgrade_trigger='tag_release', got '%s'", repository.UpgradeTrigger)
	}

	t.Logf("Successfully fetched repository: %s for chain: %s", repository.ID, chainID)
}

// TestGetChainRepositoryNotFound tests fetching a repository for a chain that doesn't have one
func TestGetChainRepositoryNotFound(t *testing.T) {
	var chainID uuid.UUID

	// Setup: Create test chain WITHOUT repository
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("No Repo Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("NORP").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain without repository: %s", chainID)
	})

	client := testutils.NewTestClient()

	// Test: Try to fetch the repository (should fail)
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))
	resp, body := client.Get(t, repoPath)

	// Should return 404 or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status, got 200 OK. Body: %s", string(body))
	}

	t.Logf("Correctly returned error status %d for chain without repository", resp.StatusCode)
}

// TestGetChainRepositoryUnauthorized tests fetching a repository for a chain owned by another user
func TestGetChainRepositoryUnauthorized(t *testing.T) {
	var chainID, repoID, otherUserID uuid.UUID

	// Setup: Create test chain and repository owned by a different user
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()

		// Create a different user
		otherUser, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("other-%d@example.com", timestamp)).
			WithUsername(fmt.Sprintf("other_user_%d", timestamp)).
			Create(context.Background(), db)
		require.NoError(t, err)
		otherUserID = otherUser.ID

		// Create test chain owned by other user
		chainFixture := fixtures.DefaultChain(otherUserID)
		chainFixture.ChainName = fmt.Sprintf("Other User Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("OTHR").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test repository
		repo, err := fixtures.DefaultChainRepository(chainID).
			WithGithubURL("https://github.com/other-org/other-repo").
			WithRepositoryName("other-repo").
			WithRepositoryOwner("other-org").
			Create(context.Background(), db)
		require.NoError(t, err)
		repoID = repo.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_repositories WHERE id = $1", repoID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", otherUserID)
		})

		t.Logf("Created test chain owned by other user: %s", chainID)
	})

	// Use the default test user (NOT the owner)
	client := testutils.NewTestClient()

	// Test: Try to fetch the repository (should fail due to authorization)
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))
	resp, body := client.Get(t, repoPath)

	// Should return 403 Forbidden or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status for unauthorized access, got 200 OK. Body: %s", string(body))
	}

	t.Logf("Correctly returned error status %d for unauthorized repository access", resp.StatusCode)
}

// TestGetChainRepositoryInvalidChainID tests fetching a repository with an invalid chain ID
func TestGetChainRepositoryInvalidChainID(t *testing.T) {
	client := testutils.NewTestClient()

	// Test with invalid UUID format
	repoPath := testutils.GetAPIPath("/chains/not-a-uuid/repository")
	resp, _ := client.Get(t, repoPath)

	// Should return 400 Bad Request or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status for invalid chain ID, got 200 OK")
	}

	t.Logf("Correctly returned error status %d for invalid chain ID format", resp.StatusCode)
}

// TestGetChainRepositoryNonExistentChain tests fetching a repository for a non-existent chain
func TestGetChainRepositoryNonExistentChain(t *testing.T) {
	client := testutils.NewTestClient()

	// Use a valid UUID that doesn't exist
	nonExistentID := uuid.New()
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", nonExistentID))
	resp, _ := client.Get(t, repoPath)

	// Should return 404 Not Found or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status for non-existent chain, got 200 OK")
	}

	t.Logf("Correctly returned error status %d for non-existent chain", resp.StatusCode)
}

// TestUpdateChainRepository tests updating a chain repository with partial updates
func TestUpdateChainRepository(t *testing.T) {
	var chainID, repoID uuid.UUID

	// Setup: Create test chain and repository using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Update Repo Test Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("UPDT").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test repository
		repoURL := "https://github.com/original-owner/original-repo"
		repo, err := fixtures.DefaultChainRepository(chainID).
			WithGithubURL(repoURL).
			WithRepositoryName("original-repo").
			WithRepositoryOwner("original-owner").
			WithDefaultBranch("main").
			Create(context.Background(), db)
		require.NoError(t, err)
		repoID = repo.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_repositories WHERE id = $1", repoID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s with repository: %s", chainID, repoID)
	})

	client := testutils.NewTestClient()

	// Test: Update the repository with partial updates
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))

	newURL := "https://github.com/updated-owner/updated-repo"
	newOwner := "updated-owner"
	newName := "updated-repo"
	newBranch := "develop"

	updatePayload := map[string]interface{}{
		"github_url":       newURL,
		"repository_owner": newOwner,
		"repository_name":  newName,
		"default_branch":   newBranch,
	}

	resp, body := client.Put(t, repoPath, updatePayload)

	testutils.AssertStatusOK(t, resp)

	var repoResponse struct {
		Data models.ChainRepository `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &repoResponse)

	repository := repoResponse.Data

	// Validate updated fields
	if repository.ID != repoID {
		t.Errorf("Expected repo ID=%s, got %s", repoID, repository.ID)
	}

	if repository.GithubURL != newURL {
		t.Errorf("Expected github_url='%s', got '%s'", newURL, repository.GithubURL)
	}

	if repository.RepositoryOwner != newOwner {
		t.Errorf("Expected repository_owner='%s', got '%s'", newOwner, repository.RepositoryOwner)
	}

	if repository.RepositoryName != newName {
		t.Errorf("Expected repository_name='%s', got '%s'", newName, repository.RepositoryName)
	}

	if repository.DefaultBranch != newBranch {
		t.Errorf("Expected default_branch='%s', got '%s'", newBranch, repository.DefaultBranch)
	}

	t.Logf("Successfully updated repository: %s for chain: %s", repository.ID, chainID)
}

// TestUpdateChainRepositoryPartial tests partial updates (only some fields)
func TestUpdateChainRepositoryPartial(t *testing.T) {
	var chainID, repoID uuid.UUID

	// Setup: Create test chain and repository using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Partial Update Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("PART").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test repository
		repo, err := fixtures.DefaultChainRepository(chainID).
			WithGithubURL("https://github.com/test-org/test-repo").
			WithRepositoryName("test-repo").
			WithRepositoryOwner("test-org").
			WithDefaultBranch("main").
			Create(context.Background(), db)
		require.NoError(t, err)
		repoID = repo.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_repositories WHERE id = $1", repoID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s with repository: %s", chainID, repoID)
	})

	client := testutils.NewTestClient()

	// Test: Update only the default_branch field
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))

	newBranch := "staging"
	updatePayload := map[string]interface{}{
		"default_branch": newBranch,
	}

	resp, body := client.Put(t, repoPath, updatePayload)

	testutils.AssertStatusOK(t, resp)

	var repoResponse struct {
		Data models.ChainRepository `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &repoResponse)

	repository := repoResponse.Data

	// Validate that only default_branch was updated
	if repository.DefaultBranch != newBranch {
		t.Errorf("Expected default_branch='%s', got '%s'", newBranch, repository.DefaultBranch)
	}

	// Original values should remain unchanged
	if repository.RepositoryOwner != "test-org" {
		t.Errorf("Expected repository_owner='test-org' (unchanged), got '%s'", repository.RepositoryOwner)
	}

	if repository.RepositoryName != "test-repo" {
		t.Errorf("Expected repository_name='test-repo' (unchanged), got '%s'", repository.RepositoryName)
	}

	if repository.GithubURL != "https://github.com/test-org/test-repo" {
		t.Errorf("Expected github_url unchanged, got '%s'", repository.GithubURL)
	}

	t.Logf("Successfully performed partial update on repository: %s", repository.ID)
}

// TestUpdateChainRepositoryUnauthorized tests updating a repository for a chain owned by another user
func TestUpdateChainRepositoryUnauthorized(t *testing.T) {
	var chainID, repoID, otherUserID uuid.UUID

	// Setup: Create test chain and repository owned by a different user
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()

		// Create a different user
		otherUser, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("other-update-%d@example.com", timestamp)).
			WithUsername(fmt.Sprintf("other_update_user_%d", timestamp)).
			Create(context.Background(), db)
		require.NoError(t, err)
		otherUserID = otherUser.ID

		// Create test chain owned by other user
		chainFixture := fixtures.DefaultChain(otherUserID)
		chainFixture.ChainName = fmt.Sprintf("Other User Update Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("OUUC").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test repository
		repo, err := fixtures.DefaultChainRepository(chainID).
			WithGithubURL("https://github.com/other-org/other-repo").
			WithRepositoryName("other-repo").
			WithRepositoryOwner("other-org").
			Create(context.Background(), db)
		require.NoError(t, err)
		repoID = repo.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_repositories WHERE id = $1", repoID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", otherUserID)
		})

		t.Logf("Created test chain owned by other user: %s", chainID)
	})

	// Use the default test user (NOT the owner)
	client := testutils.NewTestClient()

	// Test: Try to update the repository (should fail due to authorization)
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))

	updatePayload := map[string]interface{}{
		"default_branch": "hacker-branch",
	}

	resp, body := client.Put(t, repoPath, updatePayload)

	// Should return 403 Forbidden or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status for unauthorized access, got 200 OK. Body: %s", string(body))
	}

	t.Logf("Correctly returned error status %d for unauthorized repository update", resp.StatusCode)
}

// TestUpdateChainRepositoryNotFound tests updating a repository that doesn't exist
func TestUpdateChainRepositoryNotFound(t *testing.T) {
	var chainID uuid.UUID

	// Setup: Create test chain WITHOUT repository
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("No Repo Update Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("NORU").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain without repository: %s", chainID)
	})

	client := testutils.NewTestClient()

	// Test: Try to update the repository (should fail)
	repoPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/repository", chainID))

	updatePayload := map[string]interface{}{
		"default_branch": "new-branch",
	}

	resp, body := client.Put(t, repoPath, updatePayload)

	// Should return 404 or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status, got 200 OK. Body: %s", string(body))
	}

	t.Logf("Correctly returned error status %d for chain without repository", resp.StatusCode)
}

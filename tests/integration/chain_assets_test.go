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

// TestGetChainAssets tests fetching all assets for a chain
func TestGetChainAssets(t *testing.T) {
	var chainID, assetID1, assetID2 uuid.UUID

	// Setup: Create test chain and assets using fixtures
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Asset Test Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("ASST").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test assets
		asset1, err := fixtures.DefaultChainAsset(chainID, creatorID).
			WithAssetType(models.AssetTypeLogo).
			WithFileName("logo.png").
			WithFileURL("https://cdn.example.com/logo.png").
			WithDisplayOrder(0).
			Create(context.Background(), db)
		require.NoError(t, err)
		assetID1 = asset1.ID

		asset2, err := fixtures.DefaultChainAsset(chainID, creatorID).
			WithAssetType(models.AssetTypeWhitepaper).
			WithFileName("whitepaper.pdf").
			WithFileURL("https://cdn.example.com/whitepaper.pdf").
			WithDisplayOrder(1).
			Create(context.Background(), db)
		require.NoError(t, err)
		assetID2 = asset2.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_assets WHERE id IN ($1, $2)", assetID1, assetID2)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s with %d assets", chainID, 2)
	})

	client := testutils.NewTestClient()

	// Test: Fetch the assets
	assetsPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets", chainID))
	resp, body := client.Get(t, assetsPath)

	testutils.AssertStatusOK(t, resp)

	var assetsResponse struct {
		Data []models.ChainAsset `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &assetsResponse)

	assets := assetsResponse.Data

	// Validate we got 2 assets
	if len(assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(assets))
	}

	// Validate assets are sorted by display_order
	if len(assets) >= 2 {
		if assets[0].DisplayOrder > assets[1].DisplayOrder {
			t.Error("Expected assets to be sorted by display_order")
		}

		// Validate first asset (logo)
		if assets[0].AssetType != models.AssetTypeLogo {
			t.Errorf("Expected first asset type='logo', got '%s'", assets[0].AssetType)
		}

		// Validate second asset (whitepaper)
		if assets[1].AssetType != models.AssetTypeWhitepaper {
			t.Errorf("Expected second asset type='whitepaper', got '%s'", assets[1].AssetType)
		}
	}

	t.Logf("Successfully fetched %d assets for chain: %s", len(assets), chainID)
}

// TestGetChainAssetsEmpty tests fetching assets for a chain with no assets
func TestGetChainAssetsEmpty(t *testing.T) {
	var chainID uuid.UUID

	// Setup: Create test chain WITHOUT assets
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("No Assets Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("NOAS").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain without assets: %s", chainID)
	})

	client := testutils.NewTestClient()

	// Test: Fetch the assets (should return empty array)
	assetsPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets", chainID))
	resp, body := client.Get(t, assetsPath)

	testutils.AssertStatusOK(t, resp)

	var assetsResponse struct {
		Data []models.ChainAsset `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &assetsResponse)

	assets := assetsResponse.Data

	// Should return empty array, not null
	if assets == nil {
		t.Error("Expected empty array, got null")
	}

	if len(assets) != 0 {
		t.Errorf("Expected 0 assets, got %d", len(assets))
	}

	t.Logf("Correctly returned empty array for chain without assets")
}

// TestCreateChainAsset tests creating a new asset for a chain
func TestCreateChainAsset(t *testing.T) {
	var chainID uuid.UUID

	// Setup: Create test chain
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Create Asset Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("CRAS").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_assets WHERE chain_id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s", chainID)
	})

	client := testutils.NewTestClient()

	// Test: Create a new asset
	assetsPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets", chainID))

	sizeBytes := int64(54321)
	isPrimary := true
	isFeatured := false

	createPayload := map[string]interface{}{
		"asset_type":      "banner",
		"file_name":       "chain-banner.jpg",
		"file_url":        "https://cdn.example.com/banner.jpg",
		"file_size_bytes": sizeBytes,
		"mime_type":       "image/jpeg",
		"title":           "Chain Banner",
		"description":     "Official chain banner image",
		"alt_text":        "Banner showing chain logo",
		"display_order":   5,
		"is_primary":      isPrimary,
		"is_featured":     isFeatured,
	}

	resp, body := client.Post(t, assetsPath, createPayload)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var assetResponse struct {
		Data models.ChainAsset `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &assetResponse)

	asset := assetResponse.Data

	// Validate created asset fields
	if asset.ChainID != chainID {
		t.Errorf("Expected chain_id=%s, got %s", chainID, asset.ChainID)
	}

	if asset.AssetType != "banner" {
		t.Errorf("Expected asset_type='banner', got '%s'", asset.AssetType)
	}

	if asset.FileName != "chain-banner.jpg" {
		t.Errorf("Expected file_name='chain-banner.jpg', got '%s'", asset.FileName)
	}

	if asset.FileURL != "https://cdn.example.com/banner.jpg" {
		t.Errorf("Expected file_url='https://cdn.example.com/banner.jpg', got '%s'", asset.FileURL)
	}

	if asset.ModerationStatus != "pending" {
		t.Errorf("Expected moderation_status='pending', got '%s'", asset.ModerationStatus)
	}

	if asset.IsActive != true {
		t.Error("Expected is_active=true")
	}

	t.Logf("Successfully created asset: %s for chain: %s", asset.ID, chainID)
}

// TestUpdateChainAsset tests updating an existing asset
func TestUpdateChainAsset(t *testing.T) {
	var chainID, assetID uuid.UUID

	// Setup: Create test chain and asset
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Update Asset Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("UPAS").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test asset
		asset, err := fixtures.DefaultChainAsset(chainID, creatorID).
			WithAssetType(models.AssetTypeLogo).
			WithFileName("old-logo.png").
			WithFileURL("https://cdn.example.com/old-logo.png").
			WithModerationStatus("pending").
			Create(context.Background(), db)
		require.NoError(t, err)
		assetID = asset.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_assets WHERE id = $1", assetID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s with asset: %s", chainID, assetID)
	})

	client := testutils.NewTestClient()

	// Test: Update the asset
	assetPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets/%s", chainID, assetID))

	newFileName := "new-logo.png"
	newFileURL := "https://cdn.example.com/new-logo.png"
	newModerationStatus := "approved"
	newDisplayOrder := 10

	updatePayload := map[string]interface{}{
		"file_name":         newFileName,
		"file_url":          newFileURL,
		"moderation_status": newModerationStatus,
		"display_order":     newDisplayOrder,
	}

	resp, body := client.Put(t, assetPath, updatePayload)

	testutils.AssertStatusOK(t, resp)

	var assetResponse struct {
		Data models.ChainAsset `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &assetResponse)

	asset := assetResponse.Data

	// Validate updated fields
	if asset.ID != assetID {
		t.Errorf("Expected asset ID=%s, got %s", assetID, asset.ID)
	}

	if asset.FileName != newFileName {
		t.Errorf("Expected file_name='%s', got '%s'", newFileName, asset.FileName)
	}

	if asset.FileURL != newFileURL {
		t.Errorf("Expected file_url='%s', got '%s'", newFileURL, asset.FileURL)
	}

	if asset.ModerationStatus != newModerationStatus {
		t.Errorf("Expected moderation_status='%s', got '%s'", newModerationStatus, asset.ModerationStatus)
	}

	if asset.DisplayOrder != newDisplayOrder {
		t.Errorf("Expected display_order=%d, got %d", newDisplayOrder, asset.DisplayOrder)
	}

	// Validate unchanged fields
	if asset.AssetType != models.AssetTypeLogo {
		t.Errorf("Expected asset_type='logo' (unchanged), got '%s'", asset.AssetType)
	}

	t.Logf("Successfully updated asset: %s for chain: %s", asset.ID, chainID)
}

// TestUpdateChainAssetPartial tests partial update (only some fields)
func TestUpdateChainAssetPartial(t *testing.T) {
	var chainID, assetID uuid.UUID

	// Setup: Create test chain and asset
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

		// Create test asset
		asset, err := fixtures.DefaultChainAsset(chainID, creatorID).
			WithAssetType(models.AssetTypeWhitepaper).
			WithFileName("whitepaper-v1.pdf").
			WithFileURL("https://cdn.example.com/wp-v1.pdf").
			WithDisplayOrder(5).
			Create(context.Background(), db)
		require.NoError(t, err)
		assetID = asset.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_assets WHERE id = $1", assetID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s with asset: %s", chainID, assetID)
	})

	client := testutils.NewTestClient()

	// Test: Update only the display_order field
	assetPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets/%s", chainID, assetID))

	newDisplayOrder := 1
	updatePayload := map[string]interface{}{
		"display_order": newDisplayOrder,
	}

	resp, body := client.Put(t, assetPath, updatePayload)

	testutils.AssertStatusOK(t, resp)

	var assetResponse struct {
		Data models.ChainAsset `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &assetResponse)

	asset := assetResponse.Data

	// Validate that only display_order was updated
	if asset.DisplayOrder != newDisplayOrder {
		t.Errorf("Expected display_order=%d, got %d", newDisplayOrder, asset.DisplayOrder)
	}

	// Original values should remain unchanged
	if asset.FileName != "whitepaper-v1.pdf" {
		t.Errorf("Expected file_name='whitepaper-v1.pdf' (unchanged), got '%s'", asset.FileName)
	}

	if asset.FileURL != "https://cdn.example.com/wp-v1.pdf" {
		t.Errorf("Expected file_url unchanged, got '%s'", asset.FileURL)
	}

	if asset.AssetType != models.AssetTypeWhitepaper {
		t.Errorf("Expected asset_type='whitepaper' (unchanged), got '%s'", asset.AssetType)
	}

	t.Logf("Successfully performed partial update on asset: %s", asset.ID)
}

// TestGetChainAssetsUnauthorized tests fetching assets for a chain owned by another user
func TestGetChainAssetsUnauthorized(t *testing.T) {
	var chainID, assetID, otherUserID uuid.UUID

	// Setup: Create test chain and asset owned by a different user
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()

		// Create a different user
		otherUser, err := fixtures.DefaultUser().
			WithEmail(fmt.Sprintf("other-assets-%d@example.com", timestamp)).
			WithUsername(fmt.Sprintf("other_assets_user_%d", timestamp)).
			Create(context.Background(), db)
		require.NoError(t, err)
		otherUserID = otherUser.ID

		// Create test chain owned by other user
		chainFixture := fixtures.DefaultChain(otherUserID)
		chainFixture.ChainName = fmt.Sprintf("Other User Assets Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("OUAC").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Create test asset
		asset, err := fixtures.DefaultChainAsset(chainID, otherUserID).
			Create(context.Background(), db)
		require.NoError(t, err)
		assetID = asset.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chain_assets WHERE id = $1", assetID)
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
			db.ExecContext(context.Background(),
				"DELETE FROM users WHERE id = $1", otherUserID)
		})

		t.Logf("Created test chain owned by other user: %s", chainID)
	})

	// Use the default test user (NOT the owner)
	client := testutils.NewTestClient()

	// Test: Try to fetch the assets (should fail due to authorization)
	assetsPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets", chainID))
	resp, body := client.Get(t, assetsPath)

	// Should return 403 Forbidden or similar error status
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status for unauthorized access, got 200 OK. Body: %s", string(body))
	}

	t.Logf("Correctly returned error status %d for unauthorized assets access", resp.StatusCode)
}

// TestCreateChainAssetInvalidType tests creating an asset with an invalid asset type
func TestCreateChainAssetInvalidType(t *testing.T) {
	var chainID uuid.UUID

	// Setup: Create test chain
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Invalid Type Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("INVT").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s", chainID)
	})

	client := testutils.NewTestClient()

	// Test: Try to create an asset with invalid type
	assetsPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets", chainID))

	createPayload := map[string]interface{}{
		"asset_type": "invalid_type",
		"file_name":  "test.jpg",
		"file_url":   "https://cdn.example.com/test.jpg",
	}

	resp, body := client.Post(t, assetsPath, createPayload)

	// Should return 400 Bad Request or 422 Unprocessable Entity
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		t.Errorf("Expected error status for invalid asset type, got %d. Body: %s", resp.StatusCode, string(body))
	}

	t.Logf("Correctly returned error status %d for invalid asset type", resp.StatusCode)
}

// TestUpdateChainAssetNotFound tests updating an asset that doesn't exist
func TestUpdateChainAssetNotFound(t *testing.T) {
	var chainID uuid.UUID

	// Setup: Create test chain
	testutils.WithTestDB(t, func(db *sqlx.DB) {
		timestamp := time.Now().UnixNano()
		creatorID := uuid.MustParse(testutils.TestUserID)

		// Create test chain
		chainFixture := fixtures.DefaultChain(creatorID)
		chainFixture.ChainName = fmt.Sprintf("Asset Not Found Chain %d", timestamp)
		chain, err := chainFixture.WithTokenSymbol("ANFC").
			Create(context.Background(), db)
		require.NoError(t, err)
		chainID = chain.ID

		// Cleanup after test
		t.Cleanup(func() {
			db.ExecContext(context.Background(),
				"DELETE FROM chains WHERE id = $1", chainID)
		})

		t.Logf("Created test chain: %s", chainID)
	})

	client := testutils.NewTestClient()

	// Test: Try to update a non-existent asset
	nonExistentAssetID := uuid.New()
	assetPath := testutils.GetAPIPath(fmt.Sprintf("/chains/%s/assets/%s", chainID, nonExistentAssetID))

	updatePayload := map[string]interface{}{
		"file_name": "updated.png",
	}

	resp, body := client.Put(t, assetPath, updatePayload)

	// Should return 404 Not Found
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected error status for non-existent asset, got 200 OK. Body: %s", string(body))
	}

	t.Logf("Correctly returned error status %d for non-existent asset", resp.StatusCode)
}

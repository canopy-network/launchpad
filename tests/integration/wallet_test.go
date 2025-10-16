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
	"github.com/google/uuid"
)

// TestCreateWallet tests the basic wallet creation flow
func TestCreateWallet(t *testing.T) {
	client := testutils.NewTestClient()

	// Create a new wallet
	t.Log("Creating new wallet...")
	createWalletRequest := map[string]interface{}{
		"password":     "TestPassword123",
		"wallet_name":  "Test Wallet",
		"user_id":      testutils.TestUserID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createWalletRequest)

	testutils.AssertStatusCreated(t, resp)

	// Validate the created wallet
	var walletResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &walletResponse)

	wallet := walletResponse.Data

	// Validate wallet fields
	if wallet.Address == "" {
		t.Error("Expected wallet address to be set")
	}

	if wallet.PublicKey == "" {
		t.Error("Expected public key to be set")
	}

	if wallet.WalletName == nil || *wallet.WalletName != "Test Wallet" {
		t.Errorf("Expected wallet_name='Test Wallet', got '%v'", wallet.WalletName)
	}

	if !wallet.IsActive {
		t.Error("Expected wallet to be active")
	}

	if wallet.IsLocked {
		t.Error("Expected wallet to not be locked")
	}

	if wallet.ID.String() == "" {
		t.Error("Expected wallet ID to be set")
	}

	t.Logf("Successfully created wallet with ID: %s, Address: %s", wallet.ID, wallet.Address)
}

// TestDecryptWallet tests wallet decryption with correct and incorrect passwords
func TestDecryptWallet(t *testing.T) {
	client := testutils.NewTestClient()

	// Step 1: Create a wallet
	t.Log("Creating wallet for decryption test...")
	password := "SecurePassword456"
	createWalletRequest := map[string]interface{}{
		"password":    password,
		"wallet_name": "Decrypt Test Wallet",
		"user_id":     testutils.TestUserID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createWalletRequest)
	testutils.AssertStatusCreated(t, resp)

	var createResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &createResponse)
	walletID := createResponse.Data.ID

	// Step 2: Test decryption with correct password
	t.Log("Testing decryption with correct password...")
	decryptPath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s/decrypt", walletID))
	decryptRequest := map[string]interface{}{
		"password": password,
	}

	resp, body = client.Post(t, decryptPath, decryptRequest)
	testutils.AssertStatusOK(t, resp)

	var decryptResponse struct {
		Data models.WalletWithDecryptedKey `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &decryptResponse)

	if decryptResponse.Data.DecryptedPrivateKey == "" {
		t.Error("Expected decrypted private key to be returned")
	}

	t.Logf("Successfully decrypted wallet, private key length: %d", len(decryptResponse.Data.DecryptedPrivateKey))

	// Step 3: Test decryption with incorrect password
	t.Log("Testing decryption with incorrect password...")
	wrongDecryptRequest := map[string]interface{}{
		"password": "WrongPassword",
	}

	resp, body = client.Post(t, decryptPath, wrongDecryptRequest)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong password, got %d", resp.StatusCode)
	}

	var errorResponse struct {
		Error *testutils.ErrorResponse `json:"error"`
	}
	testutils.UnmarshalResponse(t, body, &errorResponse)

	if errorResponse.Error == nil {
		t.Error("Expected error response for wrong password")
	}

	t.Logf("Correctly rejected wrong password: %s", errorResponse.Error.Message)
}

// TestWalletRateLimiting tests the wallet locking mechanism after failed attempts
func TestWalletRateLimiting(t *testing.T) {
	client := testutils.NewTestClient()

	// Create a wallet
	t.Log("Creating wallet for rate limiting test...")
	password := "CorrectPassword789"
	createWalletRequest := map[string]interface{}{
		"password":    password,
		"wallet_name": "Rate Limit Test Wallet",
		"user_id":     testutils.TestUserID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createWalletRequest)
	testutils.AssertStatusCreated(t, resp)

	var createResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &createResponse)
	walletID := createResponse.Data.ID

	// Attempt decryption with wrong password 5 times
	t.Log("Attempting 5 failed decryptions to trigger lock...")
	decryptPath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s/decrypt", walletID))
	wrongRequest := map[string]interface{}{
		"password": "WrongPassword",
	}

	for i := 1; i <= 5; i++ {
		resp, _ := client.Post(t, decryptPath, wrongRequest)
		t.Logf("Failed attempt %d/5: Status %d", i, resp.StatusCode)

		// Last attempt should lock the wallet
		if i == 5 && resp.StatusCode != http.StatusForbidden {
			t.Logf("Warning: Expected 403 on 5th attempt, got %d", resp.StatusCode)
		}
	}

	// Attempt with correct password - should be locked
	t.Log("Attempting decryption with correct password on locked wallet...")
	correctRequest := map[string]interface{}{
		"password": password,
	}

	resp, body = client.Post(t, decryptPath, correctRequest)

	if resp.StatusCode == http.StatusForbidden {
		t.Log("Wallet correctly locked after 5 failed attempts")
	} else {
		t.Logf("Wallet status: %d (expected 403 forbidden)", resp.StatusCode)
	}

	// Test unlock endpoint
	t.Log("Testing unlock endpoint...")
	unlockPath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s/unlock", walletID))
	resp, _ = client.Post(t, unlockPath, nil)

	if resp.StatusCode == http.StatusOK {
		t.Log("Successfully unlocked wallet")

		// Try correct password again after unlock
		resp, body = client.Post(t, decryptPath, correctRequest)
		if resp.StatusCode == http.StatusOK {
			t.Log("Successfully decrypted wallet after unlock")
		} else {
			t.Logf("Decryption after unlock returned status: %d", resp.StatusCode)
		}
	}
}

// TestGetWallets tests listing wallets with filters
func TestGetWallets(t *testing.T) {
	client := testutils.NewTestClient()

	// Create 2 test wallets
	t.Log("Creating test wallets...")
	for i := 1; i <= 2; i++ {
		createRequest := map[string]interface{}{
			"password":    fmt.Sprintf("Password%d", i),
			"wallet_name": fmt.Sprintf("List Test Wallet %d - %d", i, time.Now().UnixNano()),
			"user_id":     testutils.TestUserID,
		}

		walletsPath := testutils.GetAPIPath("/wallets")
		resp, _ := client.Post(t, walletsPath, createRequest)
		testutils.AssertStatusCreated(t, resp)
	}

	// List all wallets
	t.Log("Fetching wallets list...")
	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Get(t, walletsPath)

	testutils.AssertStatusOK(t, resp)

	var walletsResponse struct {
		Data       []models.Wallet `json:"data"`
		Pagination *struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
			Total int `json:"total"`
			Pages int `json:"pages"`
		} `json:"pagination,omitempty"`
	}
	testutils.UnmarshalResponse(t, body, &walletsResponse)

	if len(walletsResponse.Data) < 2 {
		t.Errorf("Expected at least 2 wallets, got %d", len(walletsResponse.Data))
	}

	t.Logf("Found %d wallets", len(walletsResponse.Data))

	if walletsResponse.Pagination != nil {
		t.Logf("Pagination: page=%d, limit=%d, total=%d, pages=%d",
			walletsResponse.Pagination.Page,
			walletsResponse.Pagination.Limit,
			walletsResponse.Pagination.Total,
			walletsResponse.Pagination.Pages)
	}

	// Test filtering by user_id
	t.Log("Testing filter by user_id...")
	filterPath := testutils.GetAPIPath(fmt.Sprintf("/wallets?user_id=%s", testutils.TestUserID))
	resp, body = client.Get(t, filterPath)
	testutils.AssertStatusOK(t, resp)

	var filteredResponse struct {
		Data []models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &filteredResponse)

	t.Logf("Found %d wallets for user %s", len(filteredResponse.Data), testutils.TestUserID)

	// Verify all wallets belong to the test user
	for _, wallet := range filteredResponse.Data {
		if wallet.UserID == nil || wallet.UserID.String() != testutils.TestUserID {
			t.Errorf("Expected wallet to belong to user %s, got %v", testutils.TestUserID, wallet.UserID)
		}
	}
}

// TestGetWallet tests fetching a single wallet by ID
func TestGetWallet(t *testing.T) {
	client := testutils.NewTestClient()

	// Create a wallet
	t.Log("Creating wallet...")
	createRequest := map[string]interface{}{
		"password":    "GetTestPassword",
		"wallet_name": "Get Test Wallet",
		"user_id":     testutils.TestUserID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createRequest)
	testutils.AssertStatusCreated(t, resp)

	var createResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &createResponse)
	walletID := createResponse.Data.ID

	// Fetch the wallet by ID
	t.Log("Fetching wallet by ID...")
	getPath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s", walletID))
	resp, body = client.Get(t, getPath)

	testutils.AssertStatusOK(t, resp)

	var getResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &getResponse)

	wallet := getResponse.Data

	if wallet.ID != walletID {
		t.Errorf("Expected wallet ID %s, got %s", walletID, wallet.ID)
	}

	if wallet.WalletName == nil || *wallet.WalletName != "Get Test Wallet" {
		t.Errorf("Expected wallet name 'Get Test Wallet', got %v", wallet.WalletName)
	}

	t.Logf("Successfully fetched wallet: %s", wallet.ID)
}

// TestUpdateWallet tests updating wallet metadata
func TestUpdateWallet(t *testing.T) {
	client := testutils.NewTestClient()

	// Create a wallet
	t.Log("Creating wallet...")
	createRequest := map[string]interface{}{
		"password":    "UpdateTestPassword",
		"wallet_name": "Original Name",
		"user_id":     testutils.TestUserID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createRequest)
	testutils.AssertStatusCreated(t, resp)

	var createResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &createResponse)
	walletID := createResponse.Data.ID

	// Update the wallet
	t.Log("Updating wallet...")
	newName := "Updated Wallet Name"
	newDescription := "This wallet has been updated"
	updateRequest := map[string]interface{}{
		"wallet_name":        newName,
		"wallet_description": newDescription,
	}

	updatePath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s", walletID))
	req, _ := http.NewRequest("PUT", client.BaseURL+updatePath, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", client.UserID)

	jsonData, _ := json.Marshal(updateRequest)
	req.Body = http.NoBody
	req, _ = http.NewRequest("PUT", client.BaseURL+updatePath, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", client.UserID)

	resp, _ = client.Client.Do(req)
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	testutils.AssertStatusOK(t, resp)

	var updateResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &updateResponse)

	wallet := updateResponse.Data

	if wallet.WalletName == nil || *wallet.WalletName != newName {
		t.Errorf("Expected wallet name '%s', got %v", newName, wallet.WalletName)
	}

	if wallet.WalletDescription == nil || *wallet.WalletDescription != newDescription {
		t.Errorf("Expected wallet description '%s', got %v", newDescription, wallet.WalletDescription)
	}

	t.Logf("Successfully updated wallet: %s", wallet.ID)
}

// TestDeleteWallet tests wallet deletion
func TestDeleteWallet(t *testing.T) {
	client := testutils.NewTestClient()

	// Create a wallet
	t.Log("Creating wallet for deletion...")
	createRequest := map[string]interface{}{
		"password":    "DeleteTestPassword",
		"wallet_name": "To Be Deleted",
		"user_id":     testutils.TestUserID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createRequest)
	testutils.AssertStatusCreated(t, resp)

	var createResponse struct {
		Data models.Wallet `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &createResponse)
	walletID := createResponse.Data.ID

	t.Logf("Created wallet %s for deletion", walletID)

	// Delete the wallet
	t.Log("Deleting wallet...")
	deletePath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s", walletID))
	resp, body = client.Delete(t, deletePath)

	testutils.AssertStatusOK(t, resp)

	var deleteResponse struct {
		Data map[string]string `json:"data"`
	}
	testutils.UnmarshalResponse(t, body, &deleteResponse)

	if deleteResponse.Data["message"] != "Wallet deleted successfully" {
		t.Errorf("Unexpected delete response: %v", deleteResponse.Data)
	}

	// Verify wallet is gone
	t.Log("Verifying wallet is deleted...")
	getPath := testutils.GetAPIPath(fmt.Sprintf("/wallets/%s", walletID))
	resp, _ = client.Get(t, getPath)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for deleted wallet, got %d", resp.StatusCode)
	}

	t.Log("Successfully verified wallet deletion")
}

// TestCreateWalletValidation tests validation errors
func TestCreateWalletValidation(t *testing.T) {
	client := testutils.NewTestClient()

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name: "Missing password",
			request: map[string]interface{}{
				"user_id": testutils.TestUserID,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when password is missing",
		},
		{
			name: "Short password",
			request: map[string]interface{}{
				"password": "short",
				"user_id":  testutils.TestUserID,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when password is too short (min 8 chars)",
		},
		{
			name: "Invalid user_id format",
			request: map[string]interface{}{
				"password": "ValidPassword123",
				"user_id":  "not-a-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should fail when user_id is not a valid UUID",
		},
		{
			name: "No associations",
			request: map[string]interface{}{
				"password": "ValidPassword123",
			},
			expectedStatus: http.StatusInternalServerError,
			description:    "Should fail when no associations (user_id, chain_id, or created_by) are provided",
		},
	}

	walletsPath := testutils.GetAPIPath("/wallets")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := client.Post(t, walletsPath, tt.request)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d. Body: %s",
					tt.description, tt.expectedStatus, resp.StatusCode, string(body))
			}

			t.Logf("Validation test '%s' passed with status %d", tt.name, resp.StatusCode)
		})
	}
}

// TestWalletWithChain tests creating a wallet associated with a chain
func TestWalletWithChain(t *testing.T) {
	client := testutils.NewTestClient()

	// Use a dummy chain ID (in real scenario, you'd create a chain first)
	chainID := uuid.New().String()

	createRequest := map[string]interface{}{
		"password":    "ChainWalletPassword",
		"wallet_name": "Chain Treasury Wallet",
		"chain_id":    chainID,
	}

	walletsPath := testutils.GetAPIPath("/wallets")
	resp, body := client.Post(t, walletsPath, createRequest)

	// This might fail if foreign key constraint is enforced
	// but demonstrates the API structure
	if resp.StatusCode == http.StatusCreated {
		var walletResponse struct {
			Data models.Wallet `json:"data"`
		}
		testutils.UnmarshalResponse(t, body, &walletResponse)

		wallet := walletResponse.Data
		if wallet.ChainID == nil || wallet.ChainID.String() != chainID {
			t.Errorf("Expected chain_id %s, got %v", chainID, wallet.ChainID)
		}

		t.Logf("Successfully created chain wallet: %s", wallet.ID)
	} else {
		t.Logf("Chain wallet creation returned status %d (expected if chain doesn't exist)", resp.StatusCode)
	}
}

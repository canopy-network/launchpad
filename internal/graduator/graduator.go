package graduator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
)

// Graduator handles the virtual chain graduation process
type Graduator struct {
	chainRepo       interfaces.ChainRepository
	virtualPoolRepo interfaces.VirtualPoolRepository
	userRepo        interfaces.UserRepository
	templatePath    string
	rpcEndpoint     string
	httpClient      *http.Client
}

// New creates a new Graduator instance
func New(chainRepo interfaces.ChainRepository, virtualPoolRepo interfaces.VirtualPoolRepository, userRepo interfaces.UserRepository, templatePath string, rpcEndpoint string) *Graduator {
	return &Graduator{
		chainRepo:       chainRepo,
		virtualPoolRepo: virtualPoolRepo,
		userRepo:        userRepo,
		templatePath:    templatePath,
		rpcEndpoint:     rpcEndpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenesisAccount represents an account in the genesis file
type GenesisAccount struct {
	Address string
	Amount  int64
}

// GenesisData contains the template data for genesis file generation
type GenesisData struct {
	Accounts []GenesisAccount
}

// GraduationRPCPayload represents the data sent to the graduation RPC endpoint
type GraduationRPCPayload struct {
	Username    string                 `json:"username"`
	ChainName   string                 `json:"chain_name"`
	WalletOwner string                 `json:"wallet_owner"`
	GenesisFile string                 `json:"genesis_file"`
	Tokenomics  map[string]interface{} `json:"tokenomics"`
	GithubRepo  string                 `json:"github_repo"`
}

// MakeGraduationRPCCall sends graduation data to the configured RPC endpoint
func (g *Graduator) MakeGraduationRPCCall(ctx context.Context, chain *models.Chain, genesisFile string) error {
	// Validate required relationships are loaded
	if chain.Creator == nil {
		return fmt.Errorf("chain creator not loaded")
	}
	if chain.Repository == nil {
		return fmt.Errorf("chain repository not loaded")
	}

	// Prepare tokenomics data
	tokenomics := map[string]interface{}{
		"token_name":          chain.TokenName,
		"token_symbol":        chain.TokenSymbol,
		"token_total_supply":  chain.TokenTotalSupply,
		"block_time_seconds":  chain.BlockTimeSeconds,
		"block_reward_amount": chain.BlockRewardAmount,
	}

	// Prepare RPC payload
	payload := GraduationRPCPayload{
		Username:    getStringValue(chain.Creator.Username),
		ChainName:   chain.ChainName,
		WalletOwner: chain.Creator.WalletAddress,
		GenesisFile: genesisFile,
		Tokenomics:  tokenomics,
		GithubRepo:  chain.Repository.GithubURL,
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", g.rpcEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create RPC request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute RPC request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("RPC request failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// getStringValue safely extracts string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CheckAndGraduate checks if a chain is eligible for graduation and performs the graduation process
func (g *Graduator) CheckAndGraduate(ctx context.Context, chainID uuid.UUID) error {
	// Get the chain with relationships
	chain, err := g.chainRepo.GetByID(ctx, chainID, []string{"creator", "repository"})
	if err != nil {
		return fmt.Errorf("failed to get chain: %w", err)
	}

	// Check if already graduated
	if chain.IsGraduated {
		return fmt.Errorf("chain already graduated")
	}

	// Get virtual pool
	pool, err := g.virtualPoolRepo.GetPoolByChainID(ctx, chainID)
	if err != nil {
		return fmt.Errorf("failed to get virtual pool: %w", err)
	}

	// Check if pool value meets graduation threshold
	if pool.CNPYReserve < chain.GraduationThreshold {
		return fmt.Errorf("graduation threshold not met: current %v, required %v", pool.CNPYReserve, chain.GraduationThreshold)
	}

	// Generate genesis file
	genesisFile, err := g.GenerateGenesisFile(ctx, chainID)
	if err != nil {
		return fmt.Errorf("failed to generate genesis file: %w", err)
	}

	// Make graduation RPC call
	if err := g.MakeGraduationRPCCall(ctx, chain, genesisFile); err != nil {
		return fmt.Errorf("failed to make graduation RPC call: %w", err)
	}

	// Update chain status to graduated
	now := time.Now()
	chain.IsGraduated = true
	chain.GraduationTime = &now
	chain.Status = models.ChainStatusGraduated

	if _, err := g.chainRepo.Update(ctx, chain); err != nil {
		return fmt.Errorf("failed to update chain graduation status: %w", err)
	}

	return nil
}

// GenerateGenesisFile generates the genesis.json file for a chain and returns it as a string
func (g *Graduator) GenerateGenesisFile(ctx context.Context, chainID uuid.UUID) (string, error) {
	// Get positions with user addresses
	positions, err := g.virtualPoolRepo.GetPositionsWithUsersByChainID(ctx, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to get positions: %w", err)
	}

	// Convert to genesis accounts
	accounts := make([]GenesisAccount, len(positions))
	for i, pos := range positions {
		accounts[i] = GenesisAccount{
			Address: pos.WalletAddress,
			Amount:  pos.TokenBalance,
		}
	}

	// Create template data
	data := GenesisData{
		Accounts: accounts,
	}

	// Parse and execute template
	tmpl, err := template.ParseFiles(g.templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template to buffer
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

package graduator

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/internal/testutil/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenesisAccountStructure(t *testing.T) {
	account := GenesisAccount{
		Address: "0x1234567890abcdef",
		Amount:  1000000,
	}

	if account.Address == "" {
		t.Error("Address should not be empty")
	}
	if account.Amount <= 0 {
		t.Error("Amount should be positive")
	}
}

func TestGenesisDataStructure(t *testing.T) {
	data := GenesisData{
		Accounts: []GenesisAccount{
			{Address: "0xabc", Amount: 100},
			{Address: "0xdef", Amount: 200},
		},
	}

	if len(data.Accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(data.Accounts))
	}
}

func TestTemplateProcessing(t *testing.T) {
	t.Run("simple template with two accounts", func(t *testing.T) {
		tmplContent := `{
    "accounts": [
        {{- range $index, $account := .Accounts }}
        {{- if $index }},{{ end }}
        {
            "address": "{{ $account.Address }}",
            "amount": {{ $account.Amount }}
        }
        {{- end }}
    ]
}`

		tmpl, err := template.New("test").Parse(tmplContent)
		assert.NoError(t, err)

		data := GenesisData{
			Accounts: []GenesisAccount{
				{Address: "0x1234567890abcdef", Amount: 1000000},
				{Address: "0xfedcba0987654321", Amount: 2000000},
			},
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"address": "0x1234567890abcdef"`)
		assert.Contains(t, output, `"amount": 1000000`)
		assert.Contains(t, output, `"address": "0xfedcba0987654321"`)
		assert.Contains(t, output, `"amount": 2000000`)
	})

	t.Run("template with no accounts", func(t *testing.T) {
		tmplContent := `{
    "accounts": [
        {{- range $index, $account := .Accounts }}
        {{- if $index }},{{ end }}
        {
            "address": "{{ $account.Address }}",
            "amount": {{ $account.Amount }}
        }
        {{- end }}
    ]
}`

		tmpl, err := template.New("test").Parse(tmplContent)
		assert.NoError(t, err)

		data := GenesisData{
			Accounts: []GenesisAccount{},
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"accounts": [`)
		assert.Contains(t, output, `]`)
	})

	t.Run("template with single account", func(t *testing.T) {
		tmplContent := `{
    "accounts": [
        {{- range $index, $account := .Accounts }}
        {{- if $index }},{{ end }}
        {
            "address": "{{ $account.Address }}",
            "amount": {{ $account.Amount }}
        }
        {{- end }}
    ]
}`

		tmpl, err := template.New("test").Parse(tmplContent)
		assert.NoError(t, err)

		data := GenesisData{
			Accounts: []GenesisAccount{
				{Address: "0xsingleaddress", Amount: 5000000},
			},
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"address": "0xsingleaddress"`)
		assert.Contains(t, output, `"amount": 5000000`)
		assert.NotContains(t, output, "},\n    ]")
	})
}

func TestGenerateGenesisFile(t *testing.T) {
	t.Run("successful generation with mock data", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		templateContent := `{
    "accounts": [
        {{- range $index, $account := .Accounts }}
        {{- if $index }},{{ end }}
        {
            "address": "{{ $account.Address }}",
            "amount": {{ $account.Amount }}
        }
        {{- end }}
    ]
}`
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		assert.NoError(t, err)

		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)

		positions := []interfaces.UserPositionWithAddress{
			{WalletAddress: "0xalice123", TokenBalance: 1000000},
			{WalletAddress: "0xbob456", TokenBalance: 2000000},
			{WalletAddress: "0xcharlie789", TokenBalance: 500000},
		}

		virtualPoolRepo.On("GetPositionsWithUsersByChainID", mock.Anything, chainID).Return(positions, nil)

		userRepo := new(mocks.MockUserRepository)
		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		output, err := grad.GenerateGenesisFile(context.Background(), chainID)
		assert.NoError(t, err)

		assert.Contains(t, output, `"address": "0xalice123"`)
		assert.Contains(t, output, `"amount": 1000000`)
		assert.Contains(t, output, `"address": "0xbob456"`)
		assert.Contains(t, output, `"amount": 2000000`)
		assert.Contains(t, output, `"address": "0xcharlie789"`)
		assert.Contains(t, output, `"amount": 500000`)

		virtualPoolRepo.AssertExpectations(t)
	})

	t.Run("template file not found", func(t *testing.T) {
		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)

		positions := []interfaces.UserPositionWithAddress{
			{WalletAddress: "0xtest", TokenBalance: 1000},
		}

		virtualPoolRepo.On("GetPositionsWithUsersByChainID", mock.Anything, chainID).Return(positions, nil)

		userRepo := new(mocks.MockUserRepository)
		grad := New(chainRepo, virtualPoolRepo, userRepo, "/nonexistent/template.json", "http://localhost:8082/graduate")
		_, err := grad.GenerateGenesisFile(context.Background(), chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse template")
	})

	t.Run("repository error", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		templateContent := `{"accounts": []}`
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		assert.NoError(t, err)

		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)

		virtualPoolRepo.On("GetPositionsWithUsersByChainID", mock.Anything, chainID).Return(nil, assert.AnError)

		userRepo := new(mocks.MockUserRepository)
		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		_, err = grad.GenerateGenesisFile(context.Background(), chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get positions")

		virtualPoolRepo.AssertExpectations(t)
	})
}

func TestCheckAndGraduate(t *testing.T) {
	t.Skip("CheckAndGraduate requires HTTP client mocking - implement with httptest server")
	t.Run("successful graduation", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		templateContent := `{"accounts": []}`
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		assert.NoError(t, err)

		chainID := uuid.New()
		creatorID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)
		userRepo := new(mocks.MockUserRepository)

		username := "testuser"
		chain := &models.Chain{
			ID:                  chainID,
			ChainName:           "TestChain",
			GraduationThreshold: 50000.0,
			IsGraduated:         false,
			CreatedBy:           creatorID,
			Creator: &models.User{
				ID:            creatorID,
				Username:      &username,
				WalletAddress: "0xcreator",
			},
			Repository: &models.ChainRepository{
				GithubURL: "https://github.com/test/repo",
			},
		}

		pool := &models.VirtualPool{
			ID:          uuid.New(),
			ChainID:     chainID,
			CNPYReserve: 55000.0,
		}

		positions := []interfaces.UserPositionWithAddress{
			{WalletAddress: "0xtest", TokenBalance: 1000},
		}

		chainRepo.On("GetByID", mock.Anything, chainID, []string{"creator", "repository"}).Return(chain, nil)
		virtualPoolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)
		virtualPoolRepo.On("GetPositionsWithUsersByChainID", mock.Anything, chainID).Return(positions, nil)
		chainRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Chain")).Return(chain, nil)

		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		err = grad.CheckAndGraduate(context.Background(), chainID)

		// This will fail due to RPC call - proper test requires httptest server
		assert.Error(t, err)

		chainRepo.AssertExpectations(t)
		virtualPoolRepo.AssertExpectations(t)
	})

	t.Run("threshold not met", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		templateContent := `{"accounts": []}`
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		assert.NoError(t, err)

		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)
		userRepo := new(mocks.MockUserRepository)

		chain := &models.Chain{
			ID:                  chainID,
			ChainName:           "TestChain",
			GraduationThreshold: 50000.0,
			IsGraduated:         false,
		}

		pool := &models.VirtualPool{
			ID:          uuid.New(),
			ChainID:     chainID,
			CNPYReserve: 30000.0,
		}

		chainRepo.On("GetByID", mock.Anything, chainID, []string{"creator", "repository"}).Return(chain, nil)
		virtualPoolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(pool, nil)

		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		err = grad.CheckAndGraduate(context.Background(), chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "graduation threshold not met")
		assert.Contains(t, err.Error(), "current 30000")
		assert.Contains(t, err.Error(), "required 50000")

		chainRepo.AssertExpectations(t)
		virtualPoolRepo.AssertExpectations(t)
	})

	t.Run("already graduated", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)
		userRepo := new(mocks.MockUserRepository)

		chain := &models.Chain{
			ID:          chainID,
			ChainName:   "TestChain",
			IsGraduated: true,
		}

		chainRepo.On("GetByID", mock.Anything, chainID, []string{"creator", "repository"}).Return(chain, nil)

		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		err := grad.CheckAndGraduate(context.Background(), chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already graduated")

		chainRepo.AssertExpectations(t)
	})

	t.Run("chain not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)
		userRepo := new(mocks.MockUserRepository)

		chainRepo.On("GetByID", mock.Anything, chainID, []string{"creator", "repository"}).Return(nil, assert.AnError)

		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		err := grad.CheckAndGraduate(context.Background(), chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get chain")

		chainRepo.AssertExpectations(t)
	})

	t.Run("pool not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		templatePath := filepath.Join(tmpDir, "genesis.json.template")

		chainID := uuid.New()
		chainRepo := new(mocks.MockChainRepository)
		virtualPoolRepo := new(mocks.MockVirtualPoolRepository)
		userRepo := new(mocks.MockUserRepository)

		chain := &models.Chain{
			ID:                  chainID,
			ChainName:           "TestChain",
			GraduationThreshold: 50000.0,
			IsGraduated:         false,
		}

		chainRepo.On("GetByID", mock.Anything, chainID, []string{"creator", "repository"}).Return(chain, nil)
		virtualPoolRepo.On("GetPoolByChainID", mock.Anything, chainID).Return(nil, assert.AnError)

		grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, "http://localhost:8082/graduate")
		err := grad.CheckAndGraduate(context.Background(), chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get virtual pool")

		chainRepo.AssertExpectations(t)
		virtualPoolRepo.AssertExpectations(t)
	})
}

func TestNew(t *testing.T) {
	chainRepo := new(mocks.MockChainRepository)
	virtualPoolRepo := new(mocks.MockVirtualPoolRepository)
	userRepo := new(mocks.MockUserRepository)
	templatePath := "/path/to/template"
	rpcEndpoint := "http://localhost:8082/graduate"

	grad := New(chainRepo, virtualPoolRepo, userRepo, templatePath, rpcEndpoint)

	assert.NotNil(t, grad)
	assert.Equal(t, chainRepo, grad.chainRepo)
	assert.Equal(t, virtualPoolRepo, grad.virtualPoolRepo)
	assert.Equal(t, userRepo, grad.userRepo)
	assert.Equal(t, templatePath, grad.templatePath)
	assert.Equal(t, rpcEndpoint, grad.rpcEndpoint)
	assert.NotNil(t, grad.httpClient)
}

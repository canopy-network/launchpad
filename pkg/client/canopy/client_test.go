package canopy

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	rpcURL := "http://localhost:42069"
	client := NewClient(rpcURL)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.rpcURL != rpcURL {
		t.Errorf("Expected rpcURL %s, got %s", rpcURL, client.rpcURL)
	}
}

func TestClientURL(t *testing.T) {
	client := NewClient("http://localhost:42069")

	tests := []struct {
		routeName string
		param     string
		expected  string
	}{
		{
			routeName: HeightRouteName,
			param:     "",
			expected:  "http://localhost:42069/v1/query/height",
		},
		{
			routeName: RootChainInfoRouteName,
			param:     "",
			expected:  "http://localhost:42069/v1/query/root-chain-info",
		},
		{
			routeName: OrdersRouteName,
			param:     "",
			expected:  "http://localhost:42069/v1/query/orders",
		},
		{
			routeName: StateRouteName,
			param:     "?height=100",
			expected:  "http://localhost:42069/v1/query/state?height=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.routeName, func(t *testing.T) {
			url := client.url(tt.routeName, tt.param)
			if url != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, url)
			}
		})
	}
}

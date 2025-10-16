package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server configuration
	Port        string
	Environment string

	// Database configuration
	DatabaseURL string

	// Security configuration
	JWTSecret          string
	JWTExpirationHours int

	// External services
	GithubClientID     string
	GithubClientSecret string

	// Application settings
	MaxFileUploadSize int64
	RequestTimeout    time.Duration

	// Pagination defaults
	DefaultPageSize int
	MaxPageSize     int

	// Root chain configuration
	RootChainURL    string // WebSocket URL for root chain subscription
	RootChainID     uint64 // Chain ID to subscribe to
	RootChainRPCURL string // HTTP URL for RPC client to fetch transactions

	// Graduation configuration
	GraduationRPCURL string // HTTP URL for graduation RPC endpoint
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "3001"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		MaxFileUploadSize:  getEnvInt64("MAX_FILE_UPLOAD_SIZE", 10*1024*1024), // 10MB
		RequestTimeout:     time.Duration(getEnvInt("REQUEST_TIMEOUT_SECONDS", 60)) * time.Second,
		DefaultPageSize:    getEnvInt("DEFAULT_PAGE_SIZE", 20),
		MaxPageSize:        getEnvInt("MAX_PAGE_SIZE", 100),
		RootChainURL:       getEnv("ROOT_CHAIN_URL", "ws://localhost:8081"),
		RootChainID:        uint64(getEnvInt("ROOT_CHAIN_ID", 1)),
		RootChainRPCURL:    getEnv("ROOT_CHAIN_RPC_URL", "http://localhost:8081"),
		GraduationRPCURL:   getEnv("GRADUATION_RPC_URL", "http://localhost:8082/graduate"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

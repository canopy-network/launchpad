package main

import (
	"log"

	"github.com/enielson/launchpad/internal/config"
	"github.com/enielson/launchpad/internal/repository/postgres"
	"github.com/enielson/launchpad/internal/server"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	templateRepo := postgres.NewChainTemplateRepository(db)
	chainRepo := postgres.NewChainRepository(db, userRepo, templateRepo)
	virtualPoolRepo := postgres.NewVirtualPoolRepository(db)

	// Initialize services
	chainService := services.NewChainService(chainRepo, templateRepo, userRepo, virtualPoolRepo)

	// Create services container
	services := &server.Services{
		ChainService: chainService,
	}

	// Create and start server
	srv := server.NewServer(cfg, services)

	// Start server (this blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

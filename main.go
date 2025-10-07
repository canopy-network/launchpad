package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/enielson/launchpad/internal/config"
	"github.com/enielson/launchpad/internal/repository/postgres"
	"github.com/enielson/launchpad/internal/server"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/workers/newblock"
	"github.com/enielson/launchpad/pkg/client/canopy"
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
	templateService := services.NewTemplateService(templateRepo)
	virtualPoolService := services.NewVirtualPoolService(virtualPoolRepo)

	// Initialize email service (always use SMTP)
	emailService := services.NewSMTPEmailService()
	log.Println("Using SMTP email service")

	authService := services.NewAuthService(emailService)

	// Create services container
	servicesContainer := &server.Services{
		ChainService:       chainService,
		TemplateService:    templateService,
		AuthService:        authService,
		VirtualPoolService: virtualPoolService,
	}

	// Initialize and start root chain event worker
	workerConfig := newblock.Config{
		RootChainURL:    cfg.RootChainURL,
		RootChainID:     cfg.RootChainID,
		RootChainRPCURL: cfg.RootChainRPCURL,
	}
	rpcClient := canopy.NewClient(cfg.RootChainRPCURL)
	worker := newblock.NewWorker(workerConfig, rpcClient, chainRepo, virtualPoolRepo, userRepo)

	// Start worker in background
	if err := worker.Start(); err != nil {
		log.Fatalf("Failed to start root chain worker: %v", err)
	}
	defer worker.Stop()

	log.Printf("Started root chain worker (ChainID: %d, URL: %s)", cfg.RootChainID, cfg.RootChainURL)

	// Create and start server
	srv := server.NewServer(cfg, servicesContainer)

	log.Printf("Starting Launchpad API server...")
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Port: %s", cfg.Port)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("Received shutdown signal, cleaning up...")
		if err := worker.Stop(); err != nil {
			log.Printf("Error stopping worker: %v", err)
		}
	case err := <-errChan:
		log.Fatalf("Server failed to start: %v", err)
	}
}

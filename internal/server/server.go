package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enielson/launchpad/internal/config"
	"github.com/enielson/launchpad/internal/handlers"
	custommiddleware "github.com/enielson/launchpad/internal/middleware"
	"github.com/enielson/launchpad/internal/services"
	"github.com/enielson/launchpad/internal/validators"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	Router           *chi.Mux
	Config           *config.Config
	Services         *Services
	Handlers         *Handlers
	EmailRateLimiter *custommiddleware.RateLimiter
}

type Services struct {
	ChainService       *services.ChainService
	TemplateService    *services.TemplateService
	AuthService        *services.AuthService
	VirtualPoolService *services.VirtualPoolService
	WalletService      *services.WalletService
	UserService        *services.UserService
}

type Handlers struct {
	ChainHandler       *handlers.ChainHandler
	TemplateHandler    *handlers.TemplateHandler
	AuthHandler        *handlers.AuthHandler
	VirtualPoolHandler *handlers.VirtualPoolHandler
	WalletHandler      *handlers.WalletHandler
	UserHandler        *handlers.UserHandler
}

func NewServer(cfg *config.Config, services *Services) *Server {
	// Create validator
	validator := validators.New()

	// Create handlers
	handlers := &Handlers{
		ChainHandler:       handlers.NewChainHandler(services.ChainService, validator),
		TemplateHandler:    handlers.NewTemplateHandler(services.TemplateService, validator),
		AuthHandler:        handlers.NewAuthHandler(services.AuthService, validator),
		VirtualPoolHandler: handlers.NewVirtualPoolHandler(services.VirtualPoolService, validator),
		WalletHandler:      handlers.NewWalletHandler(services.WalletService, validator),
		UserHandler:        handlers.NewUserHandler(services.UserService, validator),
	}

	// Configure rate limiting based on environment
	// In production: 10 seconds between email requests
	// In development/test: effectively disabled (1us)
	rateLimitInterval := 1 * time.Microsecond
	if cfg.IsProduction() {
		rateLimitInterval = 10 * time.Second
	}

	s := &Server{
		Router:           chi.NewRouter(),
		Config:           cfg,
		Services:         services,
		Handlers:         handlers,
		EmailRateLimiter: custommiddleware.NewRateLimiter(rateLimitInterval),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Standard middleware
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.CleanPath)
	s.Router.Use(middleware.Timeout(s.Config.RequestTimeout))

	// Logging middleware (conditional based on environment)
	if s.Config.IsDevelopment() {
		s.Router.Use(middleware.Logger)
	}

	// CORS configuration
	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify exact origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	}))

	// Content-Type middleware for JSON APIs
	s.Router.Use(s.jsonContentType)
}

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.Router.Get("/health", handlers.HealthCheck)

	// Debug/development routes
	if s.Config.IsDevelopment() {
		s.Router.Get("/api/v1/routes", handlers.ListRoutes(s.Router))
	}

	// Create auth middleware
	authMiddleware := custommiddleware.AuthMiddleware(s.Services.AuthService)

	// API v1 routes
	s.Router.Route("/api/v1", func(r chi.Router) {
		// Public routes (no authentication required)
		r.Group(func(r chi.Router) {
			// Auth endpoints with rate limiting
			// Email sending is rate limited in production (10 seconds per IP)
			// Rate limiting is disabled in development/test environments
			r.With(custommiddleware.RateLimitMiddleware(s.EmailRateLimiter)).Post("/auth/email", s.Handlers.AuthHandler.SendEmailCode)
			r.Post("/auth/verify", s.Handlers.AuthHandler.VerifyEmailCode)

			// Public read-only endpoints
			r.Get("/templates", s.Handlers.TemplateHandler.GetTemplates)
			r.Get("/chains", s.Handlers.ChainHandler.GetChains)
		})

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			// Use the new session token authentication middleware
			// TODO disabled for now
			// r.Use(authMiddleware)

			// Auth/session management routes
			r.Route("/auth", func(r chi.Router) {
				r.Post("/logout", s.Handlers.AuthHandler.Logout)
				r.Get("/sessions", s.Handlers.AuthHandler.GetSessions)
				r.Delete("/sessions/{id}", s.Handlers.AuthHandler.RevokeSession)
			})

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Put("/profile", s.Handlers.UserHandler.UpdateProfile)
			})

			// Virtual pool routes
			r.Route("/virtual-pools", func(r chi.Router) {
				r.Get("/", s.Handlers.VirtualPoolHandler.GetVirtualPools)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.Handlers.VirtualPoolHandler.GetVirtualPool)
				})
			})

			// Chain routes (protected operations only)
			r.Post("/chains", s.Handlers.ChainHandler.CreateChain)
			r.Route("/chains/{id}", func(r chi.Router) {
				r.Get("/", s.Handlers.ChainHandler.GetChain)
				r.Delete("/", s.Handlers.ChainHandler.DeleteChain)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.Handlers.ChainHandler.GetChain)
					r.Delete("/", s.Handlers.ChainHandler.DeleteChain)

					// Virtual pool endpoints
					r.Get("/transactions", s.Handlers.ChainHandler.GetTransactions)
				})
			})

			// Wallet routes
			r.Route("/wallets", func(r chi.Router) {
				r.Get("/", s.Handlers.WalletHandler.GetWallets)
				r.Post("/", s.Handlers.WalletHandler.CreateWallet)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.Handlers.WalletHandler.GetWallet)
					r.Put("/", s.Handlers.WalletHandler.UpdateWallet)
					r.Delete("/", s.Handlers.WalletHandler.DeleteWallet)

					// Wallet operations
					r.Post("/decrypt", s.Handlers.WalletHandler.DecryptWallet)
					r.Post("/unlock", s.Handlers.WalletHandler.UnlockWallet)
				})
			})
		})
	})

	// Catch-all for 404s
	s.Router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"code":"NOT_FOUND","message":"Route not found"}}`))
	})

	// Method not allowed
	s.Router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":{"code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}}`))
	})
}

// jsonContentType middleware sets Content-Type to application/json for API routes
func (s *Server) jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         ":" + s.Config.Port,
		Handler:      s.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server starting on port %s (environment: %s)\n", s.Config.Port, s.Config.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server failed to start: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	fmt.Println("Server exited")
	return nil
}

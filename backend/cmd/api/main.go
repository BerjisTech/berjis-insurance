// Insurance Broker Platform - API Server Entry Point
// This is the main entry point for the Go backend API
// Dependencies: fiber/v2, config, database, middleware
// Usage: go run cmd/api/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/insurance-broker/backend/internal/config"
	"github.com/insurance-broker/backend/internal/database"
	"github.com/insurance-broker/backend/internal/middleware"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Connect to PostgreSQL database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Fiber app with custom config
	app := fiber.New(fiber.Config{
		// AppName: displayed in Server header
		AppName: "Insurance Broker API v1.0.0",

		// ServerHeader: custom server identification
		ServerHeader: "Insurance-API",

		// Prefork: enable prefork mode in production for better performance
		Prefork: cfg.IsProduction(),

		// ErrorHandler: custom error handler
		ErrorHandler: customErrorHandler,

		// BodyLimit: maximum request body size (10MB)
		BodyLimit: 10 * 1024 * 1024,

		// ReadTimeout and WriteTimeout prevent slowloris attacks
		ReadTimeout:  15000, // 15 seconds
		WriteTimeout: 15000, // 15 seconds

		// DisableStartupMessage: hide ASCII art in production
		DisableStartupMessage: cfg.IsProduction(),
	})

	// =================================================================
	// Global Middleware
	// =================================================================

	// Recover middleware: recover from panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))

	// CORS middleware: handle cross-origin requests
	app.Use(middleware.CORS(&cfg.Server))

	// Logger middleware: log all HTTP requests
	app.Use(middleware.Logger())

	// =================================================================
	// Health Check Routes (no authentication required)
	// =================================================================

	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database connectivity
		if err := db.HealthCheck(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "unhealthy",
				"error":   "database connection failed",
				"details": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status":    "healthy",
			"service":   "insurance-broker-api",
			"version":   "1.0.0",
			"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Insurance Broker Platform API",
			"version": "1.0.0",
			"docs":    "/docs",
			"health":  "/health",
		})
	})

	// =================================================================
	// API v1 Routes
	// =================================================================

	api := app.Group("/api/v1")

	// Public routes (no authentication)
	public := api.Group("/public")
	public.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	})

	// Authentication routes (TODO: implement in next phase)
	auth := api.Group("/auth")
	auth.Post("/register", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Registration endpoint - Coming soon",
		})
	})
	auth.Post("/login", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Login endpoint - Coming soon",
		})
	})
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Token refresh endpoint - Coming soon",
		})
	})

	// Protected routes (require authentication)
	protected := api.Group("/", middleware.Protected(&cfg.JWT))

	// User routes
	protected.Get("/users/me", func(c *fiber.Ctx) error {
		userID := middleware.GetUserID(c)
		return c.JSON(fiber.Map{
			"message": "User profile endpoint - Coming soon",
			"user_id": userID,
		})
	})

	// Admin routes (require admin role)
	admin := api.Group("/admin", middleware.Protected(&cfg.JWT), middleware.RequireRole("system_admin"))
	admin.Get("/stats", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Admin statistics endpoint - Coming soon",
		})
	})

	// =================================================================
	// 404 Handler
	// =================================================================

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"path":    c.Path(),
		})
	})

	// =================================================================
	// Graceful Shutdown
	// =================================================================

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		address := ":" + cfg.Server.Port
		log.Printf("Starting server on %s (environment: %s)", address, cfg.Server.Env)

		if err := app.Listen(address); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server gracefully...")

	// Shutdown server with timeout
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// customErrorHandler handles errors globally
// Provides consistent error response format
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default error code
	code := fiber.StatusInternalServerError

	// Check if it's a fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Send error response
	return c.Status(code).JSON(fiber.Map{
		"error":   fiber.ErrInternalServerError.Message,
		"message": err.Error(),
		"code":    code,
	})
}

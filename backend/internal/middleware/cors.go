// Package middleware provides HTTP middleware functions
// This file handles CORS (Cross-Origin Resource Sharing) configuration
// Dependencies: fiber/v2, config
// Usage: app.Use(middleware.CORS(cfg))
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/insurance-broker/backend/internal/config"
)

// CORS returns configured CORS middleware
// Allows requests from configured origins with credentials support
// Implements security best practices for API access
func CORS(cfg *config.ServerConfig) fiber.Handler {
	return cors.New(cors.Config{
		// AllowOrigins: Comma-separated list of allowed origins
		// In production, this should be limited to your frontend domains
		AllowOrigins: joinStrings(cfg.CORSAllowedOrigins, ","),

		// AllowMethods: HTTP methods allowed for CORS requests
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",

		// AllowHeaders: Headers allowed in CORS requests
		// Authorization: for JWT tokens
		// Content-Type: for JSON payloads
		// X-Request-ID: for request tracing
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",

		// AllowCredentials: Allow cookies and authorization headers
		// Required for JWT authentication
		AllowCredentials: true,

		// ExposeHeaders: Headers exposed to the client
		// X-Request-ID: for request tracing
		// X-Total-Count: for pagination
		ExposeHeaders: "X-Request-ID,X-Total-Count",

		// MaxAge: How long the preflight request can be cached (seconds)
		// 12 hours to reduce preflight requests
		MaxAge: 43200,
	})
}

// joinStrings concatenates string slice with separator
// Helper function to format CORS origins
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}

	return result
}

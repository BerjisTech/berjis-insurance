// Package middleware provides HTTP middleware functions
// This file handles JWT authentication and authorization
// Dependencies: fiber/v2, jwt/v5, config
// Usage: app.Use(middleware.Protected(cfg))
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/insurance-broker/backend/internal/config"
)

// JWTClaims represents the JWT token payload
// Contains user identification and authorization info
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Protected returns JWT authentication middleware
// Validates JWT token from Authorization header
// Blocks request if token is invalid or missing
func Protected(cfg *config.JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		// Expected format: "Bearer <token>"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Split "Bearer" prefix from token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate JWT token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid signing method")
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Extract claims from token
		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Store user info in request context for later use
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)

		// Continue to next handler
		return c.Next()
	}
}

// RequireRole returns middleware that checks user role
// Blocks request if user doesn't have required role
// Must be used after Protected middleware
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user role from context (set by Protected middleware)
		userRole := c.Locals("userRole")
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User role not found in context",
			})
		}

		// Check if user role is in allowed roles
		roleStr := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				return c.Next()
			}
		}

		// User doesn't have required role
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// GetUserID retrieves user ID from request context
// Helper function for handlers that need user ID
// Returns empty string if not authenticated
func GetUserID(c *fiber.Ctx) string {
	userID := c.Locals("userID")
	if userID == nil {
		return ""
	}
	return userID.(string)
}

// GetUserRole retrieves user role from request context
// Helper function for handlers that need user role
// Returns empty string if not authenticated
func GetUserRole(c *fiber.Ctx) string {
	userRole := c.Locals("userRole")
	if userRole == nil {
		return ""
	}
	return userRole.(string)
}

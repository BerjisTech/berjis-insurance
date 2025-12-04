// Package middleware provides HTTP middleware functions
// This file handles request logging for debugging and monitoring
// Dependencies: fiber/v2, log
// Usage: app.Use(middleware.Logger())
package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger returns request logging middleware
// Logs: method, path, status code, latency, IP address
// Useful for debugging and monitoring API usage
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Extract request details
		method := c.Method()
		path := c.Path()
		statusCode := c.Response().StatusCode()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Log request with details
		log.Printf(
			"[%s] %s %s - Status: %d - Latency: %v - IP: %s - UA: %s",
			method,
			path,
			c.Protocol(),
			statusCode,
			latency,
			ip,
			truncateString(userAgent, 50),
		)

		return err
	}
}

// truncateString limits string length for cleaner logs
// Adds ellipsis if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

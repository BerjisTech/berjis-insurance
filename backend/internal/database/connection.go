// Package database manages PostgreSQL database connections
// Provides connection pooling and health check functionality
// Dependencies: database/sql, lib/pq, config
// Usage: db := database.Connect(cfg)
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/insurance-broker/backend/internal/config"
)

// DB wraps sql.DB with additional metadata
type DB struct {
	*sql.DB
	Config *config.DatabaseConfig
}

// Connect establishes connection to PostgreSQL database
// Configures connection pool and validates connectivity
// Returns error if connection fails
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	// Build PostgreSQL connection string
	// Format: postgres://user:password@host:port/dbname?sslmode=mode
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	// Open database connection
	// sql.Open doesn't actually connect yet, just prepares the connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// MaxOpenConns: Maximum number of open connections to the database
	db.SetMaxOpenConns(cfg.MaxConnections)

	// MaxIdleConns: Maximum number of connections in the idle connection pool
	db.SetMaxIdleConns(cfg.MaxIdleConnections)

	// ConnMaxLifetime: Maximum time a connection can be reused
	db.SetConnMaxLifetime(cfg.ConnectionLifetime)

	// Verify connection with ping
	// This is where the actual connection happens
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database: %s@%s:%s/%s",
		cfg.User, cfg.Host, cfg.Port, cfg.Name)

	return &DB{
		DB:     db,
		Config: cfg,
	}, nil
}

// Close gracefully closes the database connection pool
// Should be called during application shutdown
func (db *DB) Close() error {
	log.Println("Closing database connections...")
	return db.DB.Close()
}

// HealthCheck verifies database connectivity
// Returns error if database is unreachable
// Used by health check endpoints
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping with timeout
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Verify we can execute a simple query
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database returned unexpected result")
	}

	return nil
}

// GetStats returns database connection pool statistics
// Useful for monitoring and debugging
func (db *DB) GetStats() sql.DBStats {
	return db.Stats()
}

// Package config handles application configuration management
// It loads environment variables and provides typed configuration access
// Dependencies: os, time, godotenv
// Usage: cfg := config.Load()
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
// Fields are populated from environment variables
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	AI        AIConfig
	Email     EmailConfig
	SMS       SMSConfig
	MPesa     MPesaConfig
	RateLimit RateLimitConfig
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port               string   // Server port (default: 8096)
	Env                string   // Environment: development|staging|production
	CORSAllowedOrigins []string // Allowed CORS origins
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host               string        // Database host
	Port               string        // Database port
	Name               string        // Database name
	User               string        // Database user
	Password           string        // Database password
	SSLMode            string        // SSL mode: disable|require|verify-full
	MaxConnections     int           // Maximum open connections
	MaxIdleConnections int           // Maximum idle connections
	ConnectionLifetime time.Duration // Maximum connection lifetime
}

// JWTConfig contains JWT authentication settings
type JWTConfig struct {
	Secret        string        // Secret key for signing tokens (min 32 chars)
	Expiry        time.Duration // Access token expiry duration
	RefreshExpiry time.Duration // Refresh token expiry duration
}

// AIConfig contains AI service API keys
type AIConfig struct {
	OpenAIKey    string // OpenAI API key for GPT models
	AnthropicKey string // Anthropic API key for Claude models
}

// SMSConfig contains SMS service configuration (Africa's Talking)
type SMSConfig struct {
	APIKey   string // Africa's Talking API key
	Username string // Africa's Talking username
	SenderID string // Sender ID or short code
}

// MPesaConfig contains M-Pesa payment gateway configuration
type MPesaConfig struct {
	ConsumerKey    string // Safaricom consumer key
	ConsumerSecret string // Safaricom consumer secret
	Passkey        string // Lipa Na M-Pesa passkey
	Shortcode      string // Business shortcode
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Max    int           // Maximum requests per window
	Window time.Duration // Time window for rate limiting
}

// EmailConfig contains transactional email provider settings
type EmailConfig struct {
	Provider    string // e.g., sendgrid
	APIKey      string // provider API key
	FromAddress string // default from email
	FromName    string // display name
}

// Load reads environment variables and returns populated Config
// It loads .env file if present and validates required fields
// Panics if critical configuration is missing
func Load() *Config {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Parse database max connections with default
	maxConns, err := strconv.Atoi(getEnv("DB_MAX_CONNECTIONS", "25"))
	if err != nil {
		maxConns = 25
	}

	// Parse database max idle connections with default
	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNECTIONS", "5"))
	if err != nil {
		maxIdleConns = 5
	}

	// Parse database connection lifetime
	connLifetime, err := time.ParseDuration(getEnv("DB_CONNECTION_LIFETIME", "5m"))
	if err != nil {
		connLifetime = 5 * time.Minute
	}

	// Parse JWT expiry durations
	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	jwtRefreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		jwtRefreshExpiry = 168 * time.Hour
	}

	// Parse rate limit max requests
	rateLimitMax, err := strconv.Atoi(getEnv("RATE_LIMIT_MAX", "100"))
	if err != nil {
		rateLimitMax = 100
	}

	// Parse rate limit window
	rateLimitWindow, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))
	if err != nil {
		rateLimitWindow = 1 * time.Minute
	}

	// Build configuration struct
	cfg := &Config{
		Server: ServerConfig{
			Port:               getEnv("SERVER_PORT", "8096"),
			Env:                getEnv("SERVER_ENV", "development"),
			CORSAllowedOrigins: parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:7100")),
		},
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnv("DB_PORT", "5451"),
			Name:               getEnv("DB_NAME", "insurance_broker"),
			User:               getEnv("DB_USER", "insurance_admin"),
			Password:           getEnv("DB_PASSWORD", ""),
			SSLMode:            getEnv("DB_SSLMODE", "disable"),
			MaxConnections:     maxConns,
			MaxIdleConnections: maxIdleConns,
			ConnectionLifetime: connLifetime,
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			Expiry:        jwtExpiry,
			RefreshExpiry: jwtRefreshExpiry,
		},
		AI: AIConfig{
			OpenAIKey:    getEnv("OPENAI_API_KEY", ""),
			AnthropicKey: getEnv("ANTHROPIC_API_KEY", ""),
		},
		Email: EmailConfig{
			Provider:    getEnv("EMAIL_PROVIDER", "sendgrid"),
			APIKey:      getEnv("EMAIL_API_KEY", ""),
			FromAddress: getEnv("EMAIL_FROM_ADDRESS", "no-reply@insurance.local"),
			FromName:    getEnv("EMAIL_FROM_NAME", "Insurance Broker AI"),
		},
		SMS: SMSConfig{
			APIKey:   getEnv("SMS_API_KEY", ""),
			Username: getEnv("SMS_USERNAME", ""),
			SenderID: getEnv("SMS_SENDER_ID", "InsuranceAI"),
		},
		MPesa: MPesaConfig{
			ConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
			ConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
			Passkey:        getEnv("MPESA_PASSKEY", ""),
			Shortcode:      getEnv("MPESA_SHORTCODE", ""),
		},
		RateLimit: RateLimitConfig{
			Max:    rateLimitMax,
			Window: rateLimitWindow,
		},
	}

	// Validate critical configuration
	cfg.validate()

	return cfg
}

// validate checks that critical configuration values are present
// Logs warnings for missing optional configuration
func (c *Config) validate() {
	// Critical: Database password must be set in production
	if c.Server.Env == "production" && c.Database.Password == "" {
		log.Fatal("DB_PASSWORD must be set in production environment")
	}

	// Critical: JWT secret must be at least 32 characters
	if len(c.JWT.Secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters long")
	}

	if c.Email.APIKey == "" {
		log.Println("WARNING: EMAIL_API_KEY not configured (OTP emails will be logged only)")
	}

	// Warning: AI API keys (optional for development)
	if c.AI.OpenAIKey == "" && c.AI.AnthropicKey == "" {
		log.Println("WARNING: No AI API keys configured (OpenAI or Anthropic)")
	}

	// Warning: SMS API key (optional, required for OTP in production)
	if c.Server.Env == "production" && c.SMS.APIKey == "" {
		log.Println("WARNING: SMS_API_KEY not configured (required for OTP)")
	}

	// Warning: M-Pesa configuration (optional, required for payments)
	if c.Server.Env == "production" && c.MPesa.ConsumerKey == "" {
		log.Println("WARNING: M-Pesa configuration not complete (required for payments)")
	}
}

// getEnv retrieves environment variable with fallback default value
// Returns default if environment variable is not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// parseCORSOrigins splits comma-separated CORS origins string
// Returns slice of origin URLs
func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{"http://localhost:7100"}
	}

	// Simple split by comma (more sophisticated parsing could be added)
	result := []string{}
	current := ""
	for _, char := range origins {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	return result
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

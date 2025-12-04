// Package utils provides common utility functions
// This file handles cryptographic operations (hashing, encryption)
// Dependencies: golang.org/x/crypto/bcrypt
// Usage: hash := crypto.HashPassword(password)
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost defines the computational cost for password hashing
	// Higher cost = more secure but slower
	// 10 is a good balance for production (2^10 = 1024 iterations)
	BcryptCost = 10
)

// HashPassword creates bcrypt hash from plain text password
// Returns hashed password string or error
// Use this for storing passwords in database
func HashPassword(password string) (string, error) {
	// Validate password length (min 8 characters for security)
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	// Generate bcrypt hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePassword verifies if plain text password matches hash
// Returns true if password is correct, false otherwise
// Use this for login authentication
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// GenerateRandomToken creates cryptographically secure random token
// Length parameter specifies number of bytes (will be base64 encoded)
// Use this for reset tokens, API keys, etc.
func GenerateRandomToken(length int) (string, error) {
	// Validate length
	if length < 16 {
		return "", fmt.Errorf("token length must be at least 16 bytes")
	}

	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64 for safe string representation
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateOTP creates a 6-digit numeric OTP code
// Returns string representation of OTP
// Use this for email/SMS verification
func GenerateOTP() (string, error) {
	// Generate 3 random bytes (24 bits)
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Convert to integer and mod by 1000000 to get 6 digits
	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	otp := num % 1000000

	// Format as 6-digit string with leading zeros
	return fmt.Sprintf("%06d", otp), nil
}

// Insurance Broker Platform - Password Management
// Package: internal/auth
// Purpose: Secure password hashing and verification using bcrypt

package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength minimum password length
	MinPasswordLength = 8
	// BcryptCost cost factor for bcrypt hashing
	BcryptCost = 12
)

// PasswordService handles password operations
type PasswordService struct {
	cost int
}

// NewPasswordService creates a new password service
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: BcryptCost,
	}
}

// HashPassword generates a bcrypt hash of the password
func (s *PasswordService) HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", errors.New("password must be at least 8 characters")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func (s *PasswordService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword validates password strength
func (s *PasswordService) ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	// Add more validation rules as needed
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain uppercase, lowercase, and digit")
	}

	return nil
}

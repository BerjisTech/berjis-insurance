// Package models defines core domain models for the backend
// This file contains authentication-related models (User, RefreshToken, OTP, etc.)
// These structs map closely to the database schema and are used across repositories/services
package models

import (
	"database/sql"
	"time"
)

// User represents a system user record stored in the `users` table
// Only includes fields required for authentication. Additional profile fields will live in dedicated structs later.
type User struct {
	ID             string
	Email          string
	Phone          sql.NullString
	PasswordHash   string
	Role           string
	Status         string
	EmailVerified  bool
	PhoneVerified  bool
	LastLoginAt    sql.NullTime
	FailedAttempts int
	LockedUntil    sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OTPPurpose enumerates supported OTP flows
// Matches the otp_purpose enum declared in migrations
const (
	OTPPurposeEmailVerification = "email_verification"
	OTPPurposePhoneVerification = "phone_verification"
	OTPPurposeLogin             = "login"
	OTPPurposePasswordReset     = "password_reset"
)

// OTPCode represents a time-bound one-time password used for MFA and verification
// Identifier can be email or phone depending on channel
// Channel examples: "sms", "email"
type OTPCode struct {
	ID          string
	UserID      sql.NullString
	Identifier  string
	Code        string
	Purpose     string
	Channel     string
	Attempts    int
	MaxAttempts int
	ExpiresAt   time.Time
	VerifiedAt  sql.NullTime
	CreatedAt   time.Time
}

// RefreshToken represents a stored refresh token hash for session management
// tokenHash is stored instead of raw token for security. Hashing handled in service layer.
type RefreshToken struct {
	ID            string
	UserID        string
	TokenHash     string
	UserAgent     sql.NullString
	IPAddress     sql.NullString
	IssuedAt      time.Time
	ExpiresAt     time.Time
	RevokedAt     sql.NullTime
	RevokedReason sql.NullString
	CreatedAt     time.Time
}

// PasswordResetToken represents a secure token for password reset workflow
type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

// UserSession represents an authenticated device/session for audit tracking
type UserSession struct {
	ID             string
	UserID         string
	RefreshTokenID sql.NullString
	DeviceID       sql.NullString
	DeviceName     sql.NullString
	IPAddress      sql.NullString
	Location       sql.NullString
	UserAgent      sql.NullString
	IsActive       bool
	LastActiveAt   time.Time
	CreatedAt      time.Time
}

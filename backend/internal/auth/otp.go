// Insurance Broker Platform - OTP Management
// Package: internal/auth
// Purpose: Generate and verify One-Time Passwords for MFA backed by PostgreSQL storage

package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"time"

	"github.com/insurance-broker/backend/internal/models"
	"github.com/insurance-broker/backend/internal/repository"
)

const (
	// OTPLength standard OTP length
	OTPLength = 6
	// OTPExpiry OTP validity duration
	OTPExpiry = 10 * time.Minute
)

// OTPService handles OTP operations persisted in the database
type OTPService struct {
	repo      *repository.AuthRepository
	otpLength int
	ttl       time.Duration
}

// NewOTPService creates a new OTP service backed by repository storage
func NewOTPService(repo *repository.AuthRepository) *OTPService {
	return &OTPService{
		repo:      repo,
		otpLength: OTPLength,
		ttl:       OTPExpiry,
	}
}

// GenerateAndStore creates an OTP for the provided identifier + purpose and stores it
// identifier is typically an email or phone number
func (s *OTPService) GenerateAndStore(ctx context.Context, userID sql.NullString, identifier, purpose, channel string) (string, error) {
	code, err := generateRandomOTP(s.otpLength)
	if err != nil {
		return "", err
	}

	otp := &models.OTPCode{
		UserID:      userID,
		Identifier:  identifier,
		Code:        code,
		Purpose:     purpose,
		Channel:     channel,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(s.ttl),
	}

	if _, err := s.repo.CreateOTP(ctx, otp); err != nil {
		return "", err
	}

	return code, nil
}

// Verify validates the OTP for identifier+purpose and marks it as used if valid
func (s *OTPService) Verify(ctx context.Context, identifier, purpose, code string) error {
	otp, err := s.repo.GetActiveOTP(ctx, identifier, purpose)
	if err != nil {
		return err
	}

	if time.Now().After(otp.ExpiresAt) {
		return errors.New("OTP expired")
	}

	if otp.Attempts >= otp.MaxAttempts {
		return errors.New("too many attempts")
	}

	if otp.Code != code {
		attempts, incErr := s.repo.IncrementOTPAttempts(ctx, otp.ID)
		if incErr == nil && attempts >= otp.MaxAttempts {
			_ = s.repo.DeleteOTPs(ctx, identifier, purpose)
		}
		return errors.New("invalid OTP")
	}

	if err := s.repo.MarkOTPVerified(ctx, otp.ID); err != nil {
		return err
	}

	return nil
}

// CleanupExpired removes stale OTPs
func (s *OTPService) CleanupExpired(ctx context.Context) error {
	return s.repo.CleanupExpiredOTPs(ctx)
}

// generateRandomOTP generates a random numeric OTP
func generateRandomOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}

	return string(b), nil
}

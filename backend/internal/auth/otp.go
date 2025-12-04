// Insurance Broker Platform - OTP Management
// Package: internal/auth
// Purpose: Generate and verify One-Time Passwords for MFA

package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"
)

const (
	// OTPLength standard OTP length
	OTPLength = 6
	// OTPExpiry OTP validity duration
	OTPExpiry = 10 * time.Minute
)

// OTPService handles OTP operations
type OTPService struct {
	// In production, store OTPs in Redis/database with expiry
	otps map[string]*OTPData
}

// OTPData stores OTP information
type OTPData struct {
	Code      string
	ExpiresAt time.Time
	Attempts  int
}

// NewOTPService creates a new OTP service
func NewOTPService() *OTPService {
	return &OTPService{
		otps: make(map[string]*OTPData),
	}
}

// GenerateOTP generates a 6-digit OTP
func (s *OTPService) GenerateOTP(identifier string) (string, error) {
	otp, err := generateRandomOTP(OTPLength)
	if err != nil {
		return "", err
	}

	s.otps[identifier] = &OTPData{
		Code:      otp,
		ExpiresAt: time.Now().Add(OTPExpiry),
		Attempts:  0,
	}

	return otp, nil
}

// VerifyOTP verifies an OTP for the given identifier
func (s *OTPService) VerifyOTP(identifier, code string) error {
	data, exists := s.otps[identifier]
	if !exists {
		return errors.New("OTP not found")
	}

	if time.Now().After(data.ExpiresAt) {
		delete(s.otps, identifier)
		return errors.New("OTP expired")
	}

	data.Attempts++
	if data.Attempts > 3 {
		delete(s.otps, identifier)
		return errors.New("too many attempts")
	}

	if data.Code != code {
		return errors.New("invalid OTP")
	}

	// OTP verified successfully, remove it
	delete(s.otps, identifier)
	return nil
}

// generateRandomOTP generates a random numeric OTP
func generateRandomOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}

	return string(b), nil
}

// CleanExpiredOTPs removes expired OTPs (should be called periodically)
func (s *OTPService) CleanExpiredOTPs() {
	now := time.Now()
	for key, data := range s.otps {
		if now.After(data.ExpiresAt) {
			delete(s.otps, key)
		}
	}
}

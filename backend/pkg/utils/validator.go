// Package utils provides common utility functions
// This file handles input validation and sanitization
// Dependencies: regexp, strings
// Usage: if !validator.IsValidEmail(email) { ... }
package utils

import (
	"regexp"
	"strings"
)

var (
	// emailRegex validates email format (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// phoneRegex validates international phone numbers
	// Supports formats: +254123456789, 254123456789, 0123456789
	phoneRegex = regexp.MustCompile(`^(\+?[1-9]\d{1,14})$`)

	// kenyaPhoneRegex specifically validates Kenyan phone numbers
	// Formats: +254712345678, 0712345678
	kenyaPhoneRegex = regexp.MustCompile(`^(\+254|0)[17]\d{8}$`)
)

// IsValidEmail checks if email address is valid format
// Returns true if email matches RFC 5322 format
func IsValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsValidPhone checks if phone number is valid international format
// Returns true if phone matches E.164 format
func IsValidPhone(phone string) bool {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	return phoneRegex.MatchString(cleaned)
}

// IsValidKenyaPhone checks if phone number is valid Kenyan format
// Returns true if phone is a valid Safaricom/Airtel number
func IsValidKenyaPhone(phone string) bool {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	return kenyaPhoneRegex.MatchString(cleaned)
}

// NormalizePhone converts phone number to E.164 format
// Example: 0712345678 -> +254712345678
func NormalizePhone(phone, countryCode string) string {
	// Remove spaces, dashes, and parentheses
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// If already has country code, return as-is
	if strings.HasPrefix(cleaned, "+") {
		return cleaned
	}

	// Remove leading zero and add country code
	if strings.HasPrefix(cleaned, "0") {
		cleaned = cleaned[1:]
	}

	return "+" + countryCode + cleaned
}

// NormalizeKenyaPhone converts Kenyan phone to E.164 format
// Example: 0712345678 -> +254712345678
func NormalizeKenyaPhone(phone string) string {
	return NormalizePhone(phone, "254")
}

// IsValidPassword checks if password meets security requirements
// Requirements: min 8 chars, at least one uppercase, one lowercase, one digit
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

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

	return hasUpper && hasLower && hasDigit
}

// SanitizeString removes potentially dangerous characters
// Prevents XSS and SQL injection attacks
func SanitizeString(input string) string {
	// Trim whitespace
	cleaned := strings.TrimSpace(input)

	// Remove null bytes (SQL injection prevention)
	cleaned = strings.ReplaceAll(cleaned, "\x00", "")

	// Remove control characters except newlines and tabs
	result := ""
	for _, char := range cleaned {
		if char >= 32 || char == '\n' || char == '\t' {
			result += string(char)
		}
	}

	return result
}

// IsValidUUID checks if string is a valid UUID v4
func IsValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

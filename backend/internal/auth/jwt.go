// Insurance Broker Platform - JWT Authentication
// Package: internal/auth
// Purpose: JWT token generation, validation, and refresh token management
// Dependencies: github.com/golang-jwt/jwt/v5

package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Custom JWT claims structure
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService creates a new JWT service instance
func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) (*JWTService, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 characters")
	}
	
	return &JWTService{
		secretKey:     []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateAccessToken creates a new access token
func (s *JWTService) GenerateAccessToken(userID, email, role string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken creates a new refresh token
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken creates a new access token from a valid refresh token
func (s *JWTService) RefreshAccessToken(refreshToken, email, role string) (string, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	if claims.TokenType != "refresh" {
		return "", errors.New("invalid token type")
	}

	return s.GenerateAccessToken(claims.UserID, email, role)
}

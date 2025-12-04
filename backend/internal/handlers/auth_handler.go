// Insurance Broker Platform - Authentication HTTP Handlers
// Package: internal/handlers
// Purpose: HTTP endpoints for user authentication (login, register, refresh)

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/insurance-broker/backend/internal/auth"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	jwtService      *auth.JWTService
	passwordService *auth.PasswordService
	otpService      *auth.OTPService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtService *auth.JWTService, passwordService *auth.PasswordService, otpService *auth.OTPService) *AuthHandler {
	return &AuthHandler{
		jwtService:      jwtService,
		passwordService: passwordService,
		otpService:      otpService,
	}
}

// RegisterRequest represents registration payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents login payload  
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenResponse represents token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Register handles user registration
// POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate password strength
	if err := h.passwordService.ValidatePassword(req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Hash password
	hashedPassword, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process password",
		})
	}

	// TODO: Save user to database
	// For now, return placeholder response
	_ = hashedPassword

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"email":   req.Email,
	})
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: Fetch user from database and verify password
	// For now, return placeholder tokens
	userID := "temp-user-id"
	role := "user"

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(userID, req.Email, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
		})
	}

	return c.JSON(TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds
	})
}

// RefreshToken handles token refresh
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	// TODO: Fetch user data from database
	email := "user@example.com"
	role := "user"

	// Generate new access token
	accessToken, err := h.jwtService.RefreshAccessToken(req.RefreshToken, email, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user_id":      claims.UserID,
	})
}

// Logout handles user logout
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// TODO: Invalidate token (add to blacklist in Redis)
	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

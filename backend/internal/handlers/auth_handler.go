// Insurance Broker Platform - Authentication HTTP Handlers
// Package: internal/handlers
// Purpose: HTTP endpoints for user authentication (login, register, refresh)

package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/insurance-broker/backend/internal/models"
	"github.com/insurance-broker/backend/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	service *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// RegisterRequest represents registration payload
type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// LoginRequest represents login payload
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
}

// TokenResponse represents token response
type TokenResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int          `json:"expiresIn"`
	User         UserResponse `json:"user"`
}

// UserResponse is returned to clients without sensitive fields
type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Phone         string `json:"phone,omitempty"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
	PhoneVerified bool   `json:"phoneVerified"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// VerifyOTPRequest payload
type VerifyOTPRequest struct {
	UserID  string `json:"userId"`
	Purpose string `json:"purpose"`
	Code    string `json:"code"`
}

// RefreshRequest payload
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// LogoutRequest payload
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// PasswordResetRequest payload
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirmRequest payload
type PasswordResetConfirmRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.service.Register(c.UserContext(), services.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailExists):
			return fiber.NewError(fiber.StatusConflict, err.Error())
		case errors.Is(err, services.ErrPhoneExists):
			return fiber.NewError(fiber.StatusConflict, err.Error())
		default:
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	response := fiber.Map{
		"message":              "User registered successfully",
		"userId":               result.UserID,
		"requiresVerification": result.RequiresVerification,
	}

	if result.DevelopmentOTP != "" {
		response["debugOtp"] = result.DevelopmentOTP
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	userAgent := c.Get("User-Agent")
	ipAddress := c.IP()

	result, err := h.service.Login(c.UserContext(), services.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, services.ErrAccountLocked):
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		case errors.Is(err, services.ErrVerificationRequired):
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	resp := TokenResponse{
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresIn:    result.Tokens.ExpiresIn,
		User:         mapUser(result.User),
	}

	return c.JSON(resp)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.service.RefreshTokens(c.UserContext(), services.RefreshInput{RefreshToken: req.RefreshToken})
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := TokenResponse{
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresIn:    result.Tokens.ExpiresIn,
		User:         mapUser(result.User),
	}

	return c.JSON(resp)
}

// VerifyOTP handles OTP verification
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err := h.service.VerifyOTP(c.UserContext(), services.VerifyOTPInput{
		UserID:  req.UserID,
		Purpose: req.Purpose,
		Code:    req.Code,
	})
	if err != nil {
		if errors.Is(err, services.ErrInvalidOTP) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Verification successful",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.service.Logout(c.UserContext(), services.LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// RequestPasswordReset handles reset initiation
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	var req PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.service.RequestPasswordReset(c.UserContext(), services.PasswordResetRequestInput{Email: req.Email})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := fiber.Map{"message": "If the email exists, a reset token has been sent"}
	if result.DevelopmentToken != "" {
		resp["debugToken"] = result.DevelopmentToken
	}

	return c.JSON(resp)
}

// ResetPassword handles password reset confirmation
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req PasswordResetConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.service.ResetPassword(c.UserContext(), services.PasswordResetConfirmInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		if errors.Is(err, services.ErrInvalidOTP) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}

func mapUser(user *models.User) UserResponse {
	if user == nil {
		return UserResponse{}
	}

	resp := UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
	}

	if user.Phone.Valid {
		resp.Phone = user.Phone.String
	}

	return resp
}

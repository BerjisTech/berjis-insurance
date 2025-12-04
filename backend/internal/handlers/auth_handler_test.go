package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/insurance-broker/backend/internal/auth"
	"github.com/insurance-broker/backend/internal/models"
	"github.com/insurance-broker/backend/internal/repository"
	"github.com/insurance-broker/backend/internal/services"
)

func TestResendOTPRoute(t *testing.T) {
	h := newHandlerTestHarness(t)
	res, err := h.service.Register(context.Background(), services.RegisterInput{
		Email:    "route@test.com",
		Password: "Password1",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	app := fiber.New()
	app.Post("/auth/resend-otp", h.handler.ResendOTP)

	payload := fmt.Sprintf(`{"userId":"%s","purpose":"email_verification"}`, res.UserID)
	req := httptest.NewRequest(http.MethodPost, "/auth/resend-otp", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("fiber test failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

type handlerTestHarness struct {
	service *services.AuthService
	handler *AuthHandler
}

func newHandlerTestHarness(t *testing.T) *handlerTestHarness {
	t.Helper()

	users := &handlerUserStore{users: make(map[string]*models.User)}
	authRepo := &handlerAuthRepo{}
	notifier := &handlerNotifier{}
	otp := &handlerOTP{}
	jwtService, err := auth.NewJWTService(strings.Repeat("x", 64), time.Minute*15, time.Hour*24)
	if err != nil {
		t.Fatalf("jwt init failed: %v", err)
	}

	service := services.NewAuthService(users, authRepo, jwtService, auth.NewPasswordService(), otp, notifier, services.AuthServiceOptions{
		Env:        "test",
		AccessTTL:  time.Minute * 15,
		RefreshTTL: time.Hour * 24,
	})

	return &handlerTestHarness{
		service: service,
		handler: NewAuthHandler(service),
	}
}

// Minimal implementations of interfaces for handler test

type handlerUserStore struct {
	users map[string]*models.User
}

func (h *handlerUserStore) CreateUser(ctx context.Context, user *models.User) (string, error) {
	user.ID = uuid.NewString()
	h.users[user.ID] = user
	return user.ID, nil
}

func (h *handlerUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range h.users {
		if strings.EqualFold(user.Email, email) {
			return user, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (h *handlerUserStore) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, repository.ErrNotFound
}

func (h *handlerUserStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	if user, ok := h.users[id]; ok {
		return user, nil
	}
	return nil, repository.ErrNotFound
}

func (h *handlerUserStore) UpdateVerification(ctx context.Context, id string, emailVerified, phoneVerified bool, status string) error {
	if user, ok := h.users[id]; ok {
		user.EmailVerified = emailVerified
		user.PhoneVerified = phoneVerified
		user.Status = status
		return nil
	}
	return repository.ErrNotFound
}

func (h *handlerUserStore) UpdateLastLogin(ctx context.Context, id string, ts time.Time) error {
	return nil
}
func (h *handlerUserStore) IncrementFailedAttempts(ctx context.Context, id string, lockUntil sql.NullTime) error {
	return nil
}
func (h *handlerUserStore) UpdatePassword(ctx context.Context, id, newHash string) error { return nil }

// Auth repo mock

type handlerAuthRepo struct{}

func (h *handlerAuthRepo) StoreRefreshToken(ctx context.Context, token *models.RefreshToken) (string, error) {
	return uuid.NewString(), nil
}
func (h *handlerAuthRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	return nil, repository.ErrNotFound
}
func (h *handlerAuthRepo) UpdateSessionActivityByRefresh(ctx context.Context, refreshTokenID string, lastActive time.Time) error {
	return nil
}
func (h *handlerAuthRepo) RevokeRefreshToken(ctx context.Context, id, reason string) error {
	return nil
}
func (h *handlerAuthRepo) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) (string, error) {
	return uuid.NewString(), nil
}
func (h *handlerAuthRepo) GetPasswordResetToken(ctx context.Context, hash string) (*models.PasswordResetToken, error) {
	return nil, repository.ErrNotFound
}
func (h *handlerAuthRepo) MarkPasswordResetUsed(ctx context.Context, id string) error { return nil }
func (h *handlerAuthRepo) CreateSession(ctx context.Context, session *models.UserSession) (string, error) {
	return uuid.NewString(), nil
}

// OTP + notifier fakes

type handlerOTP struct{}

func (h *handlerOTP) GenerateAndStore(ctx context.Context, userID sql.NullString, identifier, purpose, channel string) (string, error) {
	return "123456", nil
}
func (h *handlerOTP) Verify(ctx context.Context, identifier, purpose, code string) error { return nil }

type handlerNotifier struct{}

func (h *handlerNotifier) SendOTP(ctx context.Context, channel, identifier, code string) error {
	return nil
}
func (h *handlerNotifier) SendPasswordReset(ctx context.Context, email, token string) error {
	return nil
}

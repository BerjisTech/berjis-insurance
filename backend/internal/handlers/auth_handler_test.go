package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func TestAuthRoutes_RegisterVerifyLoginFlow(t *testing.T) {
	h := newHandlerTestHarness(t)
	app := fiber.New()
	app.Post("/auth/register", h.handler.Register)
	app.Post("/auth/verify-otp", h.handler.VerifyOTP)
	app.Post("/auth/login", h.handler.Login)

	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{
		"email":"integration@test.com",
		"password":"Password1"
	}`))
	registerReq.Header.Set("Content-Type", "application/json")

	registerResp, err := app.Test(registerReq)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer registerResp.Body.Close()
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 register, got %d", registerResp.StatusCode)
	}
	var registerBody struct {
		UserID string `json:"userId"`
	}
	decodeBody(t, registerResp.Body, &registerBody)
	if registerBody.UserID == "" {
		t.Fatalf("expected userId in response")
	}

	verifyPayload := fmt.Sprintf(`{"userId":"%s","purpose":"email_verification","code":"%s"}`, registerBody.UserID, h.otp.currentCode())
	verifyReq := httptest.NewRequest(http.MethodPost, "/auth/verify-otp", bytes.NewBufferString(verifyPayload))
	verifyReq.Header.Set("Content-Type", "application/json")

	verifyResp, err := app.Test(verifyReq)
	if err != nil {
		t.Fatalf("verify request failed: %v", err)
	}
	defer verifyResp.Body.Close()
	if verifyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 verify, got %d", verifyResp.StatusCode)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"email":"integration@test.com",
		"password":"Password1"
	}`))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 login, got %d", loginResp.StatusCode)
	}
	var loginBody TokenResponse
	decodeBody(t, loginResp.Body, &loginBody)
	if loginBody.AccessToken == "" || loginBody.RefreshToken == "" {
		t.Fatalf("expected tokens in login response")
	}
}

func decodeBody(t *testing.T, reader io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(reader).Decode(target); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
}

type handlerTestHarness struct {
	service *services.AuthService
	handler *AuthHandler
	otp     *handlerOTP
}

func newHandlerTestHarness(t *testing.T) *handlerTestHarness {
	t.Helper()

	users := &handlerUserStore{users: make(map[string]*models.User)}
	authRepo := &handlerAuthRepo{}
	notifier := &handlerNotifier{}
	otp := newHandlerOTP()
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
		otp:     otp,
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

type handlerOTP struct {
	code string
}

func newHandlerOTP() *handlerOTP {
	return &handlerOTP{code: "123456"}
}

func (h *handlerOTP) GenerateAndStore(ctx context.Context, userID sql.NullString, identifier, purpose, channel string) (string, error) {
	return h.code, nil
}
func (h *handlerOTP) Verify(ctx context.Context, identifier, purpose, code string) error {
	if code != h.code {
		return errors.New("invalid otp")
	}
	return nil
}
func (h *handlerOTP) currentCode() string {
	return h.code
}

type handlerNotifier struct{}

func (h *handlerNotifier) SendOTP(ctx context.Context, channel, identifier, code string) error {
	return nil
}
func (h *handlerNotifier) SendPasswordReset(ctx context.Context, email, token string) error {
	return nil
}

package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insurance-broker/backend/internal/auth"
	"github.com/insurance-broker/backend/internal/models"
	"github.com/insurance-broker/backend/internal/repository"
)

func TestAuthService_RegisterSendsOTP(t *testing.T) {
	deps := newTestAuthService(t)

	res, err := deps.service.Register(context.Background(), RegisterInput{
		Email:    "example@test.com",
		Password: "Password1",
		Phone:    "+254712345678",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if !res.RequiresVerification {
		t.Fatalf("expected verification flag")
	}
	if deps.notifier.lastChannel != "sms" {
		t.Fatalf("expected sms channel, got %s", deps.notifier.lastChannel)
	}
	if deps.otp.lastIdentifier != "+254712345678" {
		t.Fatalf("otp stored for wrong identifier")
	}
}

func TestAuthService_ResendOTP(t *testing.T) {
	deps := newTestAuthService(t)

	res, err := deps.service.Register(context.Background(), RegisterInput{
		Email:    "user@domain.com",
		Password: "Password1",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	deps.notifier.reset()
	deps.otp.codeToReturn = "654321"

	if err := deps.service.ResendOTP(context.Background(), ResendOTPInput{
		UserID:  res.UserID,
		Purpose: models.OTPPurposeEmailVerification,
	}); err != nil {
		t.Fatalf("resend failed: %v", err)
	}

	if deps.notifier.lastCode != "654321" {
		t.Fatalf("expected resent code to propagate")
	}
}

func TestAuthService_RequestPasswordResetSendsEmail(t *testing.T) {
	deps := newTestAuthService(t)
	user := &models.User{
		ID:            uuid.NewString(),
		Email:         "reset@example.com",
		PasswordHash:  "hash",
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	deps.users.seedUser(user)

	if _, err := deps.service.RequestPasswordReset(context.Background(), PasswordResetRequestInput{Email: user.Email}); err != nil {
		t.Fatalf("password reset failed: %v", err)
	}

	if deps.notifier.lastResetEmail != user.Email {
		t.Fatalf("expected email notification to be sent")
	}
}

type testDependencies struct {
	users    *memoryUserStore
	authRepo *memoryAuthRepo
	otp      *fakeOTPManager
	notifier *fakeNotifier
	service  *AuthService
}

func newTestAuthService(t *testing.T) *testDependencies {
	t.Helper()

	users := newMemoryUserStore()
	authRepo := newMemoryAuthRepo()
	otp := &fakeOTPManager{codeToReturn: "123456"}
	notifier := &fakeNotifier{}
	jwtService, err := auth.NewJWTService(strings.Repeat("x", 64), time.Minute*15, time.Hour*24)
	if err != nil {
		t.Fatalf("failed to init jwt: %v", err)
	}

	service := NewAuthService(users, authRepo, jwtService, auth.NewPasswordService(), otp, notifier, AuthServiceOptions{
		Env:        "test",
		AccessTTL:  time.Minute * 15,
		RefreshTTL: time.Hour * 24,
	})

	return &testDependencies{
		users:    users,
		authRepo: authRepo,
		otp:      otp,
		notifier: notifier,
		service:  service,
	}
}

// memoryUserStore implements UserStore for tests
type memoryUserStore struct {
	users map[string]*models.User
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{users: make(map[string]*models.User)}
}

func (m *memoryUserStore) CreateUser(ctx context.Context, user *models.User) (string, error) {
	user.ID = uuid.NewString()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = user.CreatedAt
	m.users[user.ID] = user
	return user.ID, nil
}

func (m *memoryUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	lower := strings.ToLower(email)
	for _, user := range m.users {
		if strings.ToLower(user.Email) == lower {
			return user, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memoryUserStore) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	for _, user := range m.users {
		if user.Phone.Valid && user.Phone.String == phone {
			return user, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memoryUserStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, repository.ErrNotFound
}

func (m *memoryUserStore) UpdateVerification(ctx context.Context, id string, emailVerified, phoneVerified bool, status string) error {
	if user, ok := m.users[id]; ok {
		user.EmailVerified = emailVerified
		user.PhoneVerified = phoneVerified
		user.Status = status
		user.UpdatedAt = time.Now()
		return nil
	}
	return repository.ErrNotFound
}

func (m *memoryUserStore) UpdateLastLogin(ctx context.Context, id string, ts time.Time) error {
	if user, ok := m.users[id]; ok {
		user.LastLoginAt = sqlNullTime(ts)
		user.UpdatedAt = time.Now()
		return nil
	}
	return repository.ErrNotFound
}

func (m *memoryUserStore) IncrementFailedAttempts(ctx context.Context, id string, lockUntil sql.NullTime) error {
	if user, ok := m.users[id]; ok {
		user.FailedAttempts++
		user.LockedUntil = lockUntil
		return nil
	}
	return repository.ErrNotFound
}

func (m *memoryUserStore) UpdatePassword(ctx context.Context, id, newHash string) error {
	if user, ok := m.users[id]; ok {
		user.PasswordHash = newHash
		return nil
	}
	return repository.ErrNotFound
}

func (m *memoryUserStore) seedUser(user *models.User) {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = user.CreatedAt
	m.users[user.ID] = user
}

// memoryAuthRepo implements AuthDataStore with in-memory maps
type memoryAuthRepo struct{}

func newMemoryAuthRepo() *memoryAuthRepo { return &memoryAuthRepo{} }

func (m *memoryAuthRepo) StoreRefreshToken(ctx context.Context, token *models.RefreshToken) (string, error) {
	token.ID = uuid.NewString()
	return token.ID, nil
}

func (m *memoryAuthRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryAuthRepo) UpdateSessionActivityByRefresh(ctx context.Context, refreshTokenID string, lastActive time.Time) error {
	return nil
}

func (m *memoryAuthRepo) RevokeRefreshToken(ctx context.Context, id, reason string) error {
	return nil
}

func (m *memoryAuthRepo) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) (string, error) {
	token.ID = uuid.NewString()
	return token.ID, nil
}

func (m *memoryAuthRepo) GetPasswordResetToken(ctx context.Context, hash string) (*models.PasswordResetToken, error) {
	return nil, repository.ErrNotFound
}

func (m *memoryAuthRepo) MarkPasswordResetUsed(ctx context.Context, id string) error {
	return nil
}

func (m *memoryAuthRepo) CreateSession(ctx context.Context, session *models.UserSession) (string, error) {
	session.ID = uuid.NewString()
	return session.ID, nil
}

// fakeOTPManager returns deterministic OTP codes
type fakeOTPManager struct {
	lastIdentifier string
	codeToReturn   string
}

func (f *fakeOTPManager) GenerateAndStore(ctx context.Context, userID sql.NullString, identifier, purpose, channel string) (string, error) {
	f.lastIdentifier = identifier
	if f.codeToReturn == "" {
		f.codeToReturn = "123456"
	}
	return f.codeToReturn, nil
}

func (f *fakeOTPManager) Verify(ctx context.Context, identifier, purpose, code string) error {
	if code != f.codeToReturn {
		return errors.New("invalid OTP")
	}
	return nil
}

// fakeNotifier captures notification attempts
type fakeNotifier struct {
	lastChannel    string
	lastIdentifier string
	lastCode       string
	lastResetEmail string
}

func (f *fakeNotifier) SendOTP(ctx context.Context, channel, identifier, code string) error {
	f.lastChannel = channel
	f.lastIdentifier = identifier
	f.lastCode = code
	return nil
}

func (f *fakeNotifier) SendPasswordReset(ctx context.Context, email, token string) error {
	f.lastResetEmail = email
	return nil
}

func (f *fakeNotifier) reset() {
	f.lastChannel = ""
	f.lastIdentifier = ""
	f.lastCode = ""
}

func sqlNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Valid: true, Time: t}
}

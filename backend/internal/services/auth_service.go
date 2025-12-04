// Package services contains business logic. AuthService orchestrates registration, login, OTP, and password reset flows.
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/insurance-broker/backend/internal/auth"
	"github.com/insurance-broker/backend/internal/models"
	"github.com/insurance-broker/backend/internal/repository"
	"github.com/insurance-broker/backend/pkg/utils"
)

var (
	ErrEmailExists          = errors.New("email already registered")
	ErrPhoneExists          = errors.New("phone already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountLocked        = errors.New("account temporarily locked due to failed attempts")
	ErrVerificationRequired = errors.New("account pending verification")
	ErrInvalidOTP           = errors.New("invalid or expired OTP")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
)

// UserStore defines persistence behavior needed by AuthService
type UserStore interface {
	CreateUser(ctx context.Context, user *models.User) (string, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	UpdateVerification(ctx context.Context, id string, emailVerified, phoneVerified bool, status string) error
	UpdateLastLogin(ctx context.Context, id string, ts time.Time) error
	IncrementFailedAttempts(ctx context.Context, id string, lockUntil sql.NullTime) error
	UpdatePassword(ctx context.Context, id, newHash string) error
}

// AuthDataStore defines refresh token/session persistence
type AuthDataStore interface {
	StoreRefreshToken(ctx context.Context, token *models.RefreshToken) (string, error)
	GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	UpdateSessionActivityByRefresh(ctx context.Context, refreshTokenID string, lastActive time.Time) error
	RevokeRefreshToken(ctx context.Context, id, reason string) error
	CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) (string, error)
	GetPasswordResetToken(ctx context.Context, hash string) (*models.PasswordResetToken, error)
	MarkPasswordResetUsed(ctx context.Context, id string) error
	CreateSession(ctx context.Context, session *models.UserSession) (string, error)
}

// OTPManager abstracts OTP generation/verification
type OTPManager interface {
	GenerateAndStore(ctx context.Context, userID sql.NullString, identifier, purpose, channel string) (string, error)
	Verify(ctx context.Context, identifier, purpose, code string) error
}

// AuthServiceOptions configures AuthService runtime behaviour
type AuthServiceOptions struct {
	Env        string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// AuthService provides authentication workflows
type AuthService struct {
	users      UserStore
	authRepo   AuthDataStore
	jwt        *auth.JWTService
	password   *auth.PasswordService
	otp        OTPManager
	notifier   NotificationSender
	env        string
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

// NewAuthService constructs a new AuthService
func NewAuthService(users UserStore, authRepo AuthDataStore, jwtService *auth.JWTService, passwordService *auth.PasswordService, otpService OTPManager, notifier NotificationSender, opts AuthServiceOptions) *AuthService {
	service := &AuthService{
		users:      users,
		authRepo:   authRepo,
		jwt:        jwtService,
		password:   passwordService,
		otp:        otpService,
		notifier:   notifier,
		env:        opts.Env,
		accessTTL:  opts.AccessTTL,
		refreshTTL: opts.RefreshTTL,
		now:        time.Now,
	}
	return service
}

// RegisterInput captures required registration payload
type RegisterInput struct {
	Email    string
	Password string
	Phone    string
}

// RegisterResult contains registration outcome metadata
type RegisterResult struct {
	UserID               string
	RequiresVerification bool
	DevelopmentOTP       string
}

// LoginInput captures login request parameters
type LoginInput struct {
	Email      string
	Password   string
	UserAgent  string
	IPAddress  string
	DeviceID   string
	DeviceName string
}

// TokenPair groups access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// LoginResult returns tokens and authenticated user model
type LoginResult struct {
	Tokens TokenPair
	User   *models.User
}

// VerifyOTPInput captures OTP verification payload
type VerifyOTPInput struct {
	UserID  string
	Purpose string
	Code    string
}

// ResendOTPInput captures resend payload
type ResendOTPInput struct {
	UserID  string
	Purpose string
}

// RefreshInput captures refresh token requests
type RefreshInput struct {
	RefreshToken string
}

// RefreshResult returns refreshed tokens and user info
type RefreshResult struct {
	Tokens TokenPair
	User   *models.User
}

// PasswordResetRequestInput payload
type PasswordResetRequestInput struct {
	Email string
}

// PasswordResetRequestResult returns metadata for reset initiation
type PasswordResetRequestResult struct {
	DevelopmentToken string
}

// PasswordResetConfirmInput payload
type PasswordResetConfirmInput struct {
	Token       string
	NewPassword string
}

// LogoutInput payload
type LogoutInput struct {
	RefreshToken string
}

const (
	lockThreshold   = 5
	lockDuration    = 15 * time.Minute
	defaultUserRole = "user"
)

// Register handles new user creation and OTP dispatch
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if !utils.IsValidEmail(email) {
		return nil, errors.New("invalid email format")
	}

	phone := strings.TrimSpace(input.Phone)
	var normalizedPhone string
	if phone != "" {
		if utils.IsValidKenyaPhone(phone) {
			normalizedPhone = utils.NormalizeKenyaPhone(phone)
		} else if utils.IsValidPhone(phone) {
			normalizedPhone = utils.NormalizePhone(phone, "254")
		} else {
			return nil, errors.New("invalid phone number")
		}
	}

	if err := s.password.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	} else if err != nil && err != repository.ErrNotFound {
		return nil, err
	}

	if normalizedPhone != "" {
		if _, err := s.users.GetByPhone(ctx, normalizedPhone); err == nil {
			return nil, ErrPhoneExists
		} else if err != nil && err != repository.ErrNotFound {
			return nil, err
		}
	}

	passwordHash, err := s.password.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	phoneValue := sql.NullString{Valid: normalizedPhone != "", String: normalizedPhone}

	user := &models.User{
		Email:         email,
		Phone:         phoneValue,
		PasswordHash:  passwordHash,
		Role:          defaultUserRole,
		Status:        "pending_verification",
		EmailVerified: false,
		PhoneVerified: false,
	}

	userID, err := s.users.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	purpose, channel, identifier, err := s.resolveOTPTarget(user, determineOTPPurpose(normalizedPhone))
	if err != nil {
		return nil, err
	}

	otpCode, err := s.otp.GenerateAndStore(ctx, sql.NullString{Valid: true, String: userID}, identifier, purpose, channel)
	if err != nil {
		return nil, err
	}
	s.sendOTPNotification(ctx, channel, identifier, otpCode)

	result := &RegisterResult{UserID: userID, RequiresVerification: true}
	if s.env == "development" {
		result.DevelopmentOTP = otpCode
	}
	return result, nil
}

// VerifyOTP validates OTP and updates verification flags
func (s *AuthService) VerifyOTP(ctx context.Context, input VerifyOTPInput) error {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}

	purpose := input.Purpose
	identifier := user.Email
	switch purpose {
	case models.OTPPurposePhoneVerification:
		if !user.Phone.Valid {
			return errors.New("phone number missing")
		}
		identifier = user.Phone.String
	case models.OTPPurposeEmailVerification:
		identifier = user.Email
	default:
		identifier = user.Email
	}

	if err := s.otp.Verify(ctx, identifier, purpose, strings.TrimSpace(input.Code)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidOTP
		}
		return err
	}

	emailVerified := user.EmailVerified
	phoneVerified := user.PhoneVerified

	switch purpose {
	case models.OTPPurposePhoneVerification:
		phoneVerified = true
	case models.OTPPurposeEmailVerification:
		emailVerified = true
	}

	status := user.Status
	if emailVerified && (phoneVerified || !user.Phone.Valid) {
		status = "active"
	}

	if err := s.users.UpdateVerification(ctx, user.ID, emailVerified, phoneVerified, status); err != nil {
		return err
	}

	return nil
}

// ResendOTP regenerates and sends OTP for verification
func (s *AuthService) ResendOTP(ctx context.Context, input ResendOTPInput) error {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}

	purpose, channel, identifier, err := s.resolveOTPTarget(user, input.Purpose)
	if err != nil {
		return err
	}

	code, err := s.otp.GenerateAndStore(ctx, sql.NullString{Valid: true, String: user.ID}, identifier, purpose, channel)
	if err != nil {
		return err
	}

	s.sendOTPNotification(ctx, channel, identifier, code)
	return nil
}

// Login authenticates user credentials and issues JWT tokens
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	now := s.now()
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(now) {
		return nil, ErrAccountLocked
	}

	if user.Status != "active" || !user.EmailVerified || (user.Phone.Valid && !user.PhoneVerified) {
		return nil, ErrVerificationRequired
	}

	if !s.password.CheckPassword(input.Password, user.PasswordHash) {
		lockUntil := sql.NullTime{}
		if user.FailedAttempts+1 >= lockThreshold {
			lockUntil = sql.NullTime{Valid: true, Time: now.Add(lockDuration)}
		}
		_ = s.users.IncrementFailedAttempts(ctx, user.ID, lockUntil)
		return nil, ErrInvalidCredentials
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		UserAgent: sql.NullString{Valid: input.UserAgent != "", String: input.UserAgent},
		IPAddress: sql.NullString{Valid: input.IPAddress != "", String: input.IPAddress},
		IssuedAt:  now,
		ExpiresAt: now.Add(s.refreshTTL),
	}

	refreshID, err := s.authRepo.StoreRefreshToken(ctx, refreshModel)
	if err != nil {
		return nil, err
	}

	session := &models.UserSession{
		UserID:         user.ID,
		RefreshTokenID: sql.NullString{Valid: true, String: refreshID},
		DeviceID:       sql.NullString{Valid: input.DeviceID != "", String: input.DeviceID},
		DeviceName:     sql.NullString{Valid: input.DeviceName != "", String: input.DeviceName},
		IPAddress:      sql.NullString{Valid: input.IPAddress != "", String: input.IPAddress},
		UserAgent:      sql.NullString{Valid: input.UserAgent != "", String: input.UserAgent},
	}
	_, _ = s.authRepo.CreateSession(ctx, session)

	tokens := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTTL.Seconds()),
	}

	return &LoginResult{Tokens: tokens, User: user}, nil
}

// RefreshTokens validates a refresh token and issues a new access token
func (s *AuthService) RefreshTokens(ctx context.Context, input RefreshInput) (*RefreshResult, error) {
	claims, err := s.jwt.ValidateToken(input.RefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidRefreshToken
	}

	refreshHash := hashToken(input.RefreshToken)
	storedToken, err := s.authRepo.GetRefreshTokenByHash(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	now := s.now()
	if storedToken.RevokedAt.Valid || storedToken.ExpiresAt.Before(now) {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.users.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrVerificationRequired
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	_ = s.authRepo.UpdateSessionActivityByRefresh(ctx, storedToken.ID, now)

	tokens := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: input.RefreshToken,
		ExpiresIn:    int(s.accessTTL.Seconds()),
	}

	return &RefreshResult{Tokens: tokens, User: user}, nil
}

// Logout revokes a refresh token
func (s *AuthService) Logout(ctx context.Context, input LogoutInput) error {
	refreshHash := hashToken(input.RefreshToken)
	stored, err := s.authRepo.GetRefreshTokenByHash(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidRefreshToken
		}
		return err
	}

	if stored.RevokedAt.Valid {
		return nil
	}

	if err := s.authRepo.RevokeRefreshToken(ctx, stored.ID, "user_logout"); err != nil {
		return err
	}

	return nil
}

// RequestPasswordReset creates a password reset token (development mode returns raw token)
func (s *AuthService) RequestPasswordReset(ctx context.Context, input PasswordResetRequestInput) (*PasswordResetRequestResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Avoid leaking existence
			return &PasswordResetRequestResult{}, nil
		}
		return nil, err
	}

	rawToken, err := randomString(32)
	if err != nil {
		return nil, err
	}

	token := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: s.now().Add(1 * time.Hour),
	}

	if _, err := s.authRepo.CreatePasswordResetToken(ctx, token); err != nil {
		return nil, err
	}

	s.sendPasswordResetNotification(ctx, user.Email, rawToken)

	result := &PasswordResetRequestResult{}
	if s.env == "development" {
		result.DevelopmentToken = rawToken
	}
	return result, nil
}

// ResetPassword verifies reset token and updates password
func (s *AuthService) ResetPassword(ctx context.Context, input PasswordResetConfirmInput) error {
	if err := s.password.ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	hash := hashToken(strings.TrimSpace(input.Token))
	token, err := s.authRepo.GetPasswordResetToken(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidOTP
		}
		return err
	}

	if token.UsedAt.Valid || token.ExpiresAt.Before(s.now()) {
		return ErrInvalidOTP
	}

	newHash, err := s.password.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePassword(ctx, token.UserID, newHash); err != nil {
		return err
	}

	if err := s.authRepo.MarkPasswordResetUsed(ctx, token.ID); err != nil {
		return err
	}

	return nil
}

func hashToken(token string) string {
	sum := sha256Sum(token)
	return hex.EncodeToString(sum)
}

func sha256Sum(value string) []byte {
	h := sha256.New()
	h.Write([]byte(value))
	return h.Sum(nil)
}

func randomString(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *AuthService) resolveOTPTarget(user *models.User, requestedPurpose string) (string, string, string, error) {
	purpose := requestedPurpose
	channel := "email"
	identifier := user.Email

	switch requestedPurpose {
	case models.OTPPurposePhoneVerification:
		if !user.Phone.Valid {
			return "", "", "", errors.New("phone number missing for OTP")
		}
		channel = "sms"
		identifier = user.Phone.String
	case models.OTPPurposeEmailVerification, "":
		purpose = models.OTPPurposeEmailVerification
	default:
		if user.Phone.Valid {
			purpose = models.OTPPurposePhoneVerification
			channel = "sms"
			identifier = user.Phone.String
		} else {
			purpose = models.OTPPurposeEmailVerification
			channel = "email"
			identifier = user.Email
		}
	}

	return purpose, channel, identifier, nil
}

func determineOTPPurpose(normalizedPhone string) string {
	if normalizedPhone != "" {
		return models.OTPPurposePhoneVerification
	}
	return models.OTPPurposeEmailVerification
}

func (s *AuthService) sendOTPNotification(ctx context.Context, channel, identifier, code string) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.SendOTP(ctx, channel, identifier, code); err != nil {
		log.Printf("failed to dispatch OTP notification: %v", err)
	}
}

func (s *AuthService) sendPasswordResetNotification(ctx context.Context, email, token string) {
	if s.notifier == nil {
		return
	}
	if err := s.notifier.SendPasswordReset(ctx, email, token); err != nil {
		log.Printf("failed to send password reset email: %v", err)
	}
}

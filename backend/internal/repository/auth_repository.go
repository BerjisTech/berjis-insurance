// auth_repository.go contains persistence helpers for OTP codes, refresh tokens, sessions, and password reset tokens
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/insurance-broker/backend/internal/database"
	"github.com/insurance-broker/backend/internal/models"
)

// AuthRepository aggregates authentication-related persistence concerns
type AuthRepository struct {
	db *database.DB
}

// NewAuthRepository constructs a repository instance
func NewAuthRepository(db *database.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateOTP stores a new OTP entry
func (r *AuthRepository) CreateOTP(ctx context.Context, otp *models.OTPCode) (string, error) {
	const query = `
        INSERT INTO otp_codes (user_id, identifier, code, purpose, channel, max_attempts, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at`

	var userID interface{}
	if otp.UserID.Valid {
		userID = otp.UserID.String
	} else {
		userID = nil
	}

	if err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		otp.Identifier,
		otp.Code,
		otp.Purpose,
		otp.Channel,
		otp.MaxAttempts,
		otp.ExpiresAt,
	).Scan(&otp.ID, &otp.CreatedAt); err != nil {
		return "", err
	}

	return otp.ID, nil
}

// GetActiveOTP fetches an unexpired OTP for identifier+purpose
func (r *AuthRepository) GetActiveOTP(ctx context.Context, identifier, purpose string) (*models.OTPCode, error) {
	const query = `
        SELECT id, user_id, identifier, code, purpose, channel, attempts, max_attempts,
               expires_at, verified_at, created_at
        FROM otp_codes
        WHERE identifier = $1
          AND purpose = $2
          AND verified_at IS NULL
          AND expires_at > CURRENT_TIMESTAMP
        ORDER BY created_at DESC
        LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, identifier, purpose)
	var otp models.OTPCode
	if err := row.Scan(
		&otp.ID,
		&otp.UserID,
		&otp.Identifier,
		&otp.Code,
		&otp.Purpose,
		&otp.Channel,
		&otp.Attempts,
		&otp.MaxAttempts,
		&otp.ExpiresAt,
		&otp.VerifiedAt,
		&otp.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &otp, nil
}

// IncrementOTPAttempts increments attempt count and returns updated value
func (r *AuthRepository) IncrementOTPAttempts(ctx context.Context, id string) (int, error) {
	const query = `
        UPDATE otp_codes
        SET attempts = attempts + 1
        WHERE id = $1
        RETURNING attempts`

	var attempts int
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&attempts); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return attempts, nil
}

// MarkOTPVerified marks OTP as used
func (r *AuthRepository) MarkOTPVerified(ctx context.Context, id string) error {
	const query = `
        UPDATE otp_codes
        SET verified_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

// CleanupIdentifier removes OTPs for identifier+purpose
func (r *AuthRepository) DeleteOTPs(ctx context.Context, identifier, purpose string) error {
	const query = `
        DELETE FROM otp_codes
        WHERE identifier = $1 AND purpose = $2`
	_, err := r.db.ExecContext(ctx, query, identifier, purpose)
	return err
}

// CleanupExpiredOTPs removes expired or max-attempt OTP rows
func (r *AuthRepository) CleanupExpiredOTPs(ctx context.Context) error {
	const query = `
        DELETE FROM otp_codes
        WHERE expires_at <= CURRENT_TIMESTAMP
           OR attempts >= max_attempts`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// StoreRefreshToken saves refresh token hash for user session
func (r *AuthRepository) StoreRefreshToken(ctx context.Context, token *models.RefreshToken) (string, error) {
	const query = `
        INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, issued_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at`

	var ua interface{}
	if token.UserAgent.Valid {
		ua = token.UserAgent.String
	} else {
		ua = nil
	}
	var ip interface{}
	if token.IPAddress.Valid {
		ip = token.IPAddress.String
	} else {
		ip = nil
	}

	if err := r.db.QueryRowContext(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		ua,
		ip,
		token.IssuedAt,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt); err != nil {
		return "", err
	}

	return token.ID, nil
}

// GetRefreshTokenByHash fetches refresh token by hash (used during refresh)
func (r *AuthRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	const query = `
        SELECT id, user_id, token_hash, user_agent, ip_address, issued_at, expires_at,
               revoked_at, revoked_reason, created_at
        FROM refresh_tokens
        WHERE token_hash = $1`

	row := r.db.QueryRowContext(ctx, query, hash)
	var token models.RefreshToken
	if err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.UserAgent,
		&token.IPAddress,
		&token.IssuedAt,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.RevokedReason,
		&token.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

// RevokeRefreshToken flags a token as revoked and optionally records reason
func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id, reason string) error {
	const query = `
        UPDATE refresh_tokens
        SET revoked_at = CURRENT_TIMESTAMP,
            revoked_reason = NULLIF($2, '')
        WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, reason)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

// CreatePasswordResetToken stores hashed token for password reset
func (r *AuthRepository) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) (string, error) {
	const query = `
        INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
        VALUES ($1, $2, $3)
        RETURNING id, created_at`

	if err := r.db.QueryRowContext(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt); err != nil {
		return "", err
	}
	return token.ID, nil
}

// GetPasswordResetToken fetches a reset token by hash (ensuring unused)
func (r *AuthRepository) GetPasswordResetToken(ctx context.Context, hash string) (*models.PasswordResetToken, error) {
	const query = `
        SELECT id, user_id, token_hash, expires_at, used_at, created_at
        FROM password_reset_tokens
        WHERE token_hash = $1`

	row := r.db.QueryRowContext(ctx, query, hash)
	var token models.PasswordResetToken
	if err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

// MarkPasswordResetUsed flags token as consumed
func (r *AuthRepository) MarkPasswordResetUsed(ctx context.Context, id string) error {
	const query = `
        UPDATE password_reset_tokens
        SET used_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSession stores a user session record for auditing
func (r *AuthRepository) CreateSession(ctx context.Context, session *models.UserSession) (string, error) {
	const query = `
        INSERT INTO user_sessions (user_id, refresh_token_id, device_id, device_name, ip_address, location, user_agent)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, last_active_at`

	var refreshID interface{}
	if session.RefreshTokenID.Valid {
		refreshID = session.RefreshTokenID.String
	} else {
		refreshID = nil
	}

	if err := r.db.QueryRowContext(
		ctx,
		query,
		session.UserID,
		refreshID,
		nullableValue(session.DeviceID),
		nullableValue(session.DeviceName),
		nullableValue(session.IPAddress),
		nullableValue(session.Location),
		nullableValue(session.UserAgent),
	).Scan(&session.ID, &session.CreatedAt, &session.LastActiveAt); err != nil {
		return "", err
	}
	return session.ID, nil
}

// UpdateSessionActivity updates timestamps for a session
func (r *AuthRepository) UpdateSessionActivity(ctx context.Context, id string, lastActive time.Time) error {
	const query = `
        UPDATE user_sessions
        SET last_active_at = $2
        WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, lastActive)
	return err
}

// UpdateSessionActivityByRefresh updates activity timestamp for the session linked to a refresh token
func (r *AuthRepository) UpdateSessionActivityByRefresh(ctx context.Context, refreshTokenID string, lastActive time.Time) error {
	const query = `
        UPDATE user_sessions
        SET last_active_at = $2
        WHERE refresh_token_id = $1`

	_, err := r.db.ExecContext(ctx, query, refreshTokenID, lastActive)
	return err
}

// DeactivateSession marks a session as inactive
func (r *AuthRepository) DeactivateSession(ctx context.Context, id string) error {
	const query = `
        UPDATE user_sessions
        SET is_active = FALSE,
            last_active_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func nullableValue(value sql.NullString) interface{} {
	if value.Valid {
		return value.String
	}
	return nil
}

// Package repository provides data-access helpers
// user_repository.go manages CRUD operations for the users table
package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/insurance-broker/backend/internal/database"
	"github.com/insurance-broker/backend/internal/models"
)

// UserRepository encapsulates queries for the users table
// Keeps SQL centralized to simplify future optimizations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a repository instance
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user row. Expects email to be unique (case-insensitive).
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (string, error) {
	const query = `
        INSERT INTO users (email, phone, password_hash, role, status, email_verified, phone_verified)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at`

	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))

	var phone interface{}
	if user.Phone.Valid {
		phone = user.Phone.String
	} else {
		phone = nil
	}

	var id string
	if err := r.db.QueryRowContext(
		ctx,
		query,
		normalizedEmail,
		phone,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.EmailVerified,
		user.PhoneVerified,
	).Scan(&id, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return "", ErrDuplicate
			}
		}
		return "", err
	}

	user.ID = id
	user.Email = normalizedEmail
	return id, nil
}

// GetByEmail fetches a user record by email (case-insensitive)
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
        SELECT id, email, phone, password_hash, role, status, email_verified, phone_verified,
               last_login_at, failed_login_attempts, locked_until, created_at, updated_at
        FROM users
        WHERE LOWER(email) = LOWER($1)
        LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, strings.TrimSpace(email))
	return scanUser(row)
}

// GetByPhone fetches a user record by normalized phone number
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	const query = `
        SELECT id, email, phone, password_hash, role, status, email_verified, phone_verified,
               last_login_at, failed_login_attempts, locked_until, created_at, updated_at
        FROM users
        WHERE phone = $1
        LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, phone)
	return scanUser(row)
}

// GetByID returns a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	const query = `
        SELECT id, email, phone, password_hash, role, status, email_verified, phone_verified,
               last_login_at, failed_login_attempts, locked_until, created_at, updated_at
        FROM users
        WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanUser(row)
}

// UpdateVerification toggles verification flags and status
func (r *UserRepository) UpdateVerification(ctx context.Context, id string, emailVerified, phoneVerified bool, status string) error {
	const query = `
        UPDATE users
        SET email_verified = $2,
            phone_verified = $3,
            status = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, emailVerified, phoneVerified, status)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateLastLogin stores the last_login_at timestamp and resets failed attempts
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string, ts time.Time) error {
	const query = `
        UPDATE users
        SET last_login_at = $2,
            failed_login_attempts = 0,
            locked_until = NULL,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, ts)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

// IncrementFailedAttempts increments failed login counter and optionally locks account
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, id string, lockUntil sql.NullTime) error {
	const query = `
        UPDATE users
        SET failed_login_attempts = failed_login_attempts + 1,
            locked_until = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	if _, err := r.db.ExecContext(ctx, query, id, lockUntil); err != nil {
		return err
	}
	return nil
}

// ResetPassword updates password hash and clears resets
func (r *UserRepository) UpdatePassword(ctx context.Context, id, newHash string) error {
	const query = `
        UPDATE users
        SET password_hash = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id, newHash)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func scanUser(row *sql.Row) (*models.User, error) {
	var user models.User
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.LastLoginAt,
		&user.FailedAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

package repositories

import (
	"context"
	"budget-planner/internal/common/errors"
	"budget-planner/internal/domain/user"
	"budget-planner/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository implements the user.Repository interface
type PostgresUserRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgresUserRepository creates a new PostgreSQL-backed user repository
func NewPostgresUserRepository(pool *pgxpool.Pool, logger *logger.Logger) user.Repository {
	return &PostgresUserRepository{
		pool:   pool,
		logger: logger,
	}
}

// BeginTransaction starts a new database transaction
func (r *PostgresUserRepository) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, errors.NewInfraTransactionError("begin transaction", err)
	}
	return tx, nil
}

// CommitTransaction commits the given transaction
func (r *PostgresUserRepository) CommitTransaction(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

// RollbackTransaction rolls back the given transaction
func (r *PostgresUserRepository) RollbackTransaction(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

// UsernameExists checks if a username exists
func (r *PostgresUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	const query = "SELECT EXISTS(SELECT 1 FROM user_schema.users WHERE username = $1)"
	var exists bool
	err := r.pool.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseError("checking username existence", err)
	}
	return exists, nil
}

// EmailExists checks if an email exists
func (r *PostgresUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	const query = "SELECT EXISTS(SELECT 1 FROM user_schema.users WHERE email = $1)"
	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseError("checking email existence", err)
	}
	return exists, nil
}

// CreateUser creates a new user
func (r *PostgresUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	const query = `
		INSERT INTO user_schema.users (
			id, username, email, password_hash, status, failed_login_attempts, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash, u.Status, u.FailedLoginAttempts, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return errors.NewDatabaseError("creating user", err)
	}
	return nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	const query = `
		SELECT id, username, email, password_hash, status, verified_at, last_login_at,
		       failed_login_attempts, created_at, updated_at
		FROM user_schema.users
		WHERE id = $1
	`

	u := &user.User{}
	var verifiedAt, lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Status,
		&verifiedAt, &lastLoginAt, &u.FailedLoginAttempts, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("user not found", map[string]interface{}{"id": id})
		}
		return nil, errors.NewDatabaseError("fetching user", err)
	}

	u.VerifiedAt = verifiedAt
	u.LastLoginAt = lastLoginAt
	return u, nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	const query = `
		SELECT id, username, email, password_hash, status, verified_at, last_login_at,
		       failed_login_attempts, created_at, updated_at
		FROM user_schema.users
		WHERE email = $1
	`

	u := &user.User{}
	var verifiedAt, lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Status,
		&verifiedAt, &lastLoginAt, &u.FailedLoginAttempts, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("user not found", map[string]interface{}{"email": email})
		}
		return nil, errors.NewDatabaseError("fetching user", err)
	}

	u.VerifiedAt = verifiedAt
	u.LastLoginAt = lastLoginAt
	return u, nil
}

// GetUserByUsername retrieves a user by username
func (r *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	const query = `
		SELECT id, username, email, password_hash, status, verified_at, last_login_at,
		       failed_login_attempts, created_at, updated_at
		FROM user_schema.users
		WHERE username = $1
	`

	u := &user.User{}
	var verifiedAt, lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Status,
		&verifiedAt, &lastLoginAt, &u.FailedLoginAttempts, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("user not found", map[string]interface{}{"username": username})
		}
		return nil, errors.NewDatabaseError("fetching user", err)
	}

	u.VerifiedAt = verifiedAt
	u.LastLoginAt = lastLoginAt
	return u, nil
}

// UpdateUser updates an existing user
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	const query = `
		UPDATE user_schema.users
		SET username = $2, email = $3, password_hash = $4, status = $5,
		    verified_at = $6, last_login_at = $7, failed_login_attempts = $8, updated_at = $9
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash, u.Status,
		u.VerifiedAt, u.LastLoginAt, u.FailedLoginAttempts, u.UpdatedAt)
	if err != nil {
		return errors.NewDatabaseError("updating user", err)
	}
	return nil
}

// CreatePasswordResetToken creates a password reset token
func (r *PostgresUserRepository) CreatePasswordResetToken(ctx context.Context, token *user.PasswordResetToken) error {
	const query = `
		INSERT INTO user_schema.password_reset_tokens (user_id, token, expires_at, is_used, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query, token.UserID, token.Token, token.ExpiresAt, token.IsUsed, token.CreatedAt)
	if err != nil {
		return errors.NewDatabaseError("creating password reset token", err)
	}
	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (r *PostgresUserRepository) GetPasswordResetToken(ctx context.Context, token string) (*user.PasswordResetToken, error) {
	const query = `
		SELECT user_id, token, expires_at, is_used, created_at
		FROM user_schema.password_reset_tokens
		WHERE token = $1
	`

	resetToken := &user.PasswordResetToken{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&resetToken.UserID, &resetToken.Token, &resetToken.ExpiresAt, &resetToken.IsUsed, &resetToken.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFoundError("password reset token not found", map[string]interface{}{"token": token})
		}
		return nil, errors.NewDatabaseError("fetching password reset token", err)
	}
	return resetToken, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (r *PostgresUserRepository) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	const query = `UPDATE user_schema.password_reset_tokens SET is_used = true WHERE token = $1`
	_, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return errors.NewDatabaseError("marking password reset token as used", err)
	}
	return nil
}

// DeleteOtherPasswordResetTokens deletes all other password reset tokens for a user
func (r *PostgresUserRepository) DeleteOtherPasswordResetTokens(ctx context.Context, userID uuid.UUID) error {
	const query = `DELETE FROM user_schema.password_reset_tokens WHERE user_id = $1 AND is_used = false`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return errors.NewDatabaseError("deleting password reset tokens", err)
	}
	return nil
}

// UpdatePassword updates a user's password
func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const query = `UPDATE user_schema.users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, passwordHash, time.Now())
	if err != nil {
		return errors.NewDatabaseError("updating password", err)
	}
	return nil
}

// RecordLogin records a user login
func (r *PostgresUserRepository) RecordLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	const query = `UPDATE user_schema.users SET last_login_at = $2, updated_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, now, now)
	if err != nil {
		return errors.NewDatabaseError("recording login", err)
	}
	return nil
}

// IncrementFailedLoginAttempts increments failed login attempts
func (r *PostgresUserRepository) IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE user_schema.users SET failed_login_attempts = failed_login_attempts + 1, updated_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return errors.NewDatabaseError("incrementing failed login attempts", err)
	}
	return nil
}

// ResetFailedLoginAttempts resets failed login attempts
func (r *PostgresUserRepository) ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE user_schema.users SET failed_login_attempts = 0, updated_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return errors.NewDatabaseError("resetting failed login attempts", err)
	}
	return nil
}


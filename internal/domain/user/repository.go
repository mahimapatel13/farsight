package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository defines the data access interface for users
type Repository interface {
	// Transaction management
	BeginTransaction(ctx context.Context) (pgx.Tx, error)
	CommitTransaction(ctx context.Context, tx pgx.Tx) error
	RollbackTransaction(ctx context.Context, tx pgx.Tx) error

	// User operations (Verification)
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)

	// User operations (CRUD)
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error

	// Password / Authentication operations
	CreatePasswordResetToken(ctx context.Context, resetToken *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, token string) error
	DeleteOtherPasswordResetTokens(ctx context.Context, userID uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// Login management
	RecordLogin(ctx context.Context, id uuid.UUID) error

	// Failed Login Attempt management
	IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) error
	ResetFailedLoginAttempts(ctx context.Context, id uuid.UUID) error
}


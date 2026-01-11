package user

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a user account
type Status string

const (
	StatusActivated   Status = "activated"
	StatusDeactivated Status = "deactivated"
	StatusSuspended   Status = "suspended"
	StatusPending     Status = "pending"
	StatusLocked      Status = "locked"
)

// User represents a user account in the budget planner app
type User struct {
	ID                  uuid.UUID
	Username            string
	Email               string
	PasswordHash        string
	Status              Status
	VerifiedAt          *time.Time
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// CreateUserRequest represents data needed to create a new user
type CreateUserRequest struct {
	Username string
	Email    string
}

// LoginRequest represents the credentials needed for login
type LoginRequest struct {
	Username string
	Email    string
	Password string
}

// PasswordResetRequest represents data needed to request password reset
type PasswordResetRequest struct {
	Email string
}

// PasswordResetConfirmation represents data needed to confirm password reset
type PasswordResetConfirmation struct {
	Token       string
	NewPassword string
}

// PasswordResetToken stores information for password reset functionality
type PasswordResetToken struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	IsUsed    bool
	CreatedAt time.Time
}


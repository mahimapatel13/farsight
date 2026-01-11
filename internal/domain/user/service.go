package user

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"budget-planner/internal/common/errors"
	"budget-planner/internal/domain/email"
	"budget-planner/pkg/logger"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines the business logic for users
type Service interface {
	RegisterUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	AuthenticateUser(ctx context.Context, req *LoginRequest) (*User, error)
	RequestPasswordReset(ctx context.Context, req *PasswordResetRequest) (string, error)
	ConfirmPasswordReset(ctx context.Context, req *PasswordResetConfirmation) error
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
}

// service is the concrete implementation of the Service interface
type service struct {
	repo         Repository
	emailService email.EmailService
	logger       *logger.Logger
}

// NewService creates a new user service
func NewService(
	repo Repository,
	emailService email.EmailService,
	logger *logger.Logger,
) Service {
	return &service{
		repo:         repo,
		emailService: emailService,
		logger:       logger,
	}
}

// generateRandomPassword generates a random password with the specified length
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(password)
}

func sanitizeUsername(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(input, "")
}

func (s *service) generateUniqueUsername(ctx context.Context, baseUsername string) (string, error) {
	username := sanitizeUsername(baseUsername)
	if username == "" {
		username = "user"
	}
	suffix := 1
	for {
		exists, err := s.repo.UsernameExists(ctx, username)
		if err != nil && !errors.IsNotFoundErrorDomain(err) {
			return "", err
		}
		if !exists {
			return username, nil
		}
		username = fmt.Sprintf("%s%d", sanitizeUsername(baseUsername), suffix)
		suffix++
	}
}

// RegisterUser creates a new user account
func (s *service) RegisterUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	s.logger.Debug("Starting user registration", "username", req.Username, "email", req.Email)

	// Check if email exists
	emailExists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil && !errors.IsNotFoundErrorDomain(err) {
		s.logger.Error("Failed to check email existence", "email", req.Email, "error", err)
		return nil, errors.NewDatabaseError("fetching email", err)
	}
	if emailExists {
		s.logger.Warn("Email already exists", "email", req.Email)
		return nil, errors.NewConflictError("email", map[string]interface{}{"email": req.Email})
	}

	// Generate unique username
	uniqueUsername, err := s.generateUniqueUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("Failed to generate unique username", "baseUsername", req.Username, "error", err)
		return nil, errors.NewDatabaseError("generating unique username", err)
	}
	req.Username = uniqueUsername

	// Generate system-generated password for first login
	systemPassword := generateRandomPassword(12)
	s.logger.Info("Generated system password for user", "email", req.Email)

	// Hash system-generated password securely
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(systemPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "username", req.Username, "error", err)
		return nil, errors.NewBusinessError("PASSWORD_HASHING_FAILED", "password hashing failed", nil)
	}

	now := time.Now()
	user := &User{
		ID:                  uuid.New(),
		Username:            req.Username,
		Email:               req.Email,
		PasswordHash:        string(passwordHash),
		Status:              StatusPending,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Save user to database
	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to create user", "username", req.Username, "error", err)
		return nil, errors.NewBusinessError("USER_CREATION_FAILED", "failed to create user", nil)
	}

	// Send verification email with password
	err = s.emailService.SendVerificationEmail(ctx, user.Username, user.Email, systemPassword)
	if err != nil {
		s.logger.Warn("Failed to send verification email", "email", user.Email, "error", err)
		// Don't fail registration if email fails, but log it
	}

	s.logger.Info("User registered successfully", "username", req.Username, "userID", user.ID)
	return user, nil
}

// AuthenticateUser verifies login credentials and returns the user if valid
func (s *service) AuthenticateUser(ctx context.Context, req *LoginRequest) (*User, error) {
	s.logger.Debug("Authenticating user", "username", req.Username, "email", req.Email)

	var user *User
	var err error

	// Lookup by email or username
	switch {
	case req.Email != "":
		user, err = s.repo.GetUserByEmail(ctx, req.Email)
	case req.Username != "":
		user, err = s.repo.GetUserByUsername(ctx, req.Username)
	default:
		s.logger.Warn("Username and email not provided")
		return nil, errors.NewValidationError("username or email is required", map[string]any{"field": "username_and_email"})
	}

	if err != nil {
		if errors.IsNotFoundErrorDomain(err) {
			s.logger.Warn("Invalid credentials provided", "username", req.Username, "email", req.Email)
			return nil, errors.NewUnauthorizedError("invalid credentials")
		}
		s.logger.Error("Failed to fetch user", "error", err)
		return nil, errors.NewDatabaseError("error fetching user", err)
	}

	// Check if account is locked
	if user.Status == StatusLocked {
		s.logger.Warn("Account is locked", "userID", user.ID)
		return nil, errors.NewUnauthorizedError("account is locked")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		s.logger.Warn("Invalid password provided", "userID", user.ID)
		
		// Increment failed login attempts
		if incrementErr := s.repo.IncrementFailedLoginAttempts(ctx, user.ID); incrementErr != nil {
			s.logger.Error("Failed to increment failed login attempts", "error", incrementErr)
		}

		// Lock account after 5 failed attempts
		failedAttempts := user.FailedLoginAttempts + 1
		if failedAttempts >= 5 {
			user.Status = StatusLocked
			if updateErr := s.repo.UpdateUser(ctx, user); updateErr != nil {
				s.logger.Error("Failed to lock account", "error", updateErr)
			}
			return nil, errors.NewUnauthorizedError("account locked due to too many failed login attempts")
		}

		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Reset failed login attempts on successful login
	if user.FailedLoginAttempts > 0 {
		if resetErr := s.repo.ResetFailedLoginAttempts(ctx, user.ID); resetErr != nil {
			s.logger.Error("Failed to reset failed login attempts", "error", resetErr)
		}
	}

	// Update last login time
	if err := s.repo.RecordLogin(ctx, user.ID); err != nil {
		s.logger.Warn("Failed to record login", "error", err)
	}

	// Update status to activated if it was pending
	if user.Status == StatusPending {
		now := time.Now()
		user.Status = StatusActivated
		user.VerifiedAt = &now
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			s.logger.Warn("Failed to update user status", "error", err)
		}
	}

	s.logger.Info("User authenticated successfully", "userID", user.ID)
	return user, nil
}

// RequestPasswordReset initiates the password reset process
func (s *service) RequestPasswordReset(ctx context.Context, req *PasswordResetRequest) (string, error) {
	// Check if user exists for the given email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Do not reveal if email exists or not for security reasons
		s.logger.Info("password reset requested for non-existent email", "email", req.Email)
		return "", nil
	}

	// Store reset token with expiration (1 hour)
	expires := time.Now().Add(1 * time.Hour)
	token := generateRandomPassword(32)

	passwordResetToken := PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expires,
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreatePasswordResetToken(ctx, &passwordResetToken); err != nil {
		s.logger.Error("failed to save reset token", "error", err)
		return "", errors.NewBusinessError("RESET_TOKEN_SAVE_FAILED", "failed to initiate password reset", nil)
	}

	// Send reset link via email
	err = s.emailService.SendPasswordResetEmail(ctx, user.Email, token)
	if err != nil {
		s.logger.Error("failed to send password reset email", "error", err)
		return "", errors.NewBusinessError("EMAIL_SEND_FAILED", "failed to send password reset email", nil)
	}

	s.logger.Info("password reset token generated and email sent", "userID", user.ID, "email", user.Email)
	return token, nil
}

// ConfirmPasswordReset validates the reset token and updates the password
func (s *service) ConfirmPasswordReset(ctx context.Context, req *PasswordResetConfirmation) error {
	// Validate token
	resetToken, err := s.repo.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		return errors.NewDatabaseError("fetching reset token", err)
	}

	if resetToken.Token != req.Token {
		s.logger.Warn("Invalid password reset token", "token", req.Token, "userID", resetToken.UserID)
		return errors.NewUnauthorizedError("invalid password reset token")
	}

	// Check if token is expired
	if resetToken.ExpiresAt.Before(time.Now()) {
		s.logger.Warn("Password reset token expired", "userID", resetToken.UserID)
		return errors.NewUnauthorizedError("password reset token has expired")
	}

	// Check if token is already used
	if resetToken.IsUsed {
		s.logger.Warn("Password reset token already used", "userID", resetToken.UserID)
		return errors.NewUnauthorizedError("password reset token has already been used")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return errors.NewBusinessError("PASSWORD_HASH_FAILED", "failed to update password", nil)
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, resetToken.UserID, string(passwordHash)); err != nil {
		return errors.NewBusinessError("PASSWORD_UPDATE_FAILED", "failed to update password", nil)
	}

	// Mark token as used
	if err := s.repo.MarkPasswordResetTokenUsed(ctx, req.Token); err != nil {
		s.logger.Warn("failed to mark reset token as used", "error", err)
	}

	// Invalidate / Delete all other tokens
	if err := s.repo.DeleteOtherPasswordResetTokens(ctx, resetToken.UserID); err != nil {
		s.logger.Warn("failed to delete other reset tokens", "error", err)
	}

	s.logger.Info("Password reset successfully", "userID", resetToken.UserID)
	return nil
}

// GetUser retrieves a user by ID
func (s *service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to fetch user", "userID", id, "error", err)
		return nil, errors.NewDatabaseError("fetching user", err)
	}
	return user, nil
}


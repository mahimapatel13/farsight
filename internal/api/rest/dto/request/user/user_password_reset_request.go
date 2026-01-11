package user

// UserPasswordResetRequest represents data needed to request password reset
type UserPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}



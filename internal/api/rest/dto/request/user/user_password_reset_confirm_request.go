package user

// UserPasswordResetConfirmRequest represents data needed to confirm password reset
type UserPasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}



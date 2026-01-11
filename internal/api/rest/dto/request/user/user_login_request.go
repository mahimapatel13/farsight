package user

// UserLoginRequest represents the credentials needed for login
type UserLoginRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=30"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password" validate:"required"`
}



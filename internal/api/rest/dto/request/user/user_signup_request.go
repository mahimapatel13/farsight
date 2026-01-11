package user

// UserSignupRequest represents data needed to create a new user
type UserSignupRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
}



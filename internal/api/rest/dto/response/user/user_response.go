package user

import (
	"time"

	"github.com/google/uuid"
)

// UserSignupResponse represents the response for user signup
type UserSignupResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Message  string `json:"message"`
}

// UserLoginResponse represents the response for user login
type UserLoginResponse struct {
	User         UserInfo     `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Status    string     `json:"status"`
	LastLogin *time.Time `json:"last_login_at,omitempty"`
}



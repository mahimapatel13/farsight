package user

import (
	request "budget-planner/internal/api/rest/dto/request/user"
	response "budget-planner/internal/api/rest/dto/response/user"
	"budget-planner/internal/api/rest/middlewares"
	rest_utils "budget-planner/internal/api/rest/utils"
	"budget-planner/internal/common/errors"
	"budget-planner/internal/domain/user"
	"budget-planner/internal/infrastructure/auth"
	"budget-planner/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService user.Service
	jwtProvider *auth.JWTProvider
	logger      *logger.Logger
}

func NewUserHandler(
	userService user.Service,
	jwtProvider *auth.JWTProvider,
	log *logger.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtProvider: jwtProvider,
		logger:      log,
	}
}

// Signup creates a new user
func (h *UserHandler) Signup(c *gin.Context) {
	h.logger.Debug("Received request to signup a new user")

	req, ok := middlewares.GetRequestBody[request.UserSignupRequest](c)
	if !ok {
		h.logger.Warn("Invalid or missing request body for user signup")
		rest_utils.Error(c, errors.BadRequest("Request body not found or invalid", nil))
		return
	}

	h.logger.Debug("Signing up new user", "username", req.Username, "email", req.Email)

	userReq := user.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
	}

	u, err := h.userService.RegisterUser(c.Request.Context(), &userReq)
	if err != nil {
		h.logger.Error("Failed to create user", "username", req.Username, "email", req.Email, "error", err)
		rest_utils.Error(c, err)
		return
	}

	h.logger.Info("User registered successfully", "username", req.Username, "email", req.Email, "userID", u.ID)

	resp := response.UserSignupResponse{
		Username: u.Username,
		Email:    u.Email,
		Message:  "User created successfully. Kindly check your email to verify your account.",
	}

	rest_utils.Created(c, gin.H{"user": resp}, "User created successfully")
}

// Signin authenticates a user
func (h *UserHandler) Signin(c *gin.Context) {
	h.logger.Debug("Received request to signin a user")

	req, ok := middlewares.GetRequestBody[request.UserLoginRequest](c)
	if !ok {
		h.logger.Warn("Invalid or missing request body for user login")
		rest_utils.Error(c, errors.BadRequest("Request body not found or invalid", nil))
		return
	}

	h.logger.Debug("Attempting login", "username", req.Username, "email", req.Email)

	loginReq := user.LoginRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	u, err := h.userService.AuthenticateUser(c.Request.Context(), &loginReq)
	if err != nil {
		h.logger.Warn("Login failed: Invalid credentials", "username", req.Username, "email", req.Email, "error", err)
		rest_utils.Error(c, errors.Unauthorized("Invalid credentials"))
		return
	}

	// Generate JWT tokens (empty roles for now, can be extended later)
	tokens, err := h.jwtProvider.GenerateTokenPair(u.ID.String(), []string{})
	if err != nil {
		h.logger.Error("Failed to generate tokens", "error", err)
		rest_utils.Error(c, errors.InternalServerError(err))
		return
	}

	userInfo := response.UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Status:   string(u.Status),
	}
	if u.LastLoginAt != nil {
		userInfo.LastLogin = u.LastLoginAt
	}

	resp := response.UserLoginResponse{
		User:         userInfo,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}

	h.logger.Info("User logged in successfully", "userID", u.ID)
	rest_utils.Success(c, gin.H{"data": resp}, "Login successful")
}

// RequestPasswordReset initiates the password reset process
func (h *UserHandler) RequestPasswordReset(c *gin.Context) {
	req, ok := middlewares.GetRequestBody[request.UserPasswordResetRequest](c)
	if !ok {
		h.logger.Warn("Invalid or missing request body during password reset request")
		rest_utils.Error(c, errors.BadRequest("Request body not found or invalid", nil))
		return
	}

	resetReq := user.PasswordResetRequest{
		Email: req.Email,
	}

	_, err := h.userService.RequestPasswordReset(c.Request.Context(), &resetReq)
	if err != nil {
		h.logger.Error("Failed to request password reset", "email", req.Email, "error", err)
		rest_utils.Error(c, errors.InternalServerError(err))
		return
	}

	h.logger.Info("Password reset requested successfully", "email", req.Email)
	rest_utils.Success(c, gin.H{"message": "Password reset instructions sent"}, "Password reset instructions sent")
}

// ConfirmPasswordReset confirms and processes a password reset
func (h *UserHandler) ConfirmPasswordReset(c *gin.Context) {
	req, ok := middlewares.GetRequestBody[request.UserPasswordResetConfirmRequest](c)
	if !ok {
		h.logger.Warn("Invalid or missing request body during confirm password request")
		rest_utils.Error(c, errors.BadRequest("Request body not found or invalid", nil))
		return
	}

	resetReq := user.PasswordResetConfirmation{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}

	err := h.userService.ConfirmPasswordReset(c.Request.Context(), &resetReq)
	if err != nil {
		h.logger.Error("Failed to confirm password reset", "error", err)
		rest_utils.Error(c, errors.InternalServerError(err))
		return
	}

	h.logger.Info("Password reset successfully")
	rest_utils.Success(c, gin.H{"message": "Password reset successfully"}, "Password reset successfully")
}

// GetProfile retrieves the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Warn("User ID not found in context")
		rest_utils.Error(c, errors.Unauthorized("user not authenticated"))
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Warn("Invalid user ID in context")
		rest_utils.Error(c, errors.NewBusinessError("INVALID_USER_ID", "invalid user ID", nil))
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Warn("Failed to parse user ID", "userID", userIDStr, "error", err)
		rest_utils.Error(c, errors.NewBusinessError("INVALID_USER_ID", "invalid user ID", nil))
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userUUID)
	if err != nil {
		h.logger.Error("Failed to get user profile", "userID", userUUID, "error", err)
		rest_utils.Error(c, err)
		return
	}

	userInfo := response.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Status:   string(user.Status),
	}
	if user.LastLoginAt != nil {
		userInfo.LastLogin = user.LastLoginAt
	}

	h.logger.Info("User profile retrieved successfully", "userID", user.ID)
	rest_utils.Success(c, gin.H{"data": userInfo}, "Profile retrieved successfully")
}


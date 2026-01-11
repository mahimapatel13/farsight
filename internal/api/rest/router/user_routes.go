package router

import (
	request "budget-planner/internal/api/rest/dto/request/user"
	handler "budget-planner/internal/api/rest/handler/user"
	"budget-planner/internal/api/rest/middlewares"
	"budget-planner/internal/config"
	"budget-planner/internal/domain/email"
	"budget-planner/internal/domain/user"
	"budget-planner/internal/infrastructure/auth"
	"budget-planner/internal/infrastructure/database/postgres/repositories"
	"budget-planner/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterUserRoutes sets up all user-related routes
func RegisterUserRoutes(
	r *gin.RouterGroup,
	pool *pgxpool.Pool,
	logger *logger.Logger,
	cfg *config.Config,
	jwtProvider *auth.JWTProvider,
	emailService email.EmailService,
	authMiddleware *middlewares.AuthMiddleware,
) {
	// Create repository
	userRepo := repositories.NewPostgresUserRepository(pool, logger)

	// Create service
	userService := user.NewService(userRepo, emailService, logger)

	// Create handler
	userHandler := handler.NewUserHandler(userService, jwtProvider, logger)

	// Create routes
	api := r.Group("/user")

	// Public routes (No authentication required)
	api.POST(
		"/signup",
		middlewares.BindJSONMiddleware[request.UserSignupRequest](),
		userHandler.Signup,
	)

	api.POST(
		"/signin",
		middlewares.BindJSONMiddleware[request.UserLoginRequest](),
		userHandler.Signin,
	)

	api.POST(
		"/password-reset",
		middlewares.BindJSONMiddleware[request.UserPasswordResetRequest](),
		userHandler.RequestPasswordReset,
	)

	api.POST(
		"/confirm-password-reset",
		middlewares.BindJSONMiddleware[request.UserPasswordResetConfirmRequest](),
		userHandler.ConfirmPasswordReset,
	)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(authMiddleware.JWTMiddleware())

	protected.GET("/profile", userHandler.GetProfile)
}


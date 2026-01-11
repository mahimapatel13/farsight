package router

import (
	// Go standard libraries
	"context"

	// Internal packages
	"budget-planner/internal/api/rest/middlewares"
	"budget-planner/internal/config"
	"budget-planner/internal/domain/email"
	"budget-planner/internal/domain/integration"

    worker "budget-planner/internal/worker/email"

	"budget-planner/internal/infrastructure/auth"
	"budget-planner/internal/infrastructure/database/postgres/repositories"

	"budget-planner/pkg/email/queue"
	"budget-planner/pkg/logger"

	// External packages
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterRoutes sets up all API routes
func RegisterRoutes(
	r *gin.Engine,
	pool *pgxpool.Pool,
	logger *logger.Logger,
	cfg *config.Config,
) {

	// Add Custom Global Middlewares

	// // Use your custom logging middlewares instead of gin.Logger()
	// r.Use(middlewares.LoggingMiddleware(logger))

	// // Use request ID middlewares to ensure consistent request tracking
	// r.Use(middlewares.RequestIDMiddleware())

	// API versioning
	v1 := r.Group("/api/v1")

	// ✅ Initialize EmailManager
	retryPolicy := queue.NewRetryPolicy(
		cfg.Integration.Email.MaxRetries,     // MaxRetries from config
		cfg.Integration.Email.RetryIntervals, // MaxRetries from config
		logger,                               // Logger instance
	)

	// 2️⃣ Initialize Email Manager (e.g., SMTP or AWS SES)
	emailManager, err := integration.NewEmailManager(
		cfg.Integration.Email,
		nil, // We'll set this after creating the queue
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize email manager", "error", err)
	}

	// 3️⃣ Initialize Email Queue (InMemory or Redis)
	emailQueue := queue.NewEmailQueue(
		emailManager.GetDefaultProvider(),
		retryPolicy,
		logger,
	)

	// Set the email queue in the email manager
	emailManager.SetEmailQueue(emailQueue)
	if err != nil {
		logger.Fatal("Failed to initialize EmailManager", "error", err)
	}

	// ✅ Set EmailQueue's provider after EmailManager is ready
	emailQueue.SetEmailService(emailManager.GetDefaultProvider())

	// 7️⃣ Start Email Worker
	emailWorker := worker.NewEmailWorker(
		emailManager,
		emailQueue,
		*retryPolicy,
		cfg.Integration.Email.MaxRetries, // MaxRetries from config
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerCount := 5 // Number of concurrent workers
	emailWorker.StartWorker(ctx, workerCount)

	// ===============================
	// ✅ Create/ Initialize/ Inject Repositories
	// ===============================
	templateRepo := repositories.NewPostgresTemplateRepository(pool, logger)
	// ===============================
	// ✅ Create Initialize/ Inject Services
	// ===============================
	emailService := email.NewEmailService(
		emailManager,
		templateRepo,
		logger,
	)

	
	// Create JWT provider
	jwtProvider := auth.NewJWTProvider(
		cfg.Credentials.JWTAccessSecret,
		cfg.Credentials.JWTRefreshSecret,
		cfg.Credentials.AccessTokenExpiry,
		cfg.Credentials.RefreshTokenExpiry,
	)

	apiKeyManager := auth.NewAPIKeyManager()

	// Create auth middlewares
	authMiddleware := middlewares.NewAuthMiddleware(jwtProvider, apiKeyManager, logger)

	// ===============================
	// ✅ Create Global routes
	// Register all route groups
	// ===============================

	

	// Register user routes (signup, signin, password reset)
	RegisterUserRoutes(
		v1, pool, logger, cfg,
		jwtProvider,
		emailService,
		authMiddleware,
	)

	// Routes requiring authentication
	protected := v1.Group("")
	protected.Use(authMiddleware.JWTMiddleware())


	// // Register budgeting routes (items and transactions)
	// RegisterBudgetingRoutes(
	// 	protected, pool, logger, cfg,
	// 	authMiddleware,
	// )
}

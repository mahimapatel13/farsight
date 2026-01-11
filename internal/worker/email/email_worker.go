package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"budget-planner/internal/domain/integration"
	"budget-planner/pkg/email/emailtypes"
	"budget-planner/pkg/email/queue"
	"budget-planner/pkg/logger"
)

// EmailWorker processes queued email tasks asynchronously
type EmailWorker struct {
	manager     *integration.EmailManager
	emailQueue  queue.EmailQueue
	retryPolicy queue.RetryPolicy
	maxRetries  int
	logger      *logger.Logger
}

// NewEmailWorker creates a new EmailWorker
func NewEmailWorker(manager *integration.EmailManager, emailQueue queue.EmailQueue, retryPolicy queue.RetryPolicy, maxRetries int, log *logger.Logger) *EmailWorker {
	return &EmailWorker{
		manager:     manager,
		emailQueue:  emailQueue,
		retryPolicy: retryPolicy,
		maxRetries:  maxRetries,
		logger:      log,
	}
}

// StartWorker starts the email task processing loop with multiple workers
func (w *EmailWorker) StartWorker(ctx context.Context, workerCount int) {
	w.logger.Info("Email worker started, waiting for tasks...")

	// Launch multiple workers to process the queue concurrently
	for i := 0; i < workerCount; i++ {
		go w.processQueue(ctx)
	}
}

// processQueue continuously processes email tasks based on priority
func (w *EmailWorker) processQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Warn("Email worker stopped due to context cancellation")
			return
		default:
			// ✅ Process tasks with priority using the updated queue
			err := w.emailQueue.ProcessQueue(ctx)
			if err != nil {
				w.logger.Error("Error processing email queue", "error", err)
				time.Sleep(2 * time.Second) // Sleep before retrying
			}
		}
	}
}

// processEmailTask processes an individual email task
func (w *EmailWorker) processEmailTask(ctx context.Context, task *emailtypes.EmailTask) {
	w.logger.Info("Processing email task",
		"task_id", task.TaskID,
		"recipients", task.Email.To,
		"subject", task.Email.Subject,
		"priority", task.Priority,
	)

	// ✅ Send email using EmailManager with template and data
	messageID, err := w.manager.Send(ctx, *task.Email)
	if err != nil {
		w.logger.Error("Error sending email, retrying...",
			"task_id", task.TaskID,
			"error", err,
			"attempts", task.RetryCount+1,
		)

		// Retry if allowed, else mark as failed
		if task.ShouldRetry() {
			w.logger.Info("Retrying email task",
				"task_id", task.TaskID,
				"attempts", task.RetryCount+1,
			)
			w.retryFailedTask(ctx, task)
		} else {
			w.logger.Error("Email failed after max retries, notifying admin...",
				"task_id", task.TaskID,
				"max_retries", w.maxRetries,
			)
			task.MarkAsFailed()
			w.handleFailedTask(ctx, task)
		}
	} else {
		w.logger.Info("Email sent successfully",
			"task_id", task.TaskID,
			"recipients", task.Email.To,
			"message_id", messageID,
		)
		task.MarkAsSent()
	}
}

// retryFailedTask re-enqueues the failed task with a backoff delay
func (w *EmailWorker) retryFailedTask(ctx context.Context, task *emailtypes.EmailTask) {
	delay := w.retryPolicy.GetRetryInterval(task.RetryCount)

	go func() {
		time.Sleep(delay)
		w.logger.Info("Re-enqueuing task for retry after delay",
			"task_id", task.TaskID,
			"delay", delay.String(),
		)

		// Increment the retry count before re-enqueuing
		task.RetryCount++
		err := w.emailQueue.Enqueue(ctx, task)
		if err != nil {
			w.logger.Error("Failed to re-enqueue email task for retry",
				"task_id", task.TaskID,
				"error", err,
			)
		}
	}()
}

// handleFailedTask handles email failures after max retries
func (w *EmailWorker) handleFailedTask(ctx context.Context, task *emailtypes.EmailTask) {
	w.logger.Error("Final failure for email task after max retries",
		"task_id", task.TaskID,
		"recipients", task.Email.To,
		"subject", task.Email.Subject,
	)

	// Additional failure handling logic (e.g., notify admin, store to DB, etc.)
	w.notifyAdminOnFailure(ctx, task)
}

// notifyAdminOnFailure notifies the admin about the final failure of an email task
func (w *EmailWorker) notifyAdminOnFailure(ctx context.Context, task *emailtypes.EmailTask) {
	// TODO: Update
	// 🎯 Admin email and subject setup
	adminEmail := "admin@tnprgpv.com"
	subject := "🚨 Email Task Failure Alert"

	// ✅ Prepare HTML body dynamically using fmt.Sprintf
	htmlBody := fmt.Sprintf(
		`
		<h2 style="color: red;">Email Task Failure Alert</h2>
		<p><strong>Task ID:</strong> %s</p>
		<p><strong>Recipients:</strong> %s</p>
		<p><strong>Subject:</strong> %s</p>
		<p><strong>Provider:</strong> %s</p>
		<p><strong>Priority:</strong> %d</p>
		<p><strong>Error:</strong> Max retries exceeded</p>
		`,
		task.TaskID,
		formatRecipients(task.Email.To), // Format multiple recipients safely
		task.Email.Subject,
		task.ProviderName,
		task.Priority,
	)

	// ✅ Prepare metadata for admin alert
	metadata := map[string]string{
		"type":      "admin_failure_alert",
		"task_id":   task.TaskID,
		"provider":  task.ProviderName,
		"priority":  fmt.Sprintf("%d", task.Priority),
		"max_tries": fmt.Sprintf("%d", task.MaxRetries),
	}

	// ✅ Create Email object for admin notification
	email := emailtypes.Email{
		ID:          fmt.Sprintf("admin-alert-%s", task.TaskID), // Unique ID for tracking
		To:          []string{adminEmail},                       // Admin email as recipient
		From:        "no-reply@tnprgpv.com",                     // Default sender
		Subject:     subject,                                    // Subject
		Body:        htmlBody,                                   // HTML content
		Attachments: nil,                                        // No attachments
		Metadata:    metadata,                                   // Metadata for tracking
		SentAt:      time.Now(),                                 // Current timestamp
	}

	// ✅ Send the email using EmailManager.Send
	messageID, err := w.manager.Send(ctx, email)
	if err != nil {
		w.logger.Error("Failed to notify admin about email task failure",
			"task_id", task.TaskID,
			"error", err,
		)
	} else {
		w.logger.Info("Admin notified about email task failure successfully",
			"task_id", task.TaskID,
			"recipients", adminEmail,
			"message_id", messageID,
		)
	}
}

// formatRecipients safely formats email recipients for logging/interpolation
func formatRecipients(recipients []string) string {
	if len(recipients) == 0 {
		return "No Recipients"
	}
	return strings.Join(recipients, ", ")
}

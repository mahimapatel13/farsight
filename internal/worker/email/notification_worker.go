package worker

// import (
// 	"context"
// 	"encoding/json"
// 	"time"

// 	"server/internal/config"
// 	"server/internal/infrastructure/email"
// 	"server/pkg/logger"
// )

// // EmailTask represents an email to be sent by the worker
// type EmailTask struct {
// 	TemplateName string
// 	TemplateData map[string]interface{}
// 	Recipients   []string
// 	Subject      string
// 	Priority     int // Higher number = higher priority
// 	RequestedAt  time.Time
// }

// // Queue interface for retrieving email tasks
// type Queue interface {
// 	FetchTasks(ctx context.Context, batchSize int) ([][]byte, error)
// 	MarkTaskComplete(ctx context.Context, taskID string) error
// 	MarkTaskFailed(ctx context.Context, taskID string, err error) error
// }

// // EmailProvider interface
// type EmailProvider interface {
// 	SendEmail(task EmailTask) error
// }

// // NotificationWorker processes email sending tasks from a queue
// type NotificationWorker struct {
// 	emailProvider   EmailProvider
// 	templateEngine  *email.TemplateEngine
// 	queue           Queue
// 	logger          *logger.Logger
// 	workerID        string
// 	shutdownSignal  chan struct{}
// 	batchSize       int
// 	processingDelay time.Duration
// 	emailConfig     config.EmailConfig // Add EmailConfig
// }

// // NewNotificationWorker creates a new email notification worker
// func NewNotificationWorker(
// 	emailProvider EmailProvider,
// 	templateEngine *email.TemplateEngine,
// 	queue Queue,
// 	logger *logger.Logger,
// 	workerID string,
// 	batchSize int,
// 	processingDelay time.Duration,
// 	emailConfig config.EmailConfig, // Add EmailConfig to constructor
// ) *NotificationWorker {
// 	return &NotificationWorker{
// 		emailProvider:   emailProvider,
// 		templateEngine:  templateEngine,
// 		queue:           queue,
// 		logger:          logger,
// 		workerID:        workerID,
// 		shutdownSignal:  make(chan struct{}),
// 		batchSize:       batchSize,
// 		processingDelay: processingDelay,
// 		emailConfig:     emailConfig,
// 	}
// }

// // Start begins processing email tasks
// func (w *NotificationWorker) Start(ctx context.Context) {
// 	w.logger.Info("Starting email notification worker", "workerID", w.workerID)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			w.logger.Info("Context canceled, shutting down worker", "workerID", w.workerID)
// 			return
// 		case <-w.shutdownSignal:
// 			w.logger.Info("Shutdown signal received, stopping worker", "workerID", w.workerID)
// 			return
// 		default:
// 			w.processBatch(ctx)
// 			time.Sleep(w.processingDelay)
// 		}
// 	}
// }

// // Shutdown signals the worker to stop processing
// func (w *NotificationWorker) Shutdown() {
// 	close(w.shutdownSignal)
// }

// // processBatch processes a batch of email tasks
// func (w *NotificationWorker) processBatch(ctx context.Context) {
// 	// Fetch tasks from queue
// 	taskData, err := w.queue.FetchTasks(ctx, w.batchSize)
// 	if err != nil {
// 		w.logger.Error("Failed to fetch email tasks", "error", err)
// 		return
// 	}

// 	if len(taskData) == 0 {
// 		return // No tasks to process
// 	}

// 	for _, data := range taskData {
// 		// Process each task
// 		var task EmailTask
// 		if err := json.Unmarshal(data, &task); err != nil {
// 			w.logger.Error("Failed to unmarshal email task", "error", err)
// 			continue
// 		}

// 		// Process the task
// 		if err := w.processTask(ctx, task); err != nil {
// 			w.logger.Error("Failed to process email task",
// 				"error", err,
// 				"recipients", task.Recipients)
// 			// Mark task as failed in queue
// 			// Implementation to mark as failed
// 			continue
// 		}

// 		// Mark task as complete in queue
// 		// Implementation to mark as complete
// 	}
// }

// // processTask processes a single email task
// func (w *NotificationWorker) processTask(ctx context.Context, task EmailTask) error {
// 	// Render the email template
// 	_, err := w.templateEngine.RenderTemplate(task.TemplateName, task.TemplateData)
// 	if err != nil {
// 		return err
// 	}

// 	// Send the email
// 	return w.emailProvider.SendEmail(task)
// }

// // func (w *NotificationWorker) processTask(ctx context.Context, task EmailTask) error {
// // 	// Render the email template
// // 	emailBody, err := w.templateEngine.RenderTemplate(task.TemplateName, task.TemplateData)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	// Create email content
// // 	content := email.EmailContent{
// // 		Subject:  task.Subject,
// // 		HTMLBody: emailBody,
// // 		To:       task.Recipients,
// // 	}

// // 	// Send the email
// // 	return w.emailProvider.SendEmail(task)
// // }

// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"time"
// )

// // NotificationJob represents a notification job
// type NotificationJob struct {
// 	To           []string    `json:"to"`
// 	Subject      string      `json:"subject"`
// 	TemplateName string      `json:"template_name"`
// 	TemplateData interface{} `json:"template_data"`
// 	RetryCount   int         `json:"retry_count"`
// }

// // NotificationWorker processes email notification jobs
// type NotificationWorker struct {
// 	emailUseCase *email.EmailUseCase
// 	queue        QueueService
// }

// // QueueService defines the interface for a queue service
// type QueueService interface {
// 	Consume(context.Context, func([]byte) error)
// 	Publish(context.Context, []byte) error
// }

// // NewNotificationWorker creates a new notification worker
// func NewNotificationWorker(emailUseCase *notification.EmailUseCase, queue QueueService) *NotificationWorker {
// 	return &NotificationWorker{
// 		emailUseCase: emailUseCase,
// 		queue:        queue,
// 	}
// }

// // Start starts the notification worker
// func (w *NotificationWorker) Start(ctx context.Context) {
// 	w.queue.Consume(ctx, func(payload []byte) error {
// 		var job NotificationJob
// 		if err := json.Unmarshal(payload, &job); err != nil {
// 			log.Printf("Error unmarshalling job: %v", err)
// 			return err
// 		}

// 		// Process job
// 		_, err := w.emailUseCase.SendNotification(ctx, job.To, job.Subject, job.TemplateName, job.TemplateData)
// 		if err != nil {
// 			log.Printf("Error sending notification: %v", err)

// 			// Handle retries
// 			if job.RetryCount < 3 {
// 				job.RetryCount++

// 				// Retry after delay
// 				time.Sleep(time.Second * time.Duration(job.RetryCount*5))

// 				jobBytes, _ := json.Marshal(job)
// 				return w.queue.Publish(ctx, jobBytes)
// 			}

// 			log.Printf("Max retries reached for job, dropping: %+v", job)
// 			return err
// 		}

// 		return nil
// 	})
// }

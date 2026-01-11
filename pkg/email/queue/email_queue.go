package queue

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"budget-planner/pkg/email/emailtypes"
	"budget-planner/pkg/logger"

	"github.com/google/uuid"
)

// EmailQueue defines an interface for enqueuing and processing email tasks
type EmailQueue interface {
	// Enqueue adds an email task to the queue
	Enqueue(ctx context.Context, task *emailtypes.EmailTask) error

	// ProcessQueue processes email tasks from the queue
	ProcessQueue(ctx context.Context) error

	// RetryFailedTasks retries tasks that previously failed
	RetryFailedTasks(ctx context.Context) error

	// SetEmailService dynamically assigns the email provider
	SetEmailService(provider emailtypes.EmailProvider)
}

// DefaultEmailQueue implements EmailQueue using a queueing mechanism
type DefaultEmailQueue struct {
	mutex        sync.Mutex
	taskQueue    TaskPriorityQueue
	retryPolicy  *RetryPolicy
	emailService emailtypes.EmailProvider
	logger       *logger.Logger
}

// NewEmailQueue initializes a new priority-based email queue
func NewEmailQueue(emailService emailtypes.EmailProvider, retryPolicy *RetryPolicy, log *logger.Logger) *DefaultEmailQueue {
	pq := make(TaskPriorityQueue, 0)
	heap.Init(&pq)

	return &DefaultEmailQueue{
		taskQueue:    pq,
		retryPolicy:  retryPolicy,
		emailService: emailService,
		logger:       log,
	}
}

// Enqueue adds a new email task to the priority queue
func (q *DefaultEmailQueue) Enqueue(ctx context.Context, task *emailtypes.EmailTask) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	task.TaskID = uuid.NewString()
	task.CreatedAt = time.Now()

	heap.Push(&q.taskQueue, task)

	q.logger.Info("Enqueued email task with priority",
		"task_id", task.TaskID,
		"recipients", task.Email.To,
		"priority", task.Priority,
	)
	return nil
}

// ProcessQueue processes email tasks from the priority queue
func (q *DefaultEmailQueue) ProcessQueue(ctx context.Context) error {
	for {
		q.mutex.Lock()
		if len(q.taskQueue) == 0 {
			q.mutex.Unlock()
			time.Sleep(1 * time.Second) // Wait if the queue is empty
			continue
		}

		task := heap.Pop(&q.taskQueue).(*emailtypes.EmailTask)
		q.mutex.Unlock()

		if task.IsCompleted() {
			q.logger.Info("Skipping completed task",
				"task_id", task.TaskID,
				"status", task.Status,
			)
			continue
		}

		q.logger.Info("Processing email task with priority",
			"task_id", task.TaskID,
			"priority", task.Priority,
			"recipients", task.Email.To,
		)

		if err := q.processTask(ctx, task); err != nil {
			q.logger.Error("Failed to process email task",
				"task_id", task.TaskID,
				"error", err,
			)

			if task.ShouldRetry() {
				task.IncrementRetry()
				q.retryFailedTask(ctx, task)
			} else {
				task.MarkAsFailed()
			}
		}
	}
}

// processTask sends an email and handles the result
func (q *DefaultEmailQueue) processTask(ctx context.Context, task *emailtypes.EmailTask) error {
	resp, err := q.emailService.Send(ctx, task.Email)
	if err != nil {
		q.logger.Error("Email sending failed",
			"task_id", task.TaskID,
			"recipients", task.Email.To,
			"error", err,
		)
		task.MarkAsFailed() // ❗ Mark task as failed
		return err
	}

	task.MarkAsSent() // ✅ Mark task as sent
	q.logger.Info("Email sent successfully",
		"task_id", task.TaskID,
		"recipients", task.Email.To,
		"message_id", resp.MessageID,
	)
	return nil
}

// RetryFailedTasks retries tasks that failed earlier based on retry policy
func (q *DefaultEmailQueue) RetryFailedTasks(ctx context.Context) error {
	failedTasks, err := q.retryPolicy.GetFailedTasks(ctx)
	if err != nil {
		q.logger.Error("Failed to fetch failed email tasks for retry", "error", err)
		return err
	}

	for _, task := range failedTasks {
		// ❗ Skip completed tasks
		if task.IsCompleted() {
			q.logger.Warn("Skipping already completed task",
				"task_id", task.TaskID,
				"status", task.Status,
			)
			continue
		}

		if task.ShouldRetry() {
			q.logger.Info("Retrying failed email task",
				"task_id", task.TaskID,
				"attempts", task.RetryCount,
			)
			task.IncrementRetry()
			q.retryFailedTask(ctx, task)
		} else {
			q.logger.Warn("Max retries reached, marking task as failed",
				"task_id", task.TaskID,
			)
			task.MarkAsFailed()
		}
	}
	return nil
}

// retryFailedTask re-enqueues the failed task with exponential backoff delay
func (q *DefaultEmailQueue) retryFailedTask(ctx context.Context, task *emailtypes.EmailTask) {
	go func() {
		if task.ShouldRetry() {
			q.logger.Info("Re-enqueuing task for retry after exponential backoff",
				"task_id", task.TaskID,
				"retry_count", task.RetryCount,
			)
			if err := q.Enqueue(ctx, task); err != nil {
				q.logger.Error("Failed to re-enqueue email task for retry",
					"task_id", task.TaskID,
					"error", err,
				)
			}
		} else {
			q.logger.Warn("Max retries reached, marking task as failed",
				"task_id", task.TaskID,
			)
			task.MarkAsFailed()
		}
	}()
}

// SetEmailService dynamically assigns the email provider after initialization
func (q *DefaultEmailQueue) SetEmailService(provider emailtypes.EmailProvider) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.emailService = provider
	q.logger.Info("Email service provider assigned to EmailQueue",
		"provider", provider.Name(),
	)
}

// TaskPriorityQueue implements heap.Interface for priority queue
type TaskPriorityQueue []*emailtypes.EmailTask

func (pq TaskPriorityQueue) Len() int { return len(pq) }

func (pq TaskPriorityQueue) Less(i, j int) bool {
	// Lower priority number = higher priority (1 is highest, 5 is lowest)
	return pq[i].Priority < pq[j].Priority
}

func (pq TaskPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *TaskPriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*emailtypes.EmailTask))
}

func (pq *TaskPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

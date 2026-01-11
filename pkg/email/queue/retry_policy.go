package queue

import (
	"context"
	"errors"
	"time"

	"budget-planner/pkg/email/emailtypes"
	"budget-planner/pkg/logger"
)

// RetryPolicy defines policies for retrying failed email tasks
type RetryPolicy struct {
	MaxRetries      int                              // Maximum retry attempts for a task
	RetryIntervals  []time.Duration                  // Retry intervals between attempts
	FailedTaskStore map[string]*emailtypes.EmailTask // Store for failed tasks
	logger          *logger.Logger                   // Structured logger instance
}

// DefaultRetryIntervals defines fallback retry intervals if none are provided
var DefaultRetryIntervals = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	10 * time.Minute,
}

// NewRetryPolicy creates a new retry policy with specified settings
func NewRetryPolicy(maxRetries int, retryIntervals []time.Duration, log *logger.Logger) *RetryPolicy {
	// 🎯 Use default intervals if no retry intervals are provided
	if len(retryIntervals) == 0 {
		log.Warn("No retry intervals provided, falling back to default intervals")
		retryIntervals = DefaultRetryIntervals
	}

	return &RetryPolicy{
		MaxRetries:      maxRetries,
		RetryIntervals:  retryIntervals,
		FailedTaskStore: make(map[string]*emailtypes.EmailTask),
		logger:          log,
	}
}

// GetRetryInterval returns the retry interval based on the retry count
func (r *RetryPolicy) GetRetryInterval(retryCount int) time.Duration {
	if retryCount >= len(r.RetryIntervals) {
		r.logger.Warn("Retry count exceeded defined intervals, using the longest interval",
			"retry_count", retryCount,
			"max_interval", r.RetryIntervals[len(r.RetryIntervals)-1],
		)
		return r.RetryIntervals[len(r.RetryIntervals)-1]
	}
	r.logger.Debug("Returning retry interval",
		"retry_count", retryCount,
		"interval", r.RetryIntervals[retryCount],
	)
	return r.RetryIntervals[retryCount]
}

// SaveFailedTask stores a failed task for future retries
func (r *RetryPolicy) SaveFailedTask(ctx context.Context, task *emailtypes.EmailTask) error {
	if task.RetryCount >= r.MaxRetries {
		r.logger.Warn("Max retries reached, discarding task",
			"task_id", task.TaskID,
			"retry_count", task.RetryCount,
		)
		return errors.New("max retry attempts reached")
	}

	r.FailedTaskStore[task.TaskID] = task
	r.logger.Info("Saved failed email task for retry",
		"task_id", task.TaskID,
		"retry_count", task.RetryCount,
	)
	return nil
}

// GetFailedTasks retrieves all failed tasks eligible for retry
func (r *RetryPolicy) GetFailedTasks(ctx context.Context) ([]*emailtypes.EmailTask, error) {
	var tasks []*emailtypes.EmailTask
	for _, task := range r.FailedTaskStore {
		if task.ShouldRetry() {
			tasks = append(tasks, task)
		}
	}

	r.logger.Debug("Fetched failed tasks for retry",
		"eligible_task_count", len(tasks),
	)
	return tasks, nil
}

// RemoveTask removes a task from the failed task store after successful processing
func (r *RetryPolicy) RemoveTask(taskID string) {
	if _, exists := r.FailedTaskStore[taskID]; exists {
		delete(r.FailedTaskStore, taskID)
		r.logger.Info("Removed task from failed task store",
			"task_id", taskID,
		)
	} else {
		r.logger.Warn("Attempted to remove non-existent task from failed task store",
			"task_id", taskID,
		)
	}
}

// ClearFailedTasks clears all failed tasks (useful for cleanup)
func (r *RetryPolicy) ClearFailedTasks() {
	r.FailedTaskStore = make(map[string]*emailtypes.EmailTask)
	r.logger.Info("Cleared all failed email tasks from retry store")
}

// HasFailedTask checks if a task with the given ID exists in the store
func (r *RetryPolicy) HasFailedTask(taskID string) bool {
	_, exists := r.FailedTaskStore[taskID]
	if exists {
		r.logger.Debug("Task found in failed task store",
			"task_id", taskID,
		)
	} else {
		r.logger.Debug("Task not found in failed task store",
			"task_id", taskID,
		)
	}
	return exists
}

// GetTaskByID retrieves a failed task by its ID
func (r *RetryPolicy) GetTaskByID(taskID string) (*emailtypes.EmailTask, error) {
	task, exists := r.FailedTaskStore[taskID]
	if !exists {
		r.logger.Warn("Task not found in failed task store",
			"task_id", taskID,
		)
		return nil, errors.New("task not found")
	}
	r.logger.Debug("Retrieved task from failed task store",
		"task_id", taskID,
	)
	return task, nil
}


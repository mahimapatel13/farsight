package emailtypes

import (
	"errors"
	"fmt"
	"time"
)

type EmailTask struct {
	TaskID       string    `json:"task_id"`                // Unique identifier for the task
	Email        *Email    `json:"email,omitempty"`        // Embedded Email struct
	ProviderName string    `json:"provider_name"`          // Email provider to use (e.g., "smtp", "sendgrid")
	RetryCount   int       `json:"retry_count"`            // Number of retry attempts made
	MaxRetries   int       `json:"max_retries"`            // Maximum allowed retry attempts
	RequestedAt  time.Time `json:"requested_at,omitempty"` // Timestamp when the task was requested
	CreatedAt    time.Time `json:"created_at,omitempty"`   // Timestamp when the task was created
	Status       string    `json:"status"`                 // Task status (queued, sending, sent, failed, retrying)
	Priority     int       `json:"priority"`               // 📌 Higher the number, lower the priority, Default priority 1.
}

// Validate validates the task and associated email
func (t *EmailTask) Validate() error {
	if t.Email == nil {
		return errors.New("email is missing in the task")
	}
	if err := t.Email.Validate(); err != nil {
		return fmt.Errorf("email validation failed: %w", err)
	}
	if t.ProviderName == "" {
		return errors.New("email provider is required")
	}
	return nil
}

// PrepareTask initializes task defaults and prepares the email
func (t *EmailTask) PrepareTask() {
	if t.TaskID == "" {
		t.TaskID = generateUniqueID()
	}
	if t.Email != nil {
		t.Email.PrepareForSend() // Sets SentAt for email
	}
	t.CreatedAt = time.Now()
	t.Status = EmailStatusQueued
}

// ShouldRetry checks if the task can be retried with backoff delay
func (t *EmailTask) ShouldRetry() bool {
	if t.RetryCount >= t.MaxRetries {
		return false
	}

	// Exponential backoff for retries
	backoffDuration := time.Duration(2<<t.RetryCount) * time.Second
	time.Sleep(backoffDuration)
	return true
}

// MarkAsFailed updates task status to "failed" and prevents further retries
func (t *EmailTask) MarkAsFailed() {
	t.Status = EmailStatusFailed
	t.RetryCount = t.MaxRetries
}

// MarkAsSent updates task status to "sent" and marks task as complete
func (t *EmailTask) MarkAsSent() {
	t.Status = EmailStatusSent
}

// SetStatus updates the task status
func (t *EmailTask) SetStatus(status string) {
	t.Status = status
}

// IsCompleted checks if the task has completed (sent or failed)
func (t *EmailTask) IsCompleted() bool {
	return t.Status == EmailStatusSent || t.Status == EmailStatusFailed
}

// IsValidProvider checks if the provider name is valid
func (t *EmailTask) IsValidProvider(validProviders []string) bool {
	for _, p := range validProviders {
		if t.ProviderName == p {
			return true
		}
	}
	return false
}

// generateUniqueID generates a random unique identifier
func generateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// IncrementRetry increments the retry count and updates task status if max retries are exceeded
func (t *EmailTask) IncrementRetry() {
	t.RetryCount++
	if t.RetryCount >= t.MaxRetries {
		t.MarkAsFailed()
	}
}


package emailtypes

import (
	"budget-planner/internal/common/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Email defines a standard structure for sending emails
type Email struct {
	ID          string            `json:"id"`                    // Unique ID for tracking
	To          []string          `json:"to"`                    // Recipient email addresses
	CC          []string          `json:"cc,omitempty"`          // CC email addresses
	BCC         []string          `json:"bcc,omitempty"`         // BCC email addresses
	From        string            `json:"from"`                  // Sender's email address
	Subject     string            `json:"subject"`               // Email subject line
	Body        string            `json:"body"`                  // Email content (HTML or plain text)
	Attachments []Attachment      `json:"attachments,omitempty"` // List of email attachments
	Metadata    map[string]string `json:"metadata,omitempty"`    // Additional metadata for tracking
	SentAt      time.Time         `json:"sent_at,omitempty"`     // Timestamp when the email was sent
}

// Attachment defines the structure for email attachments
type Attachment struct {
	Filename    string `json:"filename"`     // Name of the attachment file
	ContentType string `json:"content_type"` // MIME type of the attachment (e.g., "application/pdf")
	Content     []byte `json:"content"`      // Binary content of the file
}

// EmailResponse contains the result of an email send operation
type EmailResponse struct {
	MessageID string    `json:"message_id"` // Unique ID of the sent email
	Status    string    `json:"status"`     // Delivery status ("queued", "sent", "failed")
	SentAt    time.Time `json:"sent_at"`    // Timestamp when the email was sent
}

// ErrInvalidEmail is returned when an email is invalid
var ErrInvalidEmail = errors.New("invalid email address")

// Validate checks the basic validity of the email
func (e *Email) Validate() error {
	// Check recipients
	if len(e.To) == 0 && len(e.CC) == 0 && len(e.BCC) == 0 {
		return errors.New("no recipients specified in To, Cc, or Bcc")
	}

	// Validate sender
	if e.From == "" || !isValidEmail(e.From) {
		fmt.Printf("Warning: Invalid sender email address, using default email from config. Provided: %s\n", e.From)
	}

	// Validate recipients
	for _, recipient := range append(e.To, append(e.CC, e.BCC...)...) {
		if !isValidEmail(recipient) {
			return fmt.Errorf("invalid recipient email: %s", recipient)
		}
	}

	// Check subject and body
	if strings.TrimSpace(e.Subject) == "" {
		return errors.New("email subject is required")
	}
	if strings.TrimSpace(e.Body) == "" {
		return errors.New("email body is empty")
	}

	// ✅ Validate attachments
	for _, attachment := range e.Attachments {
		if attachment.Filename == "" {
			return errors.New("attachment filename is missing")
		}
		if attachment.ContentType == "" {
			return errors.New("attachment content type is missing")
		}
		if len(attachment.Content) == 0 {
			return fmt.Errorf("attachment content is empty: %s", attachment.Filename)
		}

		if !validateAttachmentType(attachment.ContentType) {
			return fmt.Errorf("attachment %s has an unsupported content type: %s", attachment.Filename, attachment.ContentType)
		}
	}

	return nil
}

// isValidEmail checks the basic validity of an email address
func isValidEmail(email string) bool {
	// Basic regex for email validation
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// JoinRecipients returns a comma-separated string of recipients
func (e *Email) JoinRecipients() string {
	recipients := append(e.To, append(e.CC, e.BCC...)...)
	return strings.Join(recipients, ", ")
}

var allowedContentTypes = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"text/plain":      true,
}

// validateAttachmentType checks whether the attachment content type is allowed
func validateAttachmentType(contentType string) bool {
	return allowedContentTypes[contentType]
}

// PrepareForSend ensures email has a valid timestamp
func (e *Email) PrepareForSend() {
	if e.SentAt.IsZero() {
		e.SentAt = time.Now()
	}
}

const (
	EmailStatusQueued = "queued"
	EmailStatusSent   = "sent"
	EmailStatusFailed = "failed"
	EmailStatusRetry  = "retry"
)

// IsValidStatus checks if the provided status is valid
func IsValidStatus(status string) bool {
	switch status {
	case EmailStatusQueued, EmailStatusSent, EmailStatusFailed, EmailStatusRetry:
		return true
	default:
		return false
	}
}

// CleanAndSanitizeMetadata cleans and sanitizes metadata for Email
func (e *Email) CleanAndSanitizeMetadata() {
	e.Metadata = utils.CleanMetadata(e.Metadata)
}

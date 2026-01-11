package emailtypes

import (
	"context"
)

// EmailProvider defines the interface for sending emails using various providers.
// Implementations should support sending both plain and templated emails.
type EmailProvider interface {
	// Send sends a plain email and returns an EmailResponse or an error.
	// `email` contains the recipient, subject, and body details.
	Send(ctx context.Context, email *Email) (*EmailResponse, error)

	// BatchSend sends multiple emails in one go, optimizing for high-volume operations.
	// `emails` contains a list of emails to be sent.
	BatchSend(ctx context.Context, emails []*Email) ([]*EmailResponse, error)

	// HealthCheck verifies the availability of the email provider.
	// Returns an error if the provider is unreachable or misconfigured.
	HealthCheck(ctx context.Context) error

	// Name returns the name of the provider (e.g., "smtp", "sendgrid")
	Name() string
}


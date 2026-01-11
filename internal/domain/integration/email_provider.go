package integration

import (
	errr "budget-planner/internal/common/errors"
	"budget-planner/internal/config"
	"budget-planner/pkg/email/emailtypes"
	"budget-planner/pkg/email/queue"
	"budget-planner/pkg/logger"
	"context"
	"errors"
	"fmt"
	"sync"
)

// EmailManager dynamically manages all email providers
type EmailManager struct {
	MaxRetries      int                                 // Max number of retry attempts
	providers       map[string]emailtypes.EmailProvider // Map of email providers
	defaultProvider emailtypes.EmailProvider            // Default email provider
	mutex           sync.Mutex                          // Mutex for provider access
	logger          *logger.Logger                      // Structured logger
	emailQueue      queue.EmailQueue                    // Email queue for async tasks
}

// NewEmailManager initializes and configures EmailManager with available providers
func NewEmailManager(
	config config.EmailConfig,
	emailQueue queue.EmailQueue,
	log *logger.Logger,
) (*EmailManager, error) {

	manager := &EmailManager{
		MaxRetries: config.MaxRetries,
		providers:  make(map[string]emailtypes.EmailProvider),
		logger:     log,
		emailQueue: emailQueue,
	}

	log.Info("EmailManager configuration loaded", "config", fmt.Sprintf("%+v", config))

	// ✅ Dynamically load available providers
	manager.loadProviders(config)

	// ✅ Set the default provider if configured
	if provider, ok := manager.providers[config.Provider]; ok && config.Enabled {
		manager.defaultProvider = provider
		log.Info("Default email provider configured", "provider", config.Provider)
	} else {
		log.Warn("Configured default provider not found, falling back to SMTP")

		// ✅ Fallback to SMTP if enabled
		if smtpProvider, ok := manager.providers["smtp"]; ok && config.Enabled {
			manager.defaultProvider = smtpProvider
			log.Info("Fallback to SMTP provider", "host", config.SMTP.Host)
		} else {
			return nil, errr.NewForbiddenError("no valid email provider configured or email sending is disabled")
		}
	}

	return manager, nil
}

// loadProviders dynamically configures providers based on the given config
func (m *EmailManager) loadProviders(config config.EmailConfig) {
	m.logger.Info("Loading email providers...")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Add SMTP provider if configured and enabled
	if config.SMTP.Host != "" && config.Enabled {
		smtpProvider := emailtypes.NewSMTPProvider(config.SMTP, m.logger)
		m.providers["smtp"] = smtpProvider
		m.logger.Info("SMTP provider configured", "host", config.SMTP.Host, "sender_email", config.SenderEmail)
	}
}

// SetDefaultProvider changes the default provider dynamically if it's not already the current default
func (m *EmailManager) SetDefaultProvider(providerName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	provider, exists := m.providers[providerName]
	if !exists {
		err := fmt.Errorf("failed to set default provider: provider '%s' not found", providerName)
		m.logger.Error("Provider not found", "provider_name", providerName, "error", err)
		return err
	}

	// ✅ Check if the provider is already the default
	if m.defaultProvider == provider {
		m.logger.Warn("Attempted to reset the same default provider", "provider_name", providerName)
		return fmt.Errorf("provider '%s' is already the default provider", providerName)
	}

	// ✅ Set as default if different
	m.defaultProvider = provider
	m.logger.Info("Default provider set successfully", "provider_name", providerName)
	return nil
}

// Send sends a plain email using the default provider
func (m *EmailManager) Send(ctx context.Context, email emailtypes.Email) (string, error) {
	if m.defaultProvider == nil {
		m.logger.Error("Email send failed: no default provider configured")
		return "", errors.New("email send failed: no default provider configured")
	}

	messageResponse, err := m.defaultProvider.Send(ctx, &email)
	if err != nil {
		m.logger.Error("Error sending email", "error", err, "to", email.To, "CC", email.CC, "BCC", email.BCC, "subject", email.Subject)
		return "", err
	}

	m.logger.Info("Email sent successfully", "message_id", messageResponse.MessageID, "to", email.To, "CC", email.CC, "BCC", email.BCC, "subject", email.Subject)
	return messageResponse.MessageID, nil
}

// QueueEmail adds an email to the queue for async sending with optional priority and maxRetries
func (m *EmailManager) QueueEmail(ctx context.Context, email emailtypes.Email, optionalParams ...int) error {
	// 🚨 Check if the email queue is initialized
	if m.emailQueue == nil {
		m.logger.Error("Email queue is not initialized")
		return errors.New("email queue not initialized")
	}

	// ✅ Validate email before enqueuing
	if err := email.Validate(); err != nil {
		m.logger.Error("Invalid email detected", "error", err, "to", email.To)
		return fmt.Errorf("email validation failed: %w", err)
	}

	// 🎯 Extract optional parameters: priority and maxRetries
	priority := 2              // Default priority
	maxRetries := m.MaxRetries // Default max retries

	// Assign optional parameters if provided
	if len(optionalParams) > 0 && optionalParams[0] > 0 {
		priority = optionalParams[0]
	}
	if len(optionalParams) > 1 && optionalParams[1] > 0 && optionalParams[1] <= m.MaxRetries {
		maxRetries = optionalParams[1]
	}

	// 🎯 Prepare the email task with valid priority and retries
	task := &emailtypes.EmailTask{
		Email:        &email,
		ProviderName: m.defaultProvider.Name(), // Dynamically set the default provider
		MaxRetries:   maxRetries,               // Set retry limit with a valid value
		Priority:     priority,                 // Set priority
	}
	task.PrepareTask() // Properly initialize CreatedAt, TaskID, and default status

	// 🚀 Enqueue the prepared email task
	err := m.emailQueue.Enqueue(ctx, task)
	if err != nil {
		m.logger.Error("Failed to enqueue email", "error", err, "to", email.To)
		return fmt.Errorf("failed to enqueue email: %w", err)
	}

	m.logger.Info("Email added to queue successfully",
		"to", email.To,
		"subject", email.Subject,
		"task_id", task.TaskID,
		"priority", task.Priority,
		"max_retries", task.MaxRetries,
	)
	return nil
}

// HealthCheck validates the availability of all configured providers
func (m *EmailManager) HealthCheck(ctx context.Context) error {
	for name, provider := range m.providers {
		if err := provider.HealthCheck(ctx); err != nil {
			m.logger.Error("Health check failed for provider", "provider", name, "error", err)
			return err
		}
		m.logger.Info("Health check passed for provider", "provider", name)
	}
	return nil
}

// GetDefaultProvider returns the default email provider
func (m *EmailManager) GetDefaultProvider() emailtypes.EmailProvider {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.defaultProvider
}

// SetEmailQueue sets the email queue for the manager
func (m *EmailManager) SetEmailQueue(emailQueue queue.EmailQueue) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.emailQueue = emailQueue
	m.logger.Info("Email queue set for EmailManager")
}

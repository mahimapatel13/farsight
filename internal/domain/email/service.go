package email

import (
	errors "budget-planner/internal/common/errors"
	"budget-planner/internal/domain/integration"
	"budget-planner/pkg/email/emailtypes"
	"budget-planner/pkg/logger"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// EmailService defines the email service interface
type EmailService interface {
	// Email Operations
	SendVerificationEmail(ctx context.Context, username, email, password string) *errors.DomainError
	SendPasswordResetEmail(ctx context.Context, email, resetToken string) *errors.DomainError
	SendAccountUnlockedEmail(ctx context.Context, email string) *errors.DomainError
	SendForcedPasswordChangeEmail(ctx context.Context, email, newPassword string) *errors.DomainError
	SendCertificateMail(ctx context.Context, certificateRequest CertificateEmail) *errors.DomainError
}

// emailService uses EmailManager to manage email providers and templates
type emailService struct {
	manager *integration.EmailManager // Email provider manager
	repo    TemplateRepository        // Template repository for DB operations
	logger  *logger.Logger            // Structured logger for logging events
}

// NewEmailService creates a new email service with dependencies
func NewEmailService(
	manager *integration.EmailManager,
	repo TemplateRepository,
	log *logger.Logger,
) EmailService {
	return &emailService{
		manager: manager,
		repo:    repo,
		logger:  log,
	}
}

// NewEmail creates a new Email instance with required fields
func NewEmail(
	to, cc, bcc []string,
	subject, body string,
	attachments []emailtypes.Attachment,
	metadata map[string]string,
) *emailtypes.Email {
	return &emailtypes.Email{
		To:          to,
		CC:          cc,
		BCC:         bcc,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
		Metadata:    metadata,
		SentAt:      time.Time{}, // SentAt is set when the email is actually sent
	}
}

// interpolateTemplate safely interpolates placeholders in the template body with HTML support
func interpolateTemplate(templateBody string, data map[string]string) (string, *errors.DomainError) {
	tmpl, err := template.New("email").Parse(templateBody)
	if err != nil {
		return "", errors.NewBusinessError("ERROR_PARSING_TEMPLATE", "error parsing email template", nil)
	}

	var renderedBody strings.Builder
	if err := tmpl.Execute(&renderedBody, data); err != nil {
		return "", errors.NewBusinessError("ERROR_RENDERING_TEMPLATE", "error rendering email template", nil)
	}

	return renderedBody.String(), nil
}

// SendVerificationEmail sends an account verification email
func (s *emailService) SendVerificationEmail(ctx context.Context, username, email, password string) *errors.DomainError {
	// ✅ Validate input to prevent invalid or empty values
	if email == "" || password == "" {
		s.logger.Error("invalid input: email or password is empty")
		return errors.NewBadInputError("email and password are required for verification email", nil)
	}

	// ✅ Fetch verification email template from DB
	template, err := s.repo.GetTemplateByName(ctx, "verification_email")
	if err != nil {
		s.logger.Error("failed to fetch template", "template_name", "verification_email", "error", err)
		return errors.NewDatabaseError("failed to load email template", err)
	}

	// ✅ Prepare template data for interpolation
	data := map[string]string{
		"UserName": username, // Placeholder, can be replaced with actual user name if available
		"Password": password,
		"email":    email, // Optional for template, useful in some cases
	}

	// ✅ Interpolate template and prepare email body
	body, errr := interpolateTemplate(template.Body, data)
	if errr != nil {
		s.logger.Error("failed to interpolate verification template", "error", errr)
		return errors.NewBusinessError("template rendering error", "ERROR_RENDERING_TEMPLATE", nil)
	}

	// ✅ Prepare email using NewEmail
	emailObj := NewEmail(
		[]string{email},  // To
		nil,              // CC (optional)
		nil,              // BCC (optional)
		template.Subject, // Subject from template
		body,             // Rendered HTML body
		nil,              // Attachments (optional)
		map[string]string{"type": "verification"}, // Metadata
	)

	// ✅ Queue email for asynchronous sending
	if err := s.manager.QueueEmail(ctx, *emailObj); err != nil {
		s.logger.Error("failed to enqueue verification email", "to", email, "error", err)
		return errors.NewDatabaseError("failed to enqueue verification email", err)
	}

	s.logger.Info("Verification email added to queue successfully", "to", email)
	return nil
}

// SendPasswordResetEmail sends a password reset email with a secure reset token
func (s *emailService) SendPasswordResetEmail(ctx context.Context, email, resetToken string) *errors.DomainError {
	// ✅ Validate input to prevent nil or empty values
	if email == "" || resetToken == "" {
		s.logger.Error("invalid input: email or resetToken is empty")
		return errors.NewBadInputError("email and resetToken are required for password reset email", nil)
	}

	// ✅ Fetch the reset password template from DB
	template, err := s.repo.GetTemplateByName(ctx, "reset_template")
	if err != nil {
		s.logger.Error("failed to fetch template", "template_name", "reset_template", "error", err)
		return errors.NewDatabaseError("failed to load password reset email template", err)
	}

	// ✅ Prepare template data for interpolation
	data := map[string]string{
		"token": resetToken,
		"email": email,
	}

	// ✅ Interpolate the reset template with provided data
	body, errr := interpolateTemplate(template.Body, data)
	if errr != nil {
		s.logger.Error("failed to interpolate password reset template", "error", errr)
		return errors.NewBusinessError("template rendering error", "ERROR_RENDERING_TEMPLATE", nil)
	}

	// ✅ Prepare the email object using NewEmail
	emailObj := NewEmail(
		[]string{email},                    // To
		nil,                                // CC (optional)
		nil,                                // BCC (optional)
		template.Subject,                   // Subject from template
		body,                               // Rendered HTML body
		nil,                                // Attachments (optional)
		map[string]string{"type": "reset"}, // Metadata for audit
	)

	// ✅ Queue the email for async sending
	if err := s.manager.QueueEmail(ctx, *emailObj); err != nil {
		s.logger.Error("failed to enqueue password reset email", "to", email, "error", err)
		return errors.NewBusinessError("failed to enqueue password reset email", "ERROR_ENQUEUEING_EMAIL", nil)
	}

	s.logger.Info("Password reset email added to queue successfully", "to", email)
	return nil
}

// SendAccountUnlockedEmail sends an account unlock notification email
func (s *emailService) SendAccountUnlockedEmail(ctx context.Context, email string) *errors.DomainError {
	// ✅ Validate input to prevent sending to an empty email
	if email == "" {
		s.logger.Error("invalid input: email is empty")
		return errors.NewBadInputError("email is required for account unlock notification", nil)
	}

	// ✅ Fetch the account unlock notification template from DB
	template, err := s.repo.GetTemplateByName(ctx, "account_unlocked_template")
	if err != nil {
		s.logger.Error("failed to fetch template", "template_name", "account_unlocked_template", "error", err)
		return errors.NewDatabaseError("failed to load account unlocked email template", err)
	}

	// ✅ Prepare the email body (no dynamic data in this case)
	body, errr := interpolateTemplate(template.Body, map[string]string{})
	if errr != nil {
		s.logger.Error("failed to interpolate account unlock template", "error", errr)
		return errors.NewBusinessError("template rendering error", "ERROR_RENDERING_TEMPLATE", nil)
	}

	// ✅ Prepare the email object using NewEmail
	emailObj := NewEmail(
		[]string{email},                       // To
		nil,                                   // CC (optional)
		nil,                                   // BCC (optional)
		template.Subject,                      // Subject from template
		body,                                  // Rendered HTML body
		nil,                                   // Attachments (optional)
		map[string]string{"type": "unlocked"}, // Metadata for audit
	)

	// ✅ Queue the email for async sending
	if err := s.manager.QueueEmail(ctx, *emailObj); err != nil {
		s.logger.Error("failed to enqueue account unlock email", "to", email, "error", err)
		return errors.NewBusinessError("failed to enqueue account unlock email", "ERROR_ENQUEUEING_EMAIL", nil)
	}

	s.logger.Info("Account unlock email added to queue successfully", "to", email)
	return nil
}

// SendForcedPasswordChangeEmail sends a forced password change notification email
func (s *emailService) SendForcedPasswordChangeEmail(ctx context.Context, email, newPassword string) *errors.DomainError {
	// ✅ Validate input to prevent sending to an empty email
	if email == "" || newPassword == "" {
		s.logger.Error("invalid input: email or newPassword is empty")
		return errors.NewBadInputError("email and newPassword are required for forced password change email", nil)
	}

	// ✅ Fetch the forced password change template from DB
	template, err := s.repo.GetTemplateByName(ctx, "forced_password_change_template")
	if err != nil {
		s.logger.Error("failed to fetch template", "template_name", "forced_password_change_template", "error", err)
		return errors.NewDatabaseError("failed to load forced password change email template", err)
	}

	// ✅ Prepare template data for interpolation
	data := map[string]string{
		"NewPassword": newPassword,
		"email":       email,
	}

	// ✅ Interpolate the template with provided data
	body, errr := interpolateTemplate(template.Body, data)
	if errr != nil {
		s.logger.Error("failed to interpolate forced password change template", "error", errr)
		return errors.NewBusinessError("template rendering error", "ERROR_RENDERING_TEMPLATE", nil)
	}

	// ✅ Prepare the email object using NewEmail
	emailObj := NewEmail(
		[]string{email},  // To
		nil,              // CC (optional)
		nil,              // BCC (optional)
		template.Subject, // Subject from template
		body,             // Rendered HTML body
		nil,              // Attachments (optional)
		map[string]string{"type": "forced_password"}, // Metadata for audit
	)

	// ✅ Queue the email for async sending
	if err := s.manager.QueueEmail(ctx, *emailObj); err != nil {
		s.logger.Error("failed to enqueue forced password change email", "to", email, "error", err)
		return errors.NewBusinessError("failed to enqueue forced password change email", "ERROR_ENQUEUEING_EMAIL", nil)
	}

	s.logger.Info("Forced password change email added to queue successfully", "to", email)
	return nil
}

// SendCertificateMail sends a certificate email with attachment
func (s *emailService) SendCertificateMail(ctx context.Context, req CertificateEmail) *errors.DomainError {
	if req.EventTitle == "" || req.Recipient.Email == "" || req.Recipient.Name == "" || req.Certificate == nil {
		s.logger.Error("invalid input: eventTitle, recipient email or certificate content is empty")
		return errors.NewValidationError("invalid input", map[string]any{
			"eventTitle":         req.EventTitle,
			"recipientEmail":     req.Recipient.Email,
			"recipientName":      req.Recipient.Name,
			"certificateContent": req.Certificate,
		})
	}

	template, err := s.repo.GetTemplateByName(ctx, "Certificate Email")
	if err != nil {
		s.logger.Error("failed to fetch template", "template_name", "certificate_email", "error", err)
		return errors.NewDatabaseError("failed to fetch email template", err)
	}

	subject, errr := interpolateTemplate(template.Subject, map[string]string{
		"eventTitle": req.EventTitle,
	})
	if errr != nil {
		s.logger.Error("failed to interpolate template subject", "recipient", req.Recipient.Email, "error", errr)
		return errors.NewBusinessError("ERROR_RENDERING_TEMPLATE", "template subject rendering error", nil)
	}

	body, errr := interpolateTemplate(template.Body, map[string]string{
		"eventTitle": req.EventTitle,
		"UserName":   req.Recipient.Name,
		"toEmail":    req.Recipient.Email,
		"certURL":    "",
	})
	if errr != nil {
		s.logger.Error("failed to interpolate template", "recipient", req.Recipient.Email, "error", errr)
		return errors.NewBusinessError("ERROR_RENDERING_TEMPLATE", "template rendering error", nil)
	}

	// Create the email object
	emailObj := NewEmail(
		[]string{req.Recipient.Email}, // To
		nil,                           // CC (optional)
		nil,                           // BCC (optional)
		subject,                       // Subject
		body,                          // Body as HTML
		[]emailtypes.Attachment{
			{
				Filename:    fmt.Sprintf("%s_certificate.pdf", req.Recipient.Name),
				ContentType: "application/pdf",
				Content:     req.Certificate, // Base64 encoded content
			},
		}, // Attachments (optional)
		map[string]string{"type": "certificate"}, // Metadata
	)

	// Queue the email for asynchronous sending
	if err := s.manager.QueueEmail(ctx, *emailObj); err != nil {
		s.logger.Error("failed to enqueue email", "recipient", req.Recipient.Email, "error", err)
		return errors.NewBusinessError("ERROR_SENDING_EMAIL", "failed to enqueue certificate email", nil)
	}

	s.logger.Info("Certificate email queued successfully", "recipient", req.Recipient.Email)
	return nil
}

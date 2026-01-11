package email

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// EmailTemplate defines a template structure
type EmailTemplate struct {
	ID        uuid.UUID
	Name      string
	Subject   string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CertificateEmail struct {
	Recipient RecipientInfo
	EventTitle string // Name of the event for context
	Certificate []byte
}
type RecipientInfo struct {
	Name  string
	Email string
}

// ===========================
// ✅ Utility Methods
// ===========================

// ToDomain maps CreateEmailTemplateRequest to EmailTemplate domain model
func (req *CreateEmailTemplateRequest) ToDomain() *EmailTemplate {
	return &EmailTemplate{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(req.Name),
		Subject:   strings.TrimSpace(req.Subject),
		Body:      strings.TrimSpace(req.Body),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ToDomain maps UpdateEmailTemplateRequest to EmailTemplate with updated fields
func (req *UpdateEmailTemplateRequest) ToDomain(existing *EmailTemplate) *EmailTemplate {
	return &EmailTemplate{
		ID:        req.TemplateID,
		Name:      strings.TrimSpace(req.Name),
		Subject:   strings.TrimSpace(req.Subject),
		Body:      strings.TrimSpace(req.Body),
		CreatedAt: existing.CreatedAt, // Retain original created_at
		UpdatedAt: time.Now(),
	}
}

// PrepareForUpdate updates the `updated_at` timestamp before modifying
func (et *EmailTemplate) PrepareForUpdate() {
	et.UpdatedAt = time.Now()
}

// Validate checks that all required fields are valid for the template
func (et *EmailTemplate) Validate() error {
	if strings.TrimSpace(et.Name) == "" {
		return errors.New("template name is required")
	}
	if strings.TrimSpace(et.Subject) == "" {
		return errors.New("template subject is required")
	}
	if strings.TrimSpace(et.Body) == "" {
		return errors.New("template body cannot be empty")
	}
	if len(et.Name) > 255 {
		return errors.New("template name exceeds maximum length of 255 characters")
	}
	if len(et.Subject) > 255 {
		return errors.New("template subject exceeds maximum length of 255 characters")
	}
	return nil
}

// Matches checks if two templates have the same Name and Subject
func (et *EmailTemplate) Matches(other *EmailTemplate) bool {
	return strings.EqualFold(et.Name, other.Name) && strings.EqualFold(et.Subject, other.Subject)
}

// IsValidTemplate checks if required fields are non-empty
func (et *EmailTemplate) IsValidTemplate() bool {
	return et.Name != "" && et.Subject != "" && et.Body != ""
}

// ============================
// 📥 DTOs for Email Template Operations
// ============================

// CreateEmailTemplateRequest DTO for creating a new template
type CreateEmailTemplateRequest struct {
	Name    string `json:"name" validate:"required,max=100"`    // Template name (unique)
	Subject string `json:"subject" validate:"required,max=255"` // Email subject
	Body    string `json:"body" validate:"required"`            // HTML/Plain text body
}

// Validate validates the CreateEmailTemplateRequest fields
func (req *CreateEmailTemplateRequest) Validate() error {
	v := validator.New()
	return v.Struct(req)
}

// UpdateEmailTemplateRequest DTO for updating an existing template
type UpdateEmailTemplateRequest struct {
	TemplateID uuid.UUID `json:"template_id" validate:"required"`   // UUID of the template
	Name       string    `json:"name" validate:"omitempty,max=100"` // Optional: Name to update
	Subject    string    `json:"subject" validate:"omitempty,max=255"`
	Body       string    `json:"body" validate:"omitempty"`
}

// Validate validates the UpdateEmailTemplateRequest fields
func (req *UpdateEmailTemplateRequest) Validate() error {
	v := validator.New()
	return v.Struct(req)
}

// DeleteEmailTemplateRequest DTO for deleting a template by ID
type DeleteEmailTemplateRequest struct {
	TemplateID uuid.UUID `json:"template_id" validate:"required"`
}

// Validate validates the DeleteEmailTemplateRequest
func (req *DeleteEmailTemplateRequest) Validate() error {
	v := validator.New()
	return v.Struct(req)
}

// GetEmailTemplateByNameRequest DTO for retrieving a template by name
type GetEmailTemplateByNameRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// Validate validates the GetEmailTemplateByNameRequest
func (req *GetEmailTemplateByNameRequest) Validate() error {
	v := validator.New()
	return v.Struct(req)
}

// ListEmailTemplatesRequest DTO for listing templates with optional filters
type ListEmailTemplatesRequest struct {
	Name   string `json:"name" validate:"omitempty,max=100"`       // Optional filter by name
	Limit  int    `json:"limit" validate:"omitempty,gt=0,lte=100"` // Pagination limit (max 100)
	Offset int    `json:"offset" validate:"omitempty,gte=0"`       // Offset for pagination
}

// Validate validates the ListEmailTemplatesRequest
func (req *ListEmailTemplatesRequest) Validate() error {
	v := validator.New()
	return v.Struct(req)
}


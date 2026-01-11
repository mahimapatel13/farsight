package email

import (
	"context"

	"github.com/google/uuid"

	"budget-planner/internal/common/errors"
)

// TemplateRepository defines the interface for email template operations
type TemplateRepository interface {
	GetTemplateByName(ctx context.Context, name string) (*EmailTemplate, *errors.InfrastructureError)
	CreateTemplate(ctx context.Context, template *EmailTemplate) *errors.InfrastructureError
	UpdateTemplate(ctx context.Context, template *EmailTemplate) *errors.InfrastructureError
	DeleteTemplate(ctx context.Context, id uuid.UUID) *errors.InfrastructureError
	ListTemplates(ctx context.Context) ([]*EmailTemplate, *errors.InfrastructureError)
}


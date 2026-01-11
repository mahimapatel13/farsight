package repositories

import (
	"context"
	"time"

	"budget-planner/internal/common/errors"
	"budget-planner/internal/domain/email"
	"budget-planner/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTemplateRepository implements TemplateRepository for PostgreSQL
type PostgresTemplateRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgresTemplateRepository initializes a new repository
func NewPostgresTemplateRepository(pool *pgxpool.Pool, logger *logger.Logger) email.TemplateRepository {
	return &PostgresTemplateRepository{
		pool:   pool,
		logger: logger,
	}
}

// GetTemplateByName fetches a template by name
func (r *PostgresTemplateRepository) GetTemplateByName(ctx context.Context, name string) (*email.EmailTemplate, *errors.InfrastructureError) {

	// ✅ Apply a timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const query = `
	SELECT id, name, subject, body_html, created_at, updated_at
	FROM email_schema.email_templates
	WHERE name = $1
	`

	template := &email.EmailTemplate{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Subject,
		&template.Body,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	// ✅ Handle "no rows found" scenario
	if err == pgx.ErrNoRows {
		r.logger.Warn("Template not found", "name", name)
		return nil, errors.NewInfraNotFoundError("email_template", map[string]any{"name": name})
	}

	// ✅ Handle database-related errors with custom infra errors
	if err != nil {
		pgErr := errors.GetInfraPgError(err)
		if pgErr != nil {
			// Handle unique constraint violation if applicable
			if errors.IsUniqueConstraintViolation(err) {
				return nil, errors.NewInfraConflictError("email_template", errors.GetInfraPgErrorDetails(err))
			}
		}
		r.logger.Error("Error fetching template by name", "error", err, "name", name)
		return nil, errors.NewInfraDatabaseError("fetching email template", err)
	}

	r.logger.Info("Template fetched successfully", "name", name, "template_id", template.ID)
	return template, nil
}

// CreateTemplate inserts a new template into the database
func (r *PostgresTemplateRepository) CreateTemplate(ctx context.Context, template *email.EmailTemplate) *errors.InfrastructureError {
	const query = `
	INSERT INTO email_schema.email_templates (id, name, subject, body_html, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	`
	template.ID = uuid.New()
	_, err := r.pool.Exec(ctx, query,
		template.ID,
		template.Name,
		template.Subject,
		template.Body,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		r.logger.Error("Error creating new email template", "error", err, "template_name", template.Name)
		return  errors.NewInfraDatabaseError("creating new email template",err)
	}
	return nil
}

// UpdateTemplate updates an existing template in the database
func (r *PostgresTemplateRepository) UpdateTemplate(ctx context.Context, template *email.EmailTemplate) *errors.InfrastructureError {

	// ✅ Apply a timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const query = `
	UPDATE email_schema.email_templates
	SET subject = $1, body_html = $2, updated_at = $3
	WHERE name = $4
	`

	// ✅ Execute the update query
	res, err := r.pool.Exec(ctx, query,
		template.Subject,
		template.Body,
		time.Now(),
		template.Name,
	)

	// ✅ Handle database error
	if err != nil {
		r.logger.Error("Error updating email template", "error", err, "template_name", template.Name)
		pgErr := errors.GetInfraPgError(err)
		if pgErr != nil {
			// Handle unique constraint violations if applicable
			if errors.IsUniqueConstraintViolation(err) {
				return errors.NewInfraConflictError("email_template", errors.GetInfraPgErrorDetails(err))
			}
		}
		return errors.NewInfraDatabaseError("updating email template", err)
	}

	// ✅ Check if the template was found and updated
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("Template not found for update", "template_name", template.Name)
		return errors.NewInfraNotFoundError("email_template", map[string]any{"name": template.Name})
	}

	// ✅ Log success and return
	r.logger.Info("Email template updated successfully", "template_name", template.Name, "rows_affected", rowsAffected)
	return nil
}

// DeleteTemplate removes a template by ID
func (r *PostgresTemplateRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) *errors.InfrastructureError {
	// ✅ Apply a timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const query = `DELETE FROM email_schema.email_templates WHERE id = $1`

	// ✅ Execute the delete query
	res, err := r.pool.Exec(ctx, query, id)

	// ✅ Handle database error
	if err != nil {
		r.logger.Error("Error deleting email template", "error", err, "template_id", id)
		pgErr := errors.GetInfraPgError(err)
		if pgErr != nil {
			// Handle potential foreign key or constraint errors
			if errors.IsForeignKeyViolation(err) {
				return errors.NewInfraConflictError("email_template_deletion", errors.GetInfraPgErrorDetails(err))
			}
		}
		return errors.NewInfraDatabaseError("deleting email template", err)
	}

	// ✅ Check if any row was deleted
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("Template not found for deletion", "template_id", id)
		return errors.NewInfraNotFoundError("email_template", map[string]any{"id": id})
	}

	// ✅ Log success and return
	r.logger.Info("Email template deleted successfully", "template_id", id, "rows_affected", rowsAffected)
	return nil
}

// ListTemplates retrieves all email templates
func (r *PostgresTemplateRepository) ListTemplates(ctx context.Context) ([]*email.EmailTemplate, *errors.InfrastructureError) {
    const query = `
	SELECT id, name, subject, body_html, created_at, updated_at
	FROM email_schema.email_templates
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		r.logger.Error("Error listing email templates", "error", err)
		return nil, errors.NewInfraDatabaseError("listing email templates", err)
	}
	defer rows.Close()

	var templates []*email.EmailTemplate
	for rows.Next() {
		template := &email.EmailTemplate{}
		if err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Subject,
			&template.Body,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			r.logger.Error("Error scanning email template", "error", err)
			return nil, errors.NewInfraDatabaseError("scanning email template",err)
		}
		templates = append(templates, template)
	}
	return templates, nil
}


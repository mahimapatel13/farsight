package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// Infrastructure-specific error types
const (
	// Database errors
	InfraDatabaseError    ErrorType = "INFRA_DATABASE"
	InfraTransactionError ErrorType = "INFRA_TRANSACTION"
	InfraBatchError       ErrorType = "INFRA_BATCH"
	InfraNotFoundError    ErrorType = "INFRA_NOT_FOUND"

	// Connection errors
	InfraConnectionError ErrorType = "INFRA_CONNECTION"

	// Migration/Schema errors
	InfraMigrationError ErrorType = "INFRA_MIGRATION"

	// Data consistency errors
	InfraBadInputError        ErrorType = "INFRA_BAD_INPUT"
	InfraDataConsistencyError ErrorType = "INFRA_DATA_INCONSISTENCY"

	// Unknown/General errors
	InfraUnknownError ErrorType = "INFRA_UNKNOWN"

	InfraNetworkError     ErrorType = "INFRA_NETWORK"
	InfraIntegrationError ErrorType = "INFRA_INTEGRATION"
	InfraConflictError    ErrorType = "INFRA_CONFLICT"
)

const (
	PgErrUniqueViolation     = "23505"
	PgErrForeignKeyViolation = "23503"
	PgErrSerializationFail   = "40001"
)

// InfrastructureError is used for errors occurring in the infrastructure layer
type InfrastructureError struct {
	Type    ErrorType
	Message string
	Code    string
	Details map[string]any
	Cause   error
}

// Implement the error interface
func (e *InfrastructureError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (code: %s, cause: %v)", e.Type, e.Message, e.Code, e.Cause)
	}
	return fmt.Sprintf("%s: %s (code: %s)", e.Type, e.Message, e.Code)
}

// Unwrap returns the underlying cause of the error
func (e *InfrastructureError) Unwrap() error {
	return e.Cause
}

func WrapInfraError(base error, msg string, typ ErrorType, code string, details map[string]any) *InfrastructureError {
	return &InfrastructureError{
		Type:    typ,
		Message: msg,
		Code:    code,
		Details: details,
		Cause:   base,
	}
}

// ===============================
// 🛠️ Generic Error Constructors
// ===============================

// NewInfraError creates a generic infrastructure error
func NewInfraError(message string, errorType ErrorType, code string, details map[string]any, cause error) *InfrastructureError {
	return &InfrastructureError{
		Type:    errorType,
		Message: message,
		Code:    code,
		Details: details,
		Cause:   cause,
	}
}

// ===============================
// 🗂️ Specific Error Constructors
// ===============================

// NewInfraConflictError returns a conflict error indicating a violation of a unique constraint
func NewInfraConflictError(entity string, details map[string]any) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Conflict with existing %s", entity),
		InfraConflictError,
		"INFRA_CONFLICT",
		details,
		nil,
	)
}

// NewInfraDatabaseError returns an error for database operations
func NewInfraDatabaseError(operation string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Database error during %s", operation),
		InfraDatabaseError,
		"INFRA_DB_ERROR",
		map[string]any{"operation": operation},
		cause,
	)
}

// NewInfraTransactionError returns an error for transaction failures
func NewInfraTransactionError(action string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Transaction error during %s", action),
		InfraTransactionError,
		"INFRA_TX_ERROR",
		map[string]any{"action": action},
		cause,
	)
}

// NewInfraBatchError returns an error for batch operation failures
func NewInfraBatchError(action string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Batch processing error during %s", action),
		InfraBatchError,
		"INFRA_BATCH_ERROR",
		map[string]any{"action": action},
		cause,
	)
}

// NewInfraConnectionError returns an error for connection issues
func NewInfraConnectionError(service string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Connection error to %s", service),
		InfraConnectionError,
		"INFRA_CONN_ERROR",
		map[string]any{"service": service},
		cause,
	)
}

// NewInfraMigrationError returns an error for migration failures
func NewInfraMigrationError(migration string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Migration error during %s", migration),
		InfraMigrationError,
		"INFRA_MIGRATION_ERROR",
		map[string]any{"migration": migration},
		cause,
	)
}

// NewInfraDataConsistencyError returns an error for data consistency issues
func NewInfraDataConsistencyError(entity string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Data inconsistency detected for %s", entity),
		InfraDataConsistencyError,
		"INFRA_DATA_INCONSISTENCY",
		map[string]any{"entity": entity},
		cause,
	)
}

// NewInfraUnknownError returns a generic unknown infrastructure error
func NewInfraUnknownError(cause error) *InfrastructureError {
	return NewInfraError(
		"An unknown infrastructure error occurred",
		InfraUnknownError,
		"INFRA_UNKNOWN_ERROR",
		nil,
		cause,
	)
}

// NewInfraNotFoundError returns an error for not found errors
func NewInfraNotFoundError(entity string, details map[string]any) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("%s not found", entity),
		InfraNotFoundError,
		"INFRA_NOT_FOUND",
		details,
		nil,
	)
}

// NewInfraBadInputError returns an error for invalid or bad input
func NewInfraBadInputError(field string, details map[string]any) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Invalid input for %s", field),
		InfraBadInputError,
		"INFRA_BAD_INPUT",
		details,
		nil,
	)
}

// NewInfraNetworkError returns an error for network-related issues
func NewInfraNetworkError(service string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Network error in %s", service),
		InfraNetworkError,
		"INFRA_NETWORK_ERROR",
		map[string]any{"service": service},
		cause,
	)
}

// NewInfraIntegrationError returns an error for integration-related issues
func NewInfraIntegrationError(service string, cause error) *InfrastructureError {
	return NewInfraError(
		fmt.Sprintf("Integration error with %s", service),
		InfraIntegrationError,
		"INFRA_INTEGRATION_ERROR",
		map[string]any{"service": service},
		cause,
	)
}

// ===============================
// 🔎 Utility Functions
// ===============================

// IsInfrastructureError checks if an error is an infrastructure error
func IsInfrastructureError(err error) bool {
	var ie *InfrastructureError
	return errors.As(err, &ie)
}

// InfraErrorTypeOf extracts the error type if it's an infrastructure error
func InfraErrorTypeOf(err error) ErrorType {
	var ie *InfrastructureError
	if errors.As(err, &ie) {
		return ie.Type
	}
	return InfraUnknownError
}

// ===============================
// ✅ Specific Error Checks
// ===============================

// IsInfraDatabaseError checks if the error is a database error
func IsInfraDatabaseError(err error) bool {
	return InfraErrorTypeOf(err) == InfraDatabaseError
}

// IsInfraTransactionError checks if the error is a transaction error
func IsInfraTransactionError(err error) bool {
	return InfraErrorTypeOf(err) == InfraTransactionError
}

// IsInfraBatchError checks if the error is a batch operation error
func IsInfraBatchError(err error) bool {
	return InfraErrorTypeOf(err) == InfraBatchError
}

// IsInfraMigrationError checks if the error is a migration error
func IsInfraMigrationError(err error) bool {
	return InfraErrorTypeOf(err) == InfraMigrationError
}

// IsInfraConnectionError checks if the error is a connection error
func IsInfraConnectionError(err error) bool {
	return InfraErrorTypeOf(err) == InfraConnectionError
}

// IsInfraDataConsistencyError checks if the error is a data consistency error
func IsInfraDataConsistencyError(err error) bool {
	return InfraErrorTypeOf(err) == InfraDataConsistencyError
}

// IsInfraNotFoundError checks if the error is a not found error
func IsInfraNotFoundError(err error) bool {
	return InfraErrorTypeOf(err) == InfraNotFoundError
}

// IsInfraBadInputError checks if the error is a bad input error
func IsInfraBadInputError(err error) bool {
	return InfraErrorTypeOf(err) == InfraBadInputError
}

// IsInfraNetworkError checks if the error is a network error
func IsInfraNetworkError(err error) bool {
	return InfraErrorTypeOf(err) == InfraNetworkError
}

// IsInfraIntegrationError checks if the error is an integration error
func IsInfraIntegrationError(err error) bool {
	return InfraErrorTypeOf(err) == InfraIntegrationError
}

// IsInfraConflictError checks if the error is a conflict error
func IsInfraConflictError(err error) bool {
	return InfraErrorTypeOf(err) == InfraConflictError
}

// ===============================
// 🔍 PostgreSQL Error Utilities
// ===============================

// GetInfraPgError extracts *pgconn.PgError from an error, if available.
func GetInfraPgError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}

// GetInfraPgErrorDetails extracts error details from a pgconn.PgError
func GetInfraPgErrorDetails(err error) map[string]any {
	pgErr := GetInfraPgError(err)
	if pgErr == nil {
		return nil
	}
	return map[string]any{
		"code":       pgErr.Code,
		"constraint": pgErr.ConstraintName,
		"detail":     pgErr.Detail,
		"table":      pgErr.TableName,
		"message":    pgErr.Message,
	}
}

// IsUniqueConstraintViolation checks if the error is a unique constraint violation
func IsUniqueConstraintViolation(err error) bool {
	pgErr := GetInfraPgError(err)
	return pgErr != nil && pgErr.Code == "23505" // UNIQUE violation error code
}

// IsForeignKeyViolation checks if the error is a foreign key violation
func IsForeignKeyViolation(err error) bool {
	pgErr := GetInfraPgError(err)
	return pgErr != nil && pgErr.Code == "23503" // Foreign key violation error code
}

func AsInfraError(err error) (*InfrastructureError, bool) {
	var ie *InfrastructureError
	ok := errors.As(err, &ie)
	return ie, ok
}

func InfraToHTTPCode(err error) int {
	switch InfraErrorTypeOf(err) {
	case InfraNotFoundError:
		return 404
	case InfraBadInputError:
		return 400
	case InfraConflictError:
		return 409
	case InfraDatabaseError, InfraTransactionError:
		return 500
	default:
		return 500
	}
}

func IsInfraRetryable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || InfraErrorTypeOf(err) == InfraNetworkError {
		return true
	}
	// Extend with pg error code like 40001 (serialization failure)
	pgErr := GetInfraPgError(err)
	return pgErr != nil && pgErr.Code == "40001"
}

func (e *InfrastructureError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":    e.Type,
		"message": e.Message,
		"code":    e.Code,
		"details": e.Details,
	})
}


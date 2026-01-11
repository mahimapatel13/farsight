package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// APIError represents an error response from the API
type APIError struct {
	Status  int            `json:"-"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s - %s", e.Status, e.Code, e.Message)
}

// RespondWithError writes the error to the Gin context response
func (e *APIError) RespondWithError(c *gin.Context) {
	c.JSON(e.Status, e)
}

// NewAPIError creates a new API error
func NewAPIError(status int, code string, message string, details map[string]interface{}) *APIError {
	return &APIError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Predefined API error for conflicts
var ErrConflict = NewAPIError(http.StatusConflict, "conflict", "Conflict error", nil)

// Common error creators
func BadRequest(message string, details map[string]interface{}) *APIError {
	return NewAPIError(http.StatusBadRequest, "bad_request", message, details)
}

func Unauthorized(message string) *APIError {
	if message == "" {
		message = "Authentication required"
	}
	return NewAPIError(http.StatusUnauthorized, "unauthorized", message, nil)
}

func Forbidden(message string) *APIError {
	if message == "" {
		message = "You don't have permission to access this resource"
	}
	return NewAPIError(http.StatusForbidden, "forbidden", message, nil)
}

func NotFound(resource string) *APIError {
	message := "Resource not found"
	if resource != "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return NewAPIError(http.StatusNotFound, "not_found", message, nil)
}

func InternalServerError(err error) *APIError {
	return NewAPIError(http.StatusInternalServerError, "internal_server_error", "Internal server error", map[string]any{
		"error": err.Error(),
	})
}

func Conflict(message string, details map[string]any) *APIError {
	return NewAPIError(http.StatusConflict, "conflict", message, details)
}

// HandleValidationErrors converts validator errors into API errors
func HandleValidationErrors(err error) *APIError {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make(map[string]interface{})
		for _, fieldError := range validationErrors {
			details[fieldError.Field()] = map[string]interface{}{
				"tag":     fieldError.Tag(),
				"value":   fieldError.Value(),
				"message": getValidationErrorMessage(fieldError),
			}
		}
		return BadRequest("Validation failed", details)
	}
	return BadRequest(err.Error(), nil)
}

// getValidationErrorMessage returns a human-readable message for validation errors
func getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Should be at least %s", fieldError.Param())
	case "max":
		return fmt.Sprintf("Should be at most %s", fieldError.Param())
	case "url":
		return "Invalid URL format"
	default:
		return fmt.Sprintf("Failed validation on %s", fieldError.Tag())
	}
}

// ErrorHandler middlewares for uniform error handling in Gin
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the error and stack trace
				stackTrace := string(debug.Stack())
				fmt.Printf("PANIC: %v\nStack: %s\n", r, stackTrace)

				// Create an API error
				apiErr := NewAPIError(
					http.StatusInternalServerError,
					"internal_server_error",
					"An unexpected error occurred",
					nil, // Don't expose internal details in production
				)
				apiErr.RespondWithError(c)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// HandleDBError handles database-related errors
func HandleDBError(err error) *APIError {
	// Here you could have specific handling for pgx errors
	// For example check for unique constraint violations
	if err != nil {
		// This is where you'd check for specific pgx error types
		// Example: if pgerrcode.IsIntegrityConstraintViolation(err)

		// For now, a simple fallback
		return InternalServerError(err)
	}
	return nil
}

// IsNotFoundError checks if an error is a "not found" error
func IsNotFoundErrorAPI(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusNotFound
	}
	return false
}

// RespondWithError is a helper to respond with an error
func RespondWithError(c *gin.Context, err error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		apiErr.RespondWithError(c)
		return
	}

	// If it's not an APIError, create an internal server error
	InternalServerError(err).RespondWithError(c)
}

func DomainToAPIError(err error) *APIError {
	var de *DomainError
	if errors.As(err, &de) {
		switch de.Type {
		case NotFoundError:
			return NewAPIError(http.StatusNotFound, de.Code, de.Message, de.Details)
		case ValidationError, BadInputError:
			return NewAPIError(http.StatusBadRequest, de.Code, de.Message, de.Details)
		case UnauthorizedError:
			return NewAPIError(http.StatusUnauthorized, de.Code, de.Message, de.Details)
		case ForbiddenError:
			return NewAPIError(http.StatusForbidden, de.Code, de.Message, de.Details)
		case ConflictError:
			return NewAPIError(http.StatusConflict, de.Code, de.Message, de.Details)
		case RateLimitError:
			return NewAPIError(http.StatusTooManyRequests, de.Code, de.Message, de.Details)
		case TimeoutError:
			return NewAPIError(http.StatusGatewayTimeout, de.Code, de.Message, de.Details)
		default:
			return NewAPIError(http.StatusInternalServerError, de.Code, de.Message, de.Details)
		}
	}
	// fallback for non-domain errors
	return InternalServerError(err)
}


package rest_utils

import (
	"net/http"

	"budget-planner/internal/common/errors"

	"github.com/gin-gonic/gin"
)

// StandardResponse defines the structure for all API responses
type StandardResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Success sends a standard success response with data
func Success(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data any, message string) {
	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends a standard error response using your APIError
func Error(c *gin.Context, err error) {
	apiErr := errors.DomainToAPIError(err)
	apiErr.RespondWithError(c)
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, err error) {
	apiErr := errors.HandleValidationErrors(err)
	apiErr.RespondWithError(c)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data any, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}


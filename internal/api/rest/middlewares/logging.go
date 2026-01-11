package middlewares

import (
	"bytes"
	"io"
	"time"

	"budget-planner/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// responseBodyWriter is a custom gin.ResponseWriter that captures the response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// LoggingMiddleware logs the incoming HTTP request and response
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Generate request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		c.Header("X-Request-ID", requestID)

		// Set request ID in context
		c.Set("requestID", requestID)

		// Create a scoped logger with request ID
		reqLogger := log.WithField("request_id", requestID)

		// Log the request
		reqLogger.Info("Request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Read request body for logging if needed
		var requestBody []byte
		if c.Request.Body != nil && c.Request.Method != "GET" {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the body for further processing
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response body
		responseBodyWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseBodyWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get status
		statusCode := c.Writer.Status()

		// Log fields
		logFields := []any{
			"duration_ms", duration.Milliseconds(),
			"status_code", statusCode,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		}

		// Add error information if any
		if len(c.Errors) > 0 {
			logFields = append(logFields, "errors", c.Errors.String())
		}

		// Add request and response body details for errors
		if statusCode >= 400 {
			// Log request body for errors (with truncation for large bodies)
			if len(requestBody) > 0 {
				bodyToLog := string(requestBody)
				if len(bodyToLog) > 1024 {
					bodyToLog = bodyToLog[:1024] + "... [truncated]"
				}
				logFields = append(logFields, "request_body", bodyToLog)
			}

			// Log response body for errors (with truncation for large bodies)
			responseBody := responseBodyWriter.body.String()
			if len(responseBody) > 0 {
				if len(responseBody) > 1024 {
					responseBody = responseBody[:1024] + "... [truncated]"
				}
				logFields = append(logFields, "response_body", responseBody)
			}

			reqLogger.Error("Request failed", logFields...)
		} else {
			reqLogger.Info("Request completed", logFields...)
		}
	}
}

// RequestIDMiddleware ensures a request ID is available in the context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		c.Next()
	}
}

// GetRequestLogger gets a logger with the current request ID from context
func GetRequestLogger(c *gin.Context, log *logger.Logger) *logger.Logger {
	requestID, exists := c.Get("requestID")
	if !exists {
		return log
	}

	return log.WithField("request_id", requestID)
}


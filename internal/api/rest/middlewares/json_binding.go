package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New() // global validator instance

// BindJSONMiddleware binds JSON and validates input with error handling
func BindJSONMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var obj T

		// Bind JSON to the target object
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			c.Abort()
			return
		}

		// Always validate the struct
		if err := validate.Struct(obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			c.Abort()
			return
		}

		// Store the parsed object in context
		c.Set("requestBody", obj)
		c.Next()
	}
}

// GetRequestBody retrieves the parsed request body from context
func GetRequestBody[T any](c *gin.Context) (T, bool) {
	obj, exists := c.Get("requestBody")
	if !exists {
		var zero T
		return zero, false
	}
	// Assert to correct type automatically
	reqBody, ok := obj.(T)
	return reqBody, ok
}


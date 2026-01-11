package rest_utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetQueryInt retrieves an integer query parameter from the request, or returns the default if missing/invalid.
func GetQueryInt(c *gin.Context, key string, defaultValue int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}


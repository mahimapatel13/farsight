package rest_utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetPlatformProfileIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.UUID{}, false
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		return uuid.UUID{}, false
	}
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, false
	}
	return userUUID, true
}

func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	roleVal, exists := c.Get("userRole")
	if !exists {
		return "", false
	}
	roleStr, ok := roleVal.(string)
	if !ok {
		return "", false
	}
	return roleStr, true
}

// func GetUserContext(c *gin.Context) (uuid.UUID, string, bool) {
// 	userID, idExists := GetPlatformProfileIDFromContext(c)
// 	role, roleExists := GetUserRoleFromContext(c)

// 	if !idExists || !roleExists {
// 		return uuid.UUID{}, "", false
// 	}

// 	return userID, role, true
// }


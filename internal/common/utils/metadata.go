package utils

import (
	"regexp"
	"strings"
)

// CleanMetadata cleans and sanitizes metadata map
func CleanMetadata(metadata map[string]string) map[string]string {
	cleaned := make(map[string]string)
	allowedKeyRegex := regexp.MustCompile(`^[a-z0-9_\-]+$`)

	for key, value := range metadata {
		cleanedKey := strings.TrimSpace(strings.ToLower(key))
		if allowedKeyRegex.MatchString(cleanedKey) {
			cleaned[cleanedKey] = strings.TrimSpace(value)
		}
	}
	return cleaned
}

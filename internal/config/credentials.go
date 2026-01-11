package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// ServerCredentials contains all sensitive information used by the application
type ServerCredentials struct {
	APIKeys            map[string]string
	JWTAccessSecret    string
	JWTRefreshSecret   string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// loadCredentials initializes credentials from environment variables or .env file
func loadCredentials() (*ServerCredentials, error) {
	// Load .env file if it exists
	loadEnvFile()

	// Get credentials from environment variables
	return loadCredentialsFromEnv()
}

// loadEnvFile loads environment variables from .env file if present
func loadEnvFile() {
	// Determine which .env file to use based on environment
	env, err := loadEnvironment()
	if err != nil {
		return // If can't determine environment, continue with default .env
	}

	// Choose appropriate .env file
	envFileName := ".env"
	if env.Production {
		envFileName = ".env.production"
	} else if env.Testing {
		envFileName = ".env.testing"
	} else {
		envFileName = ".env.development"
	}

	// Fall back to .env if the specific file doesn't exist
	if _, err := os.Stat(envFileName); os.IsNotExist(err) {
		envFileName = ".env"
	}

	// Open the .env file
	file, err := os.Open(envFileName)
	if err != nil {
		// It's okay if the file doesn't exist, we'll just use existing env vars
		return
	}
	defer file.Close()

	// Parse the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Invalid format, skip this line
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if environment variable is not already set
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func parseDurationWithDays(durationStr string) (time.Duration, error) {
	if strings.HasSuffix(durationStr, "d") {
		daysStr := strings.TrimSuffix(durationStr, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, errors.New("invalid duration format for days")
		}
		// Convert days to hours (24 hours per day)
		return time.Duration(days*24) * time.Hour, nil
	}
	return time.ParseDuration(durationStr)
}

// loadCredentialsFromEnv loads credentials from environment variables
func loadCredentialsFromEnv() (*ServerCredentials, error) {

	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if jwtAccessSecret == "" {
		return nil, errors.New("JWT secret (JWT_ACCESS_SECRET) not set")
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		return nil, errors.New("JWT secret (JWT_REFRESH_SECRET) not set")
	}

	accessTokenExpiryStr := os.Getenv("JWT_ACCESS_TOKEN_EXPIRY")
	if accessTokenExpiryStr == "" {
		return nil, errors.New("access token expiry (JWT_ACCESS_TOKEN_EXPIRY) not set")
	}

	accessTokenExpiry, err := time.ParseDuration(accessTokenExpiryStr)
	if err != nil {
		return nil, errors.New("invalid access token expiry duration format")
	}

	refreshTokenExpiryStr := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY")
	if refreshTokenExpiryStr == "" {
		return nil, errors.New("refresh token expiry (JWT_REFRESH_TOKEN_EXPIRY) not set")
	}

	refreshTokenExpiry, err := parseDurationWithDays(refreshTokenExpiryStr)
	if err != nil {
		return nil, errors.New("invalid refresh token expiry duration format")
	}

	// Initialize API keys map
	apiKeys := make(map[string]string)

	// Add any defined API keys
	// Look for any environment variables with API_KEY_ prefix
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(key, "API_KEY_") {
			// Convert API_KEY_EXTERNAL_SERVICE to external_service
			serviceName := strings.ToLower(strings.TrimPrefix(key, "API_KEY_"))
			apiKeys[serviceName] = value
		}
	}

	return &ServerCredentials{
		// DatabasePassword:   dbPassword,
		APIKeys:            apiKeys,
		JWTAccessSecret:    jwtAccessSecret,
		JWTRefreshSecret:   jwtRefreshSecret,
		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
	}, nil
}

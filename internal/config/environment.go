// Only need to provide APP_ENV in the environment file
// Other configurations are loaded from the application code

package config

import (
	"fmt"
	"os"
)

// Environment contains environment-specific configuration settings
type Environment struct {
	Name       string
	Production bool
	Testing    bool
	Debug      bool
	LogLevel   string
}

// Valid environment names
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTesting     = "testing"
)

// loadEnvironment initializes and returns environment configuration
func loadEnvironment() (*Environment, error) {
	// Get environment name from ENV variable
	envName := os.Getenv("APP_ENV")
	if envName == "" {
		envName = EnvDevelopment // Default to development
	}

	// Validate environment name
	if !isValidEnv(envName) {
		return nil, fmt.Errorf("invalid environment name: %s", envName)
	}

	// Create environment configuration
	env := &Environment{
		Name:       envName,
		Production: envName == EnvProduction,
		Testing:    envName == EnvTesting,
		Debug:      getEnvAsBool("DEBUG", envName != EnvProduction),
		LogLevel:   getLogLevel(envName),
	}

	return env, nil
}

// isValidEnv checks if the environment name is valid
func isValidEnv(env string) bool {
	validEnvs := []string{EnvDevelopment, EnvStaging, EnvProduction, EnvTesting}
	for _, validEnv := range validEnvs {
		if env == validEnv {
			return true
		}
	}
	return false
}

// getLogLevel returns the appropriate log level based on the environment
func getLogLevel(env string) string {
	// Override with explicit setting if available
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		return level
	}

	// Default levels based on environment
	switch env {
	case EnvProduction:
		return "info"
	case EnvStaging:
		return "debug"
	case EnvTesting:
		return "debug"
	default:
		return "debug"
	}
}

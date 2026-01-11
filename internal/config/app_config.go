// Update use all the fields available in the struct
// Add support for the Google services and config
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the complete application configuration
type Config struct {
	Environment Environment
	Server      ServerConfig
	Database    DatabaseConfig
	CORS        CORSConfig
	Credentials ServerCredentials
	Integration IntegrationConfig
	Features    FeatureFlags
}

// ServerConfig contains all HTTP server related settings
type ServerConfig struct {
	Port                   string
	ReadTimeoutSeconds     int
	WriteTimeoutSeconds    int
	IdleTimeoutSeconds     int
	ShutdownTimeoutSeconds int
}

// DatabaseConfig contains all database connection settings
type DatabaseConfig struct {
	Host            string
	Port            string
	DatabaseName    string
	UserName        string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	/// Don't add the Database_URI field rather
	/// supply the required details as the individual variables
	/// the string will be auto generated back.
}

// CORSConfig contains Cross-Origin Resource Sharing settings
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// Load initializes and returns the application configuration
func Load() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		log.Fatal("Error loading .env file")
	}

	// Load environment-specific configuration
	env, err := loadEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	// Load credentials
	creds, err := loadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Load integration configuration
	integration, err := loadIntegrationConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load integration configuration: %w", err)
	}

	// Load feature flags
	features, err := loadFeatureFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to load feature flags: %w", err)
	}

	// Configure server
	serverConfig := ServerConfig{
		Port:                   getEnv("SERVER_PORT", "8080"),
		ReadTimeoutSeconds:     getEnvAsInt("SERVER_READ_TIMEOUT", 30),
		WriteTimeoutSeconds:    getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
		IdleTimeoutSeconds:     getEnvAsInt("SERVER_IDLE_TIMEOUT", 60),
		ShutdownTimeoutSeconds: getEnvAsInt("SERVER_SHUTDOWN_TIMEOUT", 30),
	}

	// Configure database
	dbConfig := DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		DatabaseName:    getEnv("DB_NAME", "tnp_rgpv"),
		UserName:        getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "tnp_rgpv_db_password"),
		SSLMode:         getEnv("DB_SSL_MODE", "require"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
	}

	// Configure CORS
	corsConfig := CORSConfig{
		AllowOrigins:     strings.Split(getEnv("CORS_ALLOW_ORIGINS", "*"), ","),
		AllowMethods:     strings.Split(getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"), ","),
		AllowHeaders:     strings.Split(getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"), ","),
		ExposeHeaders:    strings.Split(getEnv("CORS_EXPOSE_HEADERS", "Content-Length,Content-Type"), ","),
		AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
		MaxAge:           time.Duration(getEnvAsInt("CORS_MAX_AGE", 300)) * time.Second,
	}

	return &Config{
		Environment: *env,
		Server:      serverConfig,
		Database:    dbConfig,
		CORS:        corsConfig,
		Credentials: *creds,
		Integration: *integration,
		Features:    *features,
	}, nil
}

// Helper function to get environment variables with fallbacks
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Helper function to get environment variables as integers
func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// Helper function to get environment variables as booleans
func getEnvAsBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// Helper function to get environment variables as string slices
func getEnvAsSlice(key string, fallback []string, separator string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return strings.Split(value, separator)
	}
	return fallback
}

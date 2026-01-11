package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"budget-planner/internal/common/errors"
)

// IntegrationConfig contains configuration for external service integrations
type IntegrationConfig struct {
	Email       EmailConfig
	SMS         SMSConfig
	Storage     StorageConfig
	Monitoring  MonitoringConfig
	ExternalAPI ExternalAPIConfig
	Credential  GoogleCredential
}

// EmailConfig contains email service configuration
type EmailConfig struct {
	Provider    string // Default Email provider name (e.g., "smtp", "sendgrid")
	SenderEmail string // Default sender email address
	SenderName  string // Sender's display name
	APIKey      string // API key for email provider (if applicable)
	// TemplateDirectory string          // Path to email templates
	MaxRetries     int             // Max number of retry attempts
	RetryIntervals []time.Duration // Array of retry intervals
	SMTP           SMTPConfig      // SMTP provider configuration
	OAuthConfig    *OAuthConfig    // OAuth configuration for API-based providers
	Enabled        bool            // Enable/disable all email sending
}

// SMTPConfig holds SMTP server configurations
type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromEmail   string
	UseTLS      bool
	UseStartTLS bool
	Enabled     bool // Enable/disable SMTP email sending
}

// OAuthConfig holds OAuth2 configuration for API-based providers
type OAuthConfig struct {
	ClientID     string // OAuth client ID
	ClientSecret string // OAuth client secret
	TokenURL     string // URL to obtain OAuth token
	Enabled      bool   // Enable/disable OAuth email sending
}

// SMSConfig contains SMS service configuration
type SMSConfig struct {
	Provider      string
	AccountSID    string
	AuthToken     string
	PhoneNumber   string
	MaxRetries    int
	RetryInterval time.Duration
	Enabled       bool
}

// StorageConfig contains file storage configuration
type StorageConfig struct {
	Provider   string
	BucketName string
	Region     string
	BasePath   string
	Enabled    bool
}

// MonitoringConfig contains monitoring and logging configuration
type MonitoringConfig struct {
	Provider       string
	APIKey         string
	FlushInterval  time.Duration
	SamplingRate   float64
	EnabledMetrics []string
	Enabled        bool
}

// ExternalAPIConfig contains configuration for external API integrations
type ExternalAPIConfig struct {
	BaseURL      string
	APIKey       string
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
	Enabled      bool
}

// GoogleCalendarCredential holds the configuration for Google Calendar integration
type GoogleCredential struct {
	CredentialFilePath         string
	SharedDriveID              string
	CalendarID                 string
	CertificateFolderID        string
	RegistrationFormTemplateID string
	AttendanceFormTemplateID   string
	FeedbackFormTemplateID     string
	CertificateTemplateID      string
}

// func loadGoogleCredential() (*GoogleCalendarCredential, error) {
// 	env, err := loadEnvironment()
// 	if err != nil {
// 		return nil, errors.NewIntegrationError("google_calendar", "load_environment", err)
// 	}

// 	credentialFilePath := getEnv("GOOGLE_CALENDAR_CREDENTIAL_FILE", "")
// 	if credentialFilePath == "" {
// 		return nil, errors.NewNotFoundError("Google Calendar credential file path", "")
// 	}

// 	enabled := getEnvAsBool("GOOGLE_CALENDAR_ENABLED", env.Production)

// 	return &GoogleCalendarCredential{
// 		CredentialFilePath: credentialFilePath,
// 		Enabled:            enabled,
// 	}, nil
// }

func loadIntegrationConfig() (*IntegrationConfig, error) {
	env, err := loadEnvironment()
	if err != nil {
		return nil, errors.NewIntegrationError("integration", "load_environment", err)
	}

	creds, err := loadCredentials()
	if err != nil {
		return nil, errors.NewIntegrationError("integration", "load_credentials", err)
	}

	emailConfig := EmailConfig{
		Provider:    getEnv("EMAIL_PROVIDER", "smtp"),
		SenderEmail: getEnv("EMAIL_SENDER", "no-reply@tnprgpv.com"),
		SenderName:  getEnv("EMAIL_SENDER_NAME", "TNP RGPV"),
		APIKey:      getEnv("EMAIL_API_KEY", ""),
		// TemplateDirectory: getEnv("EMAIL_TEMPLATE_DIR", "./templates/email"),
		MaxRetries:     getEnvAsInt("EMAIL_MAX_RETRIES", 3),
		RetryIntervals: getEnvAsIntervals("EMAIL_RETRY_INTERVALS", []int{60, 300, 600}),
		Enabled:        getEnvAsBool("EMAIL_ENABLED", true),
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", "your-email@gmail.com"),
			// For Gmail, you need to use an App Password if 2FA is enabled
			// Go to https://myaccount.google.com/apppasswords to generate one
			Password:    getEnv("SMTP_PASSWORD", "your-app-password"),
			FromEmail:   getEnv("SMTP_FROM_EMAIL", "your-email@gmail.com"),
			UseTLS:      getEnvAsBool("SMTP_USE_TLS", false),     // Gmail prefers STARTTLS on port 587
			UseStartTLS: getEnvAsBool("SMTP_USE_STARTTLS", true), // Use STARTTLS for Gmail
		},
		OAuthConfig: &OAuthConfig{
			ClientID:     getEnv("OAUTH_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
			TokenURL:     getEnv("OAUTH_TOKEN_URL", ""),
		},
	}

	smsConfig := SMSConfig{
		Provider:      getEnv("SMS_PROVIDER", "twilio"),
		AccountSID:    getEnv("SMS_ACCOUNT_SID", ""),
		AuthToken:     getAPIKey(creds, "sms_provider", getEnv("SMS_AUTH_TOKEN", "")),
		PhoneNumber:   getEnv("SMS_PHONE_NUMBER", ""),
		MaxRetries:    getEnvAsInt("SMS_MAX_RETRIES", 3),
		RetryInterval: time.Duration(getEnvAsInt("SMS_RETRY_INTERVAL", 5)) * time.Second,
		Enabled:       getEnvAsBool("SMS_ENABLED", env.Production),
	}

	storageConfig := StorageConfig{
		Provider:   getEnv("STORAGE_PROVIDER", "s3"),
		BucketName: getEnv("STORAGE_BUCKET_NAME", "tnp-rgpv-files"),
		Region:     getEnv("AWS_REGION", "us-east-1"),
		BasePath:   getEnv("STORAGE_BASE_PATH", "uploads"),
		Enabled:    getEnvAsBool("STORAGE_ENABLED", true),
	}

	monitoringConfig := MonitoringConfig{
		Provider:      getEnv("MONITORING_PROVIDER", "cloudwatch"),
		APIKey:        getAPIKey(creds, "monitoring", getEnv("MONITORING_API_KEY", "")),
		FlushInterval: time.Duration(getEnvAsInt("MONITORING_FLUSH_INTERVAL", 10)) * time.Second,
		SamplingRate:  float64(getEnvAsInt("MONITORING_SAMPLING_RATE", 100)) / 100.0,
		EnabledMetrics: getEnvAsSlice(
			"MONITORING_ENABLED_METRICS",
			[]string{"api.requests", "db.queries", "errors"},
			",",
		),
		Enabled: getEnvAsBool("MONITORING_ENABLED", env.Production),
	}

	externalAPIConfig := ExternalAPIConfig{
		BaseURL:      getEnv("EXTERNAL_API_BASE_URL", "https://api.example.com"),
		APIKey:       getAPIKey(creds, "external_service", getEnv("EXTERNAL_API_KEY", "")),
		Timeout:      time.Duration(getEnvAsInt("EXTERNAL_API_TIMEOUT", 30)) * time.Second,
		MaxRetries:   getEnvAsInt("EXTERNAL_API_MAX_RETRIES", 3),
		RetryBackoff: time.Duration(getEnvAsInt("EXTERNAL_API_RETRY_BACKOFF", 5)) * time.Second,
		Enabled:      getEnvAsBool("EXTERNAL_API_ENABLED", false),
	}

	credentialconfig := GoogleCredential{
		CredentialFilePath:         getEnv("GOOGLE_CALENDAR_FILE", ""),
		CalendarID:                 getEnv("GOOGLE_CALENDAR_ID", ""),
		SharedDriveID:              getEnv("GOOGLE_CALENDAR_SHARED_DRIVE_ID", ""),
		CertificateFolderID:        getEnv("GOOGLE_CALENDAR_CERTIFICATE_FOLDER_ID", ""),
		RegistrationFormTemplateID: getEnv("GOOGLE_CALENDAR_REGISTRATION_FORM_TEMPLATE_ID", ""),
		AttendanceFormTemplateID:   getEnv("GOOGLE_CALENDAR_ATTENDANCE_FORM_TEMPLATE_ID", ""),
		FeedbackFormTemplateID:     getEnv("GOOGLE_CALENDAR_FEEDBACK_FORM_TEMPLATE_ID", ""),
		CertificateTemplateID:      getEnv("GOOGLE_CALENDAR_CERTIFICATE_TEMPLATE_ID", ""),
	}

	return &IntegrationConfig{
		Email:       emailConfig,
		SMS:         smsConfig,
		Storage:     storageConfig,
		Monitoring:  monitoringConfig,
		ExternalAPI: externalAPIConfig,
		Credential:  credentialconfig,
	}, nil
}

func getAPIKey(creds *ServerCredentials, keyName string, defaultValue string) string {
	if creds != nil && creds.APIKeys != nil {
		if key, exists := creds.APIKeys[keyName]; exists && key != "" {
			return key
		}
	}
	return defaultValue
}

func getEnvAsIntervals(key string, fallback []int) []time.Duration {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		parts := strings.Split(value, ",")
		var intervals []time.Duration
		for _, part := range parts {
			intValue, err := strconv.Atoi(strings.TrimSpace(part))
			if err == nil && intValue > 0 {
				intervals = append(intervals, time.Duration(intValue)*time.Second)
			}
		}
		if len(intervals) > 0 {
			return intervals
		}
	}
	// fallback
	var defaultIntervals []time.Duration
	for _, sec := range fallback {
		defaultIntervals = append(defaultIntervals, time.Duration(sec)*time.Second)
	}
	return defaultIntervals
}

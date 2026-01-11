package config

import (
	"os"
	"strings"
)

// FeatureFlags represents all feature toggles used in the application
type FeatureFlags struct {
	EnableAdvancedSearch     bool
	EnableNotifications      bool
	EnableCaching            bool
	EnableRateLimiting       bool
	EnableUserTracking       bool
	EnableDocumentGeneration bool
	ExperimentalFeatures     map[string]bool
}

// loadFeatureFlags initializes feature flags based on environment variables
func loadFeatureFlags() (*FeatureFlags, error) {
	flags := &FeatureFlags{
		EnableAdvancedSearch:     getEnvAsBool("FEATURE_ADVANCED_SEARCH", false),
		EnableNotifications:      getEnvAsBool("FEATURE_NOTIFICATIONS", true),
		EnableCaching:            getEnvAsBool("FEATURE_CACHING", true),
		EnableRateLimiting:       getEnvAsBool("FEATURE_RATE_LIMITING", true),
		EnableUserTracking:       getEnvAsBool("FEATURE_USER_TRACKING", false),
		EnableDocumentGeneration: getEnvAsBool("FEATURE_DOCUMENT_GENERATION", true),
		ExperimentalFeatures:     loadExperimentalFeatures(),
	}

	return flags, nil
}

// loadExperimentalFeatures loads any experimental features from a comma-separated list
func loadExperimentalFeatures() map[string]bool {
	features := make(map[string]bool)
	
	// Get experimental features from environment variable
	expFeaturesStr := os.Getenv("EXPERIMENTAL_FEATURES")
	if expFeaturesStr == "" {
		return features
	}
	
	// Parse comma-separated list
	expFeaturesList := strings.Split(expFeaturesStr, ",")
	for _, feature := range expFeaturesList {
		feature = strings.TrimSpace(feature)
		if feature != "" {
			features[feature] = true
		}
	}
	
	return features
}

// IsExperimentalFeatureEnabled checks if a specific experimental feature is enabled
func (f *FeatureFlags) IsExperimentalFeatureEnabled(featureName string) bool {
	if enabled, exists := f.ExperimentalFeatures[featureName]; exists {
		return enabled
	}
	return false
}
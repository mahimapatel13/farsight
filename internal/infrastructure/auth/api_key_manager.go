package auth

import (
	"context"
	"errors"
	"time"
)

// APIKeyInfo holds metadata about an API key
type APIKeyInfo struct {
	ClientID  string   `json:"client_id"`
	Scopes    []string `json:"scopes"`
	CreatedAt time.Time
	ExpiresAt time.Time
	IsRevoked bool
}

// APIKeyManager is responsible for managing and validating API keys
type APIKeyManager struct {
	store map[string]*APIKeyInfo // In-memory store (replace with DB in production)
}

// NewAPIKeyManager creates a new APIKeyManager
func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		store: make(map[string]*APIKeyInfo),
	}
}

// ValidateKey checks if the provided API key is valid
func (m *APIKeyManager) ValidateKey(ctx context.Context, apiKey string) (*APIKeyInfo, error) {
	keyInfo, exists := m.store[apiKey]
	if !exists {
		return nil, errors.New("API key not found")
	}

	// Check if key is revoked
	if keyInfo.IsRevoked {
		return nil, errors.New("API key has been revoked")
	}

	// Check if key is expired
	if keyInfo.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("API key has expired")
	}

	return keyInfo, nil
}

// AddKey adds a new API key to the store
func (m *APIKeyManager) AddKey(apiKey string, keyInfo *APIKeyInfo) error {
	if _, exists := m.store[apiKey]; exists {
		return errors.New("API key already exists")
	}
	m.store[apiKey] = keyInfo
	return nil
}

// RevokeKey revokes an existing API key
func (m *APIKeyManager) RevokeKey(apiKey string) error {
	keyInfo, exists := m.store[apiKey]
	if !exists {
		return errors.New("API key not found")
	}

	keyInfo.IsRevoked = true
	return nil
}

// HasScope checks if the API key has the required scope(s)
func (m *APIKeyManager) HasScope(apiKey string, requiredScope string) (bool, error) {
	keyInfo, err := m.ValidateKey(context.Background(), apiKey)
	if err != nil {
		return false, err
	}

	for _, scope := range keyInfo.Scopes {
		if scope == requiredScope {
			return true, nil
		}
	}
	return false, nil
}

// ListKeys returns a list of all registered API keys
func (m *APIKeyManager) ListKeys() map[string]*APIKeyInfo {
	return m.store
}


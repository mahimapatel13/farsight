package auth

import (
	"errors"
	"time"

	"slices"

	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type CustomClaims struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"role"`
	TokenType string   `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

// NewJWTProvider creates a new JWT provider with the given settings
func NewJWTProvider(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *JWTProvider {
	return &JWTProvider{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair creates a new access and refresh token pair
func (p *JWTProvider) GenerateTokenPair(userID string, roles []string) (*TokenPair, error) {
	// Create access token
	accessClaims := CustomClaims{
		UserID:    userID,
		Roles:     roles,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   "budget_planner",
			Audience: jwt.ClaimStrings{"budget-planner-client"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(p.accessSecret)
	if err != nil {
		return nil, err
	}

	// Create refresh token with longer expiry but fewer claims
	refreshClaims := CustomClaims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   "budget_planner",
			Audience: jwt.ClaimStrings{"budget-planner-client"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(p.refreshSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(p.accessExpiry.Seconds()),
	}, nil
}

// contains checks if a string is present in a slice of strings
func contains(audience jwt.ClaimStrings, target string) bool {
	return slices.Contains(audience, target)
}

// ValidateToken validates the given token and returns the claims
func (p *JWTProvider) ValidateToken(tokenString string, isRefresh bool) (*CustomClaims, error) {
	secret := p.accessSecret
	if isRefresh {
		secret = p.refreshSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, errors.New("invalid token signature")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, errors.New("failed to parse token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Validate issuer and audience
	if claims.Issuer != "budget_planner" {
		return nil, errors.New("invalid token issuer")
	}

	// Verify audience
	expectedAudience := "budget-planner-client"
	if !contains(claims.Audience, expectedAudience) {
		return nil, errors.New("invalid token audience")
	}

	return claims, nil
}

// RefreshTokens generates a new token pair using a valid refresh token
func (p *JWTProvider) RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	// Validate the refresh token (isRefresh = true)
	claims, err := p.ValidateToken(refreshTokenString, true)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Check if the refresh token has the correct audience
	expectedAudience := "budget-planner-client"
	if !contains(claims.Audience, expectedAudience) {
		return nil, errors.New("invalid refresh token audience")
	}

	// Check if the token is of type 'refresh'
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid refresh token type")
	}

	// Regenerate a new token pair with the same userID and role
	tokenPair, err := p.GenerateTokenPair(claims.UserID, claims.Roles)
	if err != nil {
		return nil, errors.New("failed to generate new token pair")
	}

	return tokenPair, nil
}


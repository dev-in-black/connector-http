package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OAuth2Auth implements OAuth2 Client Credentials flow
type OAuth2Auth struct {
	config      *clientcredentials.Config
	tokenSource oauth2.TokenSource
	mu          sync.RWMutex
}

// NewOAuth2Auth creates a new OAuth2 authenticator with token caching
func NewOAuth2Auth(cfg *OAuth2Config) (*OAuth2Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("OAuth2Config is required")
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.TokenURL == "" {
		return nil, fmt.Errorf("OAuth2 requires clientID, clientSecret, and tokenURL")
	}

	config := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		Scopes:       cfg.Scopes,
	}

	// Create token source with automatic caching and refresh
	// TokenSource is thread-safe and handles token expiration automatically
	tokenSource := config.TokenSource(context.Background())

	return &OAuth2Auth{
		config:      config,
		tokenSource: tokenSource,
	}, nil
}

// Authenticate adds OAuth2 Bearer token authentication to the request
func (a *OAuth2Auth) Authenticate(ctx context.Context, req *http.Request) error {
	// Token() is thread-safe and returns cached token if valid
	// Automatically requests new token if expired
	token, err := a.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get OAuth2 token: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return nil
}

// Type returns the auth type
func (a *OAuth2Auth) Type() string {
	return "oauth2"
}

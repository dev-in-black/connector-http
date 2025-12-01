package auth

import (
	"context"
	"fmt"
	"net/http"
)

// Manager handles authentication for HTTP requests
type Manager interface {
	// Authenticate adds authentication to the request
	Authenticate(ctx context.Context, req *http.Request) error

	// Type returns the authentication type
	Type() string
}

// Config holds authentication configuration
type Config struct {
	Type           string
	BasicUsername  string
	BasicPassword  string
	BearerToken    string
	OAuth2Config   *OAuth2Config
}

// OAuth2Config holds OAuth2 client credentials configuration
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	Scopes       []string
}

// NewManager creates an authentication manager based on the config
func NewManager(cfg Config) (Manager, error) {
	switch cfg.Type {
	case "none":
		return &NoneAuth{}, nil
	case "basic":
		if cfg.BasicUsername == "" || cfg.BasicPassword == "" {
			return nil, fmt.Errorf("basic auth requires username and password")
		}
		return NewBasicAuth(cfg.BasicUsername, cfg.BasicPassword), nil
	case "bearer":
		if cfg.BearerToken == "" {
			return nil, fmt.Errorf("bearer auth requires token")
		}
		return NewBearerAuth(cfg.BearerToken), nil
	case "oauth2":
		if cfg.OAuth2Config == nil {
			return nil, fmt.Errorf("oauth2 auth requires OAuth2Config")
		}
		return NewOAuth2Auth(cfg.OAuth2Config)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.Type)
	}
}

// NoneAuth is a no-op authenticator
type NoneAuth struct{}

// Authenticate does nothing for none auth
func (a *NoneAuth) Authenticate(ctx context.Context, req *http.Request) error {
	return nil
}

// Type returns the auth type
func (a *NoneAuth) Type() string {
	return "none"
}

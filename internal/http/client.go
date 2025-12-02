package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dev-in-black/connector-http/internal/auth"
)

// Config holds HTTP client configuration
type Config struct {
	Timeout         time.Duration
	MaxIdleConns    int
	MaxConnsPerHost int
}

// Client wraps an HTTP client with authentication and header management
type Client struct {
	httpClient    *http.Client
	authManager   auth.Manager
	staticHeaders map[string]string
	envHeaders    map[string]string
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(cfg Config, authMgr auth.Manager, staticHeaders, envHeaders map[string]string) *Client {
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
		authManager:   authMgr,
		staticHeaders: staticHeaders,
		envHeaders:    envHeaders,
	}
}

// Post sends an HTTP POST request with authentication and custom headers
func (c *Client) Post(ctx context.Context, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Apply static headers (from config)
	for k, v := range c.staticHeaders {
		req.Header.Set(k, v)
	}

	// Apply environment headers (override static)
	for k, v := range c.envHeaders {
		req.Header.Set(k, v)
	}

	// Apply authentication
	if err := c.authManager.Authenticate(ctx, req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

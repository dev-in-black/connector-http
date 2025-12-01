package http

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries        int
	BackoffBase       time.Duration
	BackoffMax        time.Duration
	RetryOn5xx        bool
	RetryOn429        bool
	RetryOnNetworkErr bool
}

// RetryEngine handles retry logic with exponential backoff
type RetryEngine struct {
	config RetryConfig
}

// NewRetryEngine creates a new retry engine
func NewRetryEngine(cfg RetryConfig) *RetryEngine {
	return &RetryEngine{config: cfg}
}

// Do executes the given function with retry logic
func (r *RetryEngine) Do(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			backoff := r.calculateBackoff(attempt)

			select {
			case <-time.After(backoff):
				// Continue to retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Execute the function
		resp, err := fn()

		// Success case: 2xx status
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Store last error and response
		lastErr = err
		lastResp = resp

		// Check if error is retryable
		if !r.isRetryable(err, resp) {
			if resp != nil {
				return resp, fmt.Errorf("non-retryable error: status %d", resp.StatusCode)
			}
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		// Close response body to reuse connection
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	// Max retries exceeded
	if lastResp != nil {
		return lastResp, fmt.Errorf("max retries (%d) exceeded, last status: %d", r.config.MaxRetries, lastResp.StatusCode)
	}
	return nil, fmt.Errorf("max retries (%d) exceeded: %w", r.config.MaxRetries, lastErr)
}

// calculateBackoff calculates exponential backoff duration
func (r *RetryEngine) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: 2^attempt * base
	backoff := time.Duration(math.Pow(2, float64(attempt))) * r.config.BackoffBase

	// Cap at maximum backoff
	if backoff > r.config.BackoffMax {
		backoff = r.config.BackoffMax
	}

	return backoff
}

// isRetryable determines if an error/response is retryable
func (r *RetryEngine) isRetryable(err error, resp *http.Response) bool {
	// Network errors are retryable if configured
	if err != nil {
		if r.config.RetryOnNetworkErr {
			// Check for net.Error (includes timeouts, connection errors)
			if _, ok := err.(net.Error); ok {
				return true
			}
		}
		// Non-network errors are not retryable
		return false
	}

	// HTTP status code based retryability
	if resp != nil {
		// 5xx errors (server errors) are retryable if configured
		if r.config.RetryOn5xx && resp.StatusCode >= 500 && resp.StatusCode < 600 {
			return true
		}

		// 429 Too Many Requests is retryable if configured
		if r.config.RetryOn429 && resp.StatusCode == http.StatusTooManyRequests {
			return true
		}

		// 4xx errors (except 429) are client errors, not retryable
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return false
		}
	}

	// Default: not retryable
	return false
}

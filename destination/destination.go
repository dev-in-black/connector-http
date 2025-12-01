package destination

import (
	"context"
	"fmt"
	stdhttp "net/http"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-commons/config"
	"github.com/conduitio/conduit-commons/opencdc"
	"github.com/conduitio/conduit-connector-http/internal/auth"
	"github.com/conduitio/conduit-connector-http/internal/http"
	"github.com/conduitio/conduit-connector-http/internal/response"
)

// Destination implements the Conduit destination interface for HTTP endpoints
type Destination struct {
	sdk.UnimplementedDestination

	config         Config
	httpClient     *http.Client
	authManager    auth.Manager
	responseWriter *response.Writer
	retryEngine    *http.RetryEngine
}

// NewDestination creates a new HTTP destination
func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{})
}

// Config returns the configuration structure
func (d *Destination) Config() sdk.DestinationConfig {
	return &d.config
}

// LifecycleOnCreated is called when the connector is created for the first time
func (d *Destination) LifecycleOnCreated(ctx context.Context, cfg config.Config) error {
	// No special initialization needed on first creation
	return nil
}

// LifecycleOnUpdated is called when the connector configuration is updated
func (d *Destination) LifecycleOnUpdated(ctx context.Context, configBefore, configAfter config.Config) error {
	// No special handling needed for config updates
	return nil
}

// LifecycleOnDeleted is called when the connector is deleted
func (d *Destination) LifecycleOnDeleted(ctx context.Context, cfg config.Config) error {
	// No special cleanup needed on deletion
	return nil
}

// Open prepares the destination for writing records
func (d *Destination) Open(ctx context.Context) error {
	sdk.Logger(ctx).Info().Msg("Opening HTTP destination")

	// Load custom headers from environment
	d.config.LoadEnvHeaders()

	// Validate configuration
	if err := d.config.Validate(ctx); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	sdk.Logger(ctx).Info().
		Str("url", d.config.URL).
		Str("method", d.config.Method).
		Str("authType", d.config.AuthType).
		Msg("HTTP destination configured")

	// Initialize authentication manager
	authConfig := auth.Config{
		Type:          d.config.AuthType,
		BasicUsername: d.config.BasicUsername,
		BasicPassword: d.config.BasicPassword,
		BearerToken:   d.config.BearerToken,
	}

	if d.config.AuthType == "oauth2" {
		authConfig.OAuth2Config = &auth.OAuth2Config{
			ClientID:     d.config.OAuth2ClientID,
			ClientSecret: d.config.OAuth2ClientSecret,
			TokenURL:     d.config.OAuth2TokenURL,
			Scopes:       d.config.GetOAuth2Scopes(),
		}
	}

	var err error
	d.authManager, err = auth.NewManager(authConfig)
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Initialize HTTP client
	httpConfig := http.Config{
		Timeout:         d.config.Timeout,
		MaxIdleConns:    d.config.MaxIdleConns,
		MaxConnsPerHost: d.config.MaxConnsPerHost,
	}

	d.httpClient = http.NewClient(
		httpConfig,
		d.authManager,
		d.config.StaticHeaders,
		d.config.LoadedEnvHeaders(),
	)

	// Initialize retry engine
	retryConfig := http.RetryConfig{
		MaxRetries:        d.config.MaxRetries,
		BackoffBase:       d.config.RetryBackoffBase,
		BackoffMax:        d.config.RetryBackoffMax,
		RetryOn5xx:        d.config.RetryOn5xx,
		RetryOn429:        d.config.RetryOn429,
		RetryOnNetworkErr: d.config.RetryOnNetworkErr,
	}

	d.retryEngine = http.NewRetryEngine(retryConfig)

	// Initialize response writer
	responseConfig := response.Config{
		OutputPath:             d.config.ResponseOutputPath,
		SuccessFile:            d.config.SuccessFile,
		ErrorFile:              d.config.ErrorFile,
		IncludeResponseHeaders: d.config.IncludeResponseHeaders,
		IncludeRequestMetadata: d.config.IncludeRequestMetadata,
	}

	d.responseWriter, err = response.NewWriter(responseConfig)
	if err != nil {
		return fmt.Errorf("failed to create response writer: %w", err)
	}

	sdk.Logger(ctx).Info().Msg("HTTP destination opened successfully")
	return nil
}

// Write sends records to the HTTP endpoint
func (d *Destination) Write(ctx context.Context, records []opencdc.Record) (int, error) {
	logger := sdk.Logger(ctx)

	for i, record := range records {
		// Prepare request body from record payload
		body, err := d.prepareRequestBody(record)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to prepare request body")
			// Write error to error file
			d.responseWriter.WriteError(record, err, nil)
			return i, fmt.Errorf("failed to prepare request body: %w", err)
		}

		// Send HTTP request with retry logic
		resp, err := d.retryEngine.Do(ctx, func() (*stdhttp.Response, error) {
			return d.httpClient.Post(ctx, d.config.URL, body)
		})

		if err != nil {
			logger.Error().Err(err).Msg("HTTP request failed after retries")
			// Write error to error file
			d.responseWriter.WriteError(record, err, resp)
			return i, fmt.Errorf("HTTP request failed: %w", err)
		}

		// Route response based on status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logger.Debug().
				Int("status", resp.StatusCode).
				Msg("HTTP request successful")

			if err := d.responseWriter.WriteSuccess(record, resp); err != nil {
				logger.Warn().Err(err).Msg("Failed to write success response")
			}
		} else {
			logger.Warn().
				Int("status", resp.StatusCode).
				Msg("HTTP request returned non-2xx status")

			err := fmt.Errorf("HTTP %d", resp.StatusCode)
			if writeErr := d.responseWriter.WriteError(record, err, resp); writeErr != nil {
				logger.Warn().Err(writeErr).Msg("Failed to write error response")
			}
		}
	}

	return len(records), nil
}

// Teardown cleans up resources
func (d *Destination) Teardown(ctx context.Context) error {
	sdk.Logger(ctx).Info().Msg("Tearing down HTTP destination")

	if d.responseWriter != nil {
		if err := d.responseWriter.Close(); err != nil {
			return fmt.Errorf("failed to close response writer: %w", err)
		}
	}

	sdk.Logger(ctx).Info().Msg("HTTP destination torn down successfully")
	return nil
}

// prepareRequestBody extracts the payload from the record
func (d *Destination) prepareRequestBody(record opencdc.Record) ([]byte, error) {
	// Use the After payload (for inserts/updates)
	if d.config.UsePayloadAfter && record.Payload.After != nil {
		return record.Payload.After.Bytes(), nil
	}

	// Fallback to Before payload (for deletes)
	if record.Payload.Before != nil {
		return record.Payload.Before.Bytes(), nil
	}

	return nil, fmt.Errorf("record has no payload")
}

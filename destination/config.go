package destination

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

// Config holds the configuration for the HTTP destination connector
type Config struct {
	sdk.UnimplementedDestinationConfig

	// Core HTTP Settings
	URL             string        `json:"url" validate:"required,url"`
	Method          string        `json:"method" default:"POST"`
	Timeout         time.Duration `json:"timeout" default:"30s"`
	MaxIdleConns    int           `json:"maxIdleConns" default:"100"`
	MaxConnsPerHost int           `json:"maxConnsPerHost" default:"10"`

	// Authentication
	AuthType string `json:"authType" default:"none"`

	// Basic Auth (from environment)
	BasicUsername string `json:"basicUsername"`
	BasicPassword string `json:"basicPassword"`

	// Bearer Token (from environment)
	BearerToken string `json:"bearerToken"`

	// OAuth2 Client Credentials
	OAuth2ClientID     string `json:"oauth2ClientId"`
	OAuth2ClientSecret string `json:"oauth2ClientSecret"`
	OAuth2TokenURL     string `json:"oauth2TokenUrl"`
	OAuth2Scopes       string `json:"oauth2Scopes"` // Comma-separated

	// Custom Headers
	StaticHeaders   map[string]string `json:"staticHeaders"` // From config
	EnvHeaderPrefix string            `json:"envHeaderPrefix" default:"HTTP_HEADER_"`
	envHeaders      map[string]string // Loaded from environment

	// Request Body Transformation
	BodyTemplate    string `json:"bodyTemplate"`
	UsePayloadAfter bool   `json:"usePayloadAfter" default:"true"`

	// Schema Validation
	ValidateRequest   bool   `json:"validateRequest" default:"false"`
	ValidateResponse  bool   `json:"validateResponse" default:"false"`
	RequestSchemaURL  string `json:"requestSchemaUrl"`
	ResponseSchemaURL string `json:"responseSchemaUrl"`
	SchemaRegistryURL string `json:"schemaRegistryUrl"`
	SchemaType        string `json:"schemaType" default:"json"`
	FailOnValidation  bool   `json:"failOnValidation" default:"true"`

	// Retry Configuration
	MaxRetries        int           `json:"maxRetries" default:"3"`
	RetryBackoffBase  time.Duration `json:"retryBackoffBase" default:"1s"`
	RetryBackoffMax   time.Duration `json:"retryBackoffMax" default:"30s"`
	RetryOn5xx        bool          `json:"retryOn5xx" default:"true"`
	RetryOn429        bool          `json:"retryOn429" default:"true"`
	RetryOnNetworkErr bool          `json:"retryOnNetworkErr" default:"true"`

	// Kafka Configuration for Response Publishing
	KafkaEnabled       bool   `json:"kafkaEnabled" default:"false"`
	KafkaBrokers       string `json:"kafkaBrokers"` // Comma-separated list of brokers
	KafkaTopic         string `json:"kafkaTopic" default:"http-responses"`
	KafkaClientID      string `json:"kafkaClientId" default:"http-connector"`
	KafkaCompression   string `json:"kafkaCompression" default:"snappy"` // none, gzip, snappy, lz4, zstd
	KafkaEnableIdempotence bool `json:"kafkaEnableIdempotence" default:"true"`

	// Kafka Authentication (SASL)
	KafkaSASLEnabled   bool   `json:"kafkaSaslEnabled" default:"false"`
	KafkaSASLMechanism string `json:"kafkaSaslMechanism" default:"PLAIN"` // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	KafkaSASLUsername  string `json:"kafkaSaslUsername"`
	KafkaSASLPassword  string `json:"kafkaSaslPassword"`

	// Kafka TLS
	KafkaTLSEnabled    bool `json:"kafkaTlsEnabled" default:"false"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate(ctx context.Context) error {
	if c.URL == "" {
		return fmt.Errorf("url is required")
	}

	validMethods := map[string]bool{"POST": true, "PUT": true, "PATCH": true}
	if !validMethods[c.Method] {
		return fmt.Errorf("invalid method: %s (must be POST, PUT, or PATCH)", c.Method)
	}

	validAuthTypes := map[string]bool{"none": true, "basic": true, "bearer": true, "oauth2": true}
	if !validAuthTypes[c.AuthType] {
		return fmt.Errorf("invalid authType: %s (must be none, basic, bearer, or oauth2)", c.AuthType)
	}

	// Validate auth-specific requirements
	if c.AuthType == "basic" {
		if c.BasicUsername == "" || c.BasicPassword == "" {
			return fmt.Errorf("basicUsername and basicPassword are required for basic auth")
		}
	}

	if c.AuthType == "bearer" {
		if c.BearerToken == "" {
			return fmt.Errorf("bearerToken is required for bearer auth")
		}
	}

	if c.AuthType == "oauth2" {
		if c.OAuth2ClientID == "" || c.OAuth2ClientSecret == "" || c.OAuth2TokenURL == "" {
			return fmt.Errorf("oauth2ClientId, oauth2ClientSecret, and oauth2TokenUrl are required for oauth2 auth")
		}
	}

	// Validate retry configuration
	if c.MaxRetries < 0 || c.MaxRetries > 10 {
		return fmt.Errorf("maxRetries must be between 0 and 10")
	}

	validSchemaTypes := map[string]bool{"json": true, "avro": true}
	if !validSchemaTypes[c.SchemaType] {
		return fmt.Errorf("invalid schemaType: %s (must be json or avro)", c.SchemaType)
	}

	// Validate Kafka configuration if enabled
	if c.KafkaEnabled {
		if c.KafkaBrokers == "" {
			return fmt.Errorf("kafkaBrokers is required when kafkaEnabled is true")
		}
		if c.KafkaTopic == "" {
			return fmt.Errorf("kafkaTopic is required when kafkaEnabled is true")
		}

		validCompressions := map[string]bool{"none": true, "gzip": true, "snappy": true, "lz4": true, "zstd": true}
		if !validCompressions[c.KafkaCompression] {
			return fmt.Errorf("invalid kafkaCompression: %s (must be none, gzip, snappy, lz4, or zstd)", c.KafkaCompression)
		}

		if c.KafkaSASLEnabled {
			validMechanisms := map[string]bool{"PLAIN": true, "SCRAM-SHA-256": true, "SCRAM-SHA-512": true}
			if !validMechanisms[c.KafkaSASLMechanism] {
				return fmt.Errorf("invalid kafkaSaslMechanism: %s (must be PLAIN, SCRAM-SHA-256, or SCRAM-SHA-512)", c.KafkaSASLMechanism)
			}
			if c.KafkaSASLUsername == "" || c.KafkaSASLPassword == "" {
				return fmt.Errorf("kafkaSaslUsername and kafkaSaslPassword are required when kafkaSaslEnabled is true")
			}
		}
	}

	return nil
}

// LoadEnvHeaders loads custom headers from environment variables with the configured prefix
func (c *Config) LoadEnvHeaders() {
	c.envHeaders = make(map[string]string)

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(key, c.EnvHeaderPrefix) {
			// Remove prefix to get header name
			headerName := strings.TrimPrefix(key, c.EnvHeaderPrefix)
			// Convert underscore to hyphen for standard HTTP headers
			headerName = strings.ReplaceAll(headerName, "_", "-")
			c.envHeaders[headerName] = value
		}
	}
}

// LoadedEnvHeaders returns the loaded environment headers
func (c *Config) LoadedEnvHeaders() map[string]string {
	return c.envHeaders
}

// GetOAuth2Scopes parses the comma-separated scopes string
func (c *Config) GetOAuth2Scopes() []string {
	if c.OAuth2Scopes == "" {
		return []string{}
	}
	scopes := strings.Split(c.OAuth2Scopes, ",")
	// Trim whitespace from each scope
	for i := range scopes {
		scopes[i] = strings.TrimSpace(scopes[i])
	}
	return scopes
}

// GetKafkaBrokers parses the comma-separated brokers string
func (c *Config) GetKafkaBrokers() []string {
	if c.KafkaBrokers == "" {
		return []string{}
	}
	brokers := strings.Split(c.KafkaBrokers, ",")
	// Trim whitespace from each broker
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return brokers
}

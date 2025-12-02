package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

// Config holds the configuration for the Kafka producer
type Config struct {
	Brokers           []string
	Topic             string
	ClientID          string
	Compression       string
	EnableIdempotence bool
	SASLEnabled       bool
	SASLMechanism     string
	SASLUsername      string
	SASLPassword      string
	TLSEnabled        bool
}

// Producer wraps the Kafka producer client
type Producer struct {
	client *kgo.Client
	topic  string
}

// ResponseMessage represents the HTTP response to be published to Kafka
type ResponseMessage struct {
	StatusCode      int               `json:"status_code"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Body            string            `json:"body"`
	RequestURL      string            `json:"request_url"`
	RequestMethod   string            `json:"request_method"`
	Timestamp       time.Time         `json:"timestamp"`
}

// NewProducer creates a new Kafka producer
func NewProducer(ctx context.Context, cfg Config) (*Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ClientID),
		kgo.AllowAutoTopicCreation(),
	}

	// Set compression
	switch cfg.Compression {
	case "gzip":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.GzipCompression()))
	case "snappy":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
	case "lz4":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
	case "zstd":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
	case "none":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.NoCompression()))
	default:
		opts = append(opts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
	}

	// Enable idempotent producer
	if cfg.EnableIdempotence {
		opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
		opts = append(opts, kgo.DisableIdempotentWrite())
	}

	// Configure SASL authentication
	if cfg.SASLEnabled {
		switch cfg.SASLMechanism {
		case "PLAIN":
			opts = append(opts, kgo.SASL(plain.Auth{
				User: cfg.SASLUsername,
				Pass: cfg.SASLPassword,
			}.AsMechanism()))
		case "SCRAM-SHA-256":
			mechanism := scram.Auth{
				User: cfg.SASLUsername,
				Pass: cfg.SASLPassword,
			}.AsSha256Mechanism()
			opts = append(opts, kgo.SASL(mechanism))
		case "SCRAM-SHA-512":
			mechanism := scram.Auth{
				User: cfg.SASLUsername,
				Pass: cfg.SASLPassword,
			}.AsSha512Mechanism()
			opts = append(opts, kgo.SASL(mechanism))
		default:
			return nil, fmt.Errorf("unsupported SASL mechanism: %s", cfg.SASLMechanism)
		}
	}

	// Configure TLS
	if cfg.TLSEnabled {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}))
	}

	// Create Kafka client
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	return &Producer{
		client: client,
		topic:  cfg.Topic,
	}, nil
}

// PublishResponse publishes an HTTP response to Kafka
func (p *Producer) PublishResponse(ctx context.Context, statusCode int, responseHeaders map[string][]string, body []byte, requestURL, requestMethod string, recordHeaders map[string]string) error {
	// Convert HTTP response headers to map[string]string for JSON serialization
	flatResponseHeaders := make(map[string]string)
	for key, values := range responseHeaders {
		if len(values) > 0 {
			flatResponseHeaders[key] = values[0] // Take first value for simplicity
		}
	}

	// Create response message (record headers go to Kafka headers, not JSON body)
	msg := ResponseMessage{
		StatusCode:      statusCode,
		ResponseHeaders: flatResponseHeaders,
		Body:            string(body),
		RequestURL:      requestURL,
		RequestMethod:   requestMethod,
		Timestamp:       time.Now(),
	}

	// Serialize to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response message: %w", err)
	}

	// Create Kafka record with record headers as Kafka headers
	record := &kgo.Record{
		Topic: p.topic,
		Value: data,
		Key:   []byte(fmt.Sprintf("%s-%d", requestURL, time.Now().UnixNano())),
	}

	// Add record headers as Kafka record headers for easier filtering
	for key, value := range recordHeaders {
		record.Headers = append(record.Headers, kgo.RecordHeader{
			Key:   key,
			Value: []byte(value),
		})
	}

	// Produce record
	results := p.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce message to Kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() {
	if p.client != nil {
		p.client.Close()
	}
}

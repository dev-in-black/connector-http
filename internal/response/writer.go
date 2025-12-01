package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/conduitio/conduit-commons/opencdc"
	"github.com/google/uuid"
)

// Config holds configuration for the response writer
type Config struct {
	OutputPath             string
	SuccessFile            string
	ErrorFile              string
	IncludeResponseHeaders bool
	IncludeRequestMetadata bool
}

// Writer writes HTTP responses to NDJSON files
type Writer struct {
	config      Config
	successFile *os.File
	errorFile   *os.File
	mu          sync.Mutex
}

// ResponseRecord represents an HTTP response record written to file
type ResponseRecord struct {
	OriginalRecord  map[string]interface{} `json:"original_record,omitempty"`
	HTTPStatus      int                    `json:"http_status,omitempty"`
	ResponseBody    string                 `json:"response_body,omitempty"`
	ResponseHeaders map[string][]string    `json:"response_headers,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	CorrelationID   string                 `json:"correlation_id"`
	Timestamp       string                 `json:"timestamp"`
}

// NewWriter creates a new response writer
func NewWriter(cfg Config) (*Writer, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open success file
	successPath := filepath.Join(cfg.OutputPath, cfg.SuccessFile)
	successFile, err := os.OpenFile(successPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open success file: %w", err)
	}

	// Open error file
	errorPath := filepath.Join(cfg.OutputPath, cfg.ErrorFile)
	errorFile, err := os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		successFile.Close()
		return nil, fmt.Errorf("failed to open error file: %w", err)
	}

	return &Writer{
		config:      cfg,
		successFile: successFile,
		errorFile:   errorFile,
	}, nil
}

// WriteSuccess writes a successful HTTP response to the success file
func (w *Writer) WriteSuccess(record opencdc.Record, resp *http.Response) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body.Close()

	respRecord := ResponseRecord{
		HTTPStatus:    resp.StatusCode,
		ResponseBody:  string(bodyBytes),
		CorrelationID: generateCorrelationID(record),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	if w.config.IncludeResponseHeaders {
		respRecord.ResponseHeaders = resp.Header
	}

	if w.config.IncludeRequestMetadata {
		respRecord.OriginalRecord = recordToMap(record)
	}

	// Write as NDJSON
	encoder := json.NewEncoder(w.successFile)
	if err := encoder.Encode(respRecord); err != nil {
		return fmt.Errorf("failed to write success record: %w", err)
	}

	return nil
}

// WriteError writes a failed HTTP response or error to the error file
func (w *Writer) WriteError(record opencdc.Record, err error, resp *http.Response) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	respRecord := ResponseRecord{
		ErrorMessage:  err.Error(),
		CorrelationID: generateCorrelationID(record),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	if resp != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		respRecord.HTTPStatus = resp.StatusCode
		respRecord.ResponseBody = string(bodyBytes)

		if w.config.IncludeResponseHeaders {
			respRecord.ResponseHeaders = resp.Header
		}
	}

	if w.config.IncludeRequestMetadata {
		respRecord.OriginalRecord = recordToMap(record)
	}

	encoder := json.NewEncoder(w.errorFile)
	if err := encoder.Encode(respRecord); err != nil {
		return fmt.Errorf("failed to write error record: %w", err)
	}

	return nil
}

// Close closes the response writer files
func (w *Writer) Close() error {
	var errs []error

	if err := w.successFile.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close success file: %w", err))
	}

	if err := w.errorFile.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close error file: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing files: %v", errs)
	}

	return nil
}

// generateCorrelationID generates a correlation ID from the record or creates a new UUID
func generateCorrelationID(record opencdc.Record) string {
	// Try to use record position as correlation ID
	if record.Position != nil && len(record.Position) > 0 {
		return string(record.Position)
	}

	// Try to use record key
	if record.Key != nil {
		keyBytes := record.Key.Bytes()
		if len(keyBytes) > 0 {
			return string(keyBytes)
		}
	}

	// Generate new UUID
	return uuid.New().String()
}

// recordToMap converts a Conduit record to a map for JSON serialization
func recordToMap(record opencdc.Record) map[string]interface{} {
	result := make(map[string]interface{})

	if record.Position != nil {
		result["position"] = string(record.Position)
	}

	if record.Key != nil {
		result["key"] = string(record.Key.Bytes())
	}

	if record.Payload.Before != nil {
		result["payload_before"] = string(record.Payload.Before.Bytes())
	}

	if record.Payload.After != nil {
		result["payload_after"] = string(record.Payload.After.Bytes())
	}

	if len(record.Metadata) > 0 {
		result["metadata"] = record.Metadata
	}

	return result
}

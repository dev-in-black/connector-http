# Conduit HTTP Sink Connector

A production-ready HTTP sink connector for [Conduit](https://conduit.io) that sends records to HTTP endpoints with advanced features including authentication, retry logic, and response handling.

## Features

- **HTTP Methods**: POST, PUT, PATCH support
- **Authentication**:
  - None
  - Basic Authentication
  - Bearer Token
  - OAuth2 Client Credentials
- **Custom Headers**: From environment variables and static configuration
- **Retry Logic**: Configurable exponential backoff with retryable error detection
- **Response Handling**: Success and error responses written to NDJSON files
- **Connection Pooling**: Efficient HTTP client with configurable connection limits
- **Thread-Safe**: Safe for concurrent use

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/conduitio/conduit-connector-http.git
cd conduit-connector-http

# Install dependencies
make install

# Build the connector
make build
```

### Basic Usage

#### 1. Create Conduit Pipeline Configuration

```yaml
version: "2.0"
pipelines:
  - id: http-sink-pipeline
    status: running
    connectors:
      - id: source
        type: source
        plugin: builtin:kafka
        settings:
          servers: "localhost:9092"
          topics: "input-topic"

      - id: http-sink
        type: destination
        plugin: builtin:http
        settings:
          url: "https://api.example.com/webhook"
          authType: "bearer"
          bearerToken: "${BEARER_TOKEN}"
          maxRetries: "3"
```

#### 2. Set Environment Variables

```bash
# For Bearer Token authentication
export BEARER_TOKEN="your-api-token"

# For Basic Authentication
export BASIC_USERNAME="user"
export BASIC_PASSWORD="password"

# For OAuth2
export OAUTH2_CLIENT_ID="your-client-id"
export OAUTH2_CLIENT_SECRET="your-client-secret"

# Custom headers (prefix with HTTP_HEADER_)
export HTTP_HEADER_X_API_KEY="secret-key"
export HTTP_HEADER_X_TENANT_ID="tenant-123"
```

## Configuration

### Core HTTP Settings

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `url` | string | *required* | HTTP endpoint URL |
| `method` | string | `POST` | HTTP method (POST, PUT, PATCH) |
| `timeout` | duration | `30s` | Request timeout |
| `maxIdleConns` | int | `100` | Max idle connections |
| `maxConnsPerHost` | int | `10` | Max connections per host |

### Authentication

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `authType` | string | `none` | Authentication type: none, basic, bearer, oauth2 |
| `basicUsername` | string | | Basic auth username (from env) |
| `basicPassword` | string | | Basic auth password (from env) |
| `bearerToken` | string | | Bearer token (from env) |
| `oauth2ClientId` | string | | OAuth2 client ID (from env) |
| `oauth2ClientSecret` | string | | OAuth2 client secret (from env) |
| `oauth2TokenUrl` | string | | OAuth2 token endpoint URL |
| `oauth2Scopes` | string | | OAuth2 scopes (comma-separated) |

### Retry Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `maxRetries` | int | `3` | Max retry attempts (0-10) |
| `retryBackoffBase` | duration | `1s` | Base backoff duration |
| `retryBackoffMax` | duration | `30s` | Max backoff duration |
| `retryOn5xx` | bool | `true` | Retry on 5xx errors |
| `retryOn429` | bool | `true` | Retry on 429 Too Many Requests |
| `retryOnNetworkErr` | bool | `true` | Retry on network errors |

### Response Handling

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `responseOutputPath` | string | `/tmp/conduit-http-responses` | Path for response files |
| `successFile` | string | `success.ndjson` | Success response filename |
| `errorFile` | string | `error.ndjson` | Error response filename |
| `includeResponseHeaders` | bool | `true` | Include HTTP response headers |
| `includeRequestMetadata` | bool | `true` | Include original request metadata |

## Authentication Examples

### Basic Authentication

```yaml
settings:
  url: "https://api.example.com/data"
  authType: "basic"
  basicUsername: "${BASIC_USERNAME}"
  basicPassword: "${BASIC_PASSWORD}"
```

### Bearer Token

```yaml
settings:
  url: "https://api.example.com/data"
  authType: "bearer"
  bearerToken: "${BEARER_TOKEN}"
```

### OAuth2 Client Credentials

```yaml
settings:
  url: "https://api.example.com/data"
  authType: "oauth2"
  oauth2ClientId: "${OAUTH2_CLIENT_ID}"
  oauth2ClientSecret: "${OAUTH2_CLIENT_SECRET}"
  oauth2TokenUrl: "https://auth.example.com/oauth/token"
  oauth2Scopes: "read,write"
```

## Custom Headers

Headers can be set via environment variables using the `HTTP_HEADER_` prefix:

```bash
# These environment variables
export HTTP_HEADER_X_API_KEY="secret123"
export HTTP_HEADER_X_TENANT_ID="tenant-456"
export HTTP_HEADER_USER_AGENT="MyApp/1.0"

# Become HTTP headers
X-Api-Key: secret123
X-Tenant-Id: tenant-456
User-Agent: MyApp/1.0
```

## Response Routing

HTTP responses are written to NDJSON files for further processing:

### Success File Format (`success.ndjson`)

```json
{
  "http_status": 200,
  "response_body": "{\"id\":\"123\",\"status\":\"created\"}",
  "response_headers": {
    "Content-Type": ["application/json"],
    "X-Request-Id": ["abc-123"]
  },
  "correlation_id": "record-uuid",
  "timestamp": "2025-01-01T12:00:00Z",
  "original_record": {
    "position": "...",
    "payload_after": "..."
  }
}
```

### Error File Format (`error.ndjson`)

```json
{
  "http_status": 500,
  "response_body": "{\"error\":\"Internal Server Error\"}",
  "response_headers": {...},
  "error_message": "HTTP 500",
  "correlation_id": "record-uuid",
  "timestamp": "2025-01-01T12:00:00Z",
  "original_record": {...}
}
```

### Setting Up Response Pipeline

To process HTTP responses back to Kafka, create a second pipeline:

```yaml
pipelines:
  - id: http-response-pipeline
    status: running
    connectors:
      - id: response-source
        type: source
        plugin: builtin:file
        settings:
          path: "/tmp/conduit-http-responses/success.ndjson"

      - id: kafka-sink
        type: destination
        plugin: builtin:kafka
        settings:
          servers: "localhost:9092"
          topic: "http-responses"
```

## Development

### Project Structure

```
conduit-connector-http/
├── cmd/connector/         # Main entry point
├── destination/           # Destination implementation
├── internal/
│   ├── auth/             # Authentication (Basic, Bearer, OAuth2)
│   ├── http/             # HTTP client and retry logic
│   └── response/         # Response file writer
├── connector.go          # Connector registration
├── connector.yaml        # Connector metadata
├── Makefile              # Build automation
└── README.md             # This file
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install dependencies
make install

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Run all checks
make check
```

### Testing

```bash
# Run all tests
make test

# Run specific test
go test -v ./destination -run TestDestination_Configure

# Run with coverage
go test -cover ./...
```

## Retry Behavior

The connector implements intelligent retry logic:

### Retryable Errors
- **5xx errors**: Server errors (if `retryOn5xx=true`)
- **429 Too Many Requests**: Rate limiting (if `retryOn429=true`)
- **Network errors**: Connection timeouts, DNS failures (if `retryOnNetworkErr=true`)

### Non-Retryable Errors
- **4xx errors** (except 429): Client errors (bad request, unauthorized, etc.)
- **Authentication failures**: Invalid credentials

### Exponential Backoff
- Formula: `2^attempt × retryBackoffBase`
- Capped at `retryBackoffMax`
- Example with defaults: 1s, 2s, 4s, 8s (max 3 retries)

## Monitoring

### Logging

The connector uses structured logging with the following levels:

- **Debug**: Request/response bodies, detailed flow
- **Info**: Successful operations, configuration
- **Warn**: Validation failures, retries
- **Error**: Failed requests, authentication errors

### Metrics to Track

- HTTP request success/failure rates
- Response time percentiles (p50, p95, p99)
- Retry counts by reason
- OAuth2 token refresh frequency
- Response file sizes

## Limitations

1. **Response Latency**: File-based response handling adds latency (typically 100ms-1s)
2. **OAuth2 Flows**: Only supports Client Credentials; Authorization Code, PKCE not supported
3. **Request Transformation**: Uses raw payload; complex transformations require processors
4. **Concurrent Writes**: Response writer uses mutex; may have contention at very high throughput

## Roadmap

- [ ] Schema validation (JSON Schema, Avro)
- [ ] Request body templates
- [ ] More OAuth2 flows (Authorization Code, PKCE)
- [ ] Response streaming for large payloads
- [ ] Metrics exporter (Prometheus)
- [ ] Health check endpoint

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `make check`
5. Submit a pull request

## License

Apache 2.0 License - see LICENSE file for details.

## Support

- GitHub Issues: [https://github.com/conduitio/conduit-connector-http/issues](https://github.com/conduitio/conduit-connector-http/issues)
- Conduit Docs: [https://conduit.io/docs](https://conduit.io/docs)
- Connector SDK: [https://github.com/ConduitIO/conduit-connector-sdk](https://github.com/ConduitIO/conduit-connector-sdk)

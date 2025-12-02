# Conduit HTTP Destination Connector

A production-ready HTTP destination connector for [Conduit](https://conduit.io) that sends records to HTTP endpoints and publishes responses to Kafka for event-driven architectures.

## Features

- **HTTP Methods**: POST, PUT, PATCH support
- **Authentication**:
  - None
  - Basic Authentication
  - Bearer Token
  - OAuth2 Client Credentials (automatic token management)
- **Custom Headers**: From environment variables and static configuration
- **Retry Logic**: Configurable exponential backoff with intelligent error detection
- **Connection Pooling**: Efficient HTTP client with configurable connection limits
- **Kafka Response Publishing**: Native Kafka producer with SASL/TLS support for response streaming
- **Thread-Safe**: Safe for concurrent use
- **SDK Compliant**: 100% compliant with Conduit Connector SDK v0.14.1

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/conduitio/conduit-connector-http.git
cd conduit-connector-http

# Build the connector
go build -o conduit-connector-http
```

### Basic Usage

#### 1. Create Conduit Pipeline Configuration

```yaml
version: "2.2"
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
          maxRetries: 3
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
| `maxIdleConns` | int | `100` | Max idle connections in pool |
| `maxConnsPerHost` | int | `10` | Max connections per host |

### Authentication

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `authType` | string | `none` | Authentication type: `none`, `basic`, `bearer`, `oauth2` |
| `basicUsername` | string | | Basic auth username (from environment) |
| `basicPassword` | string | | Basic auth password (from environment) |
| `bearerToken` | string | | Bearer token (from environment) |
| `oauth2ClientId` | string | | OAuth2 client ID (from environment) |
| `oauth2ClientSecret` | string | | OAuth2 client secret (from environment) |
| `oauth2TokenUrl` | string | | OAuth2 token endpoint URL |
| `oauth2Scopes` | string | | OAuth2 scopes (comma-separated) |

### Custom Headers

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `staticHeaders` | map | | Static headers to include in all requests |
| `envHeaderPrefix` | string | `HTTP_HEADER_` | Prefix for loading headers from environment |

### Retry Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `maxRetries` | int | `3` | Max retry attempts (0-10) |
| `retryBackoffBase` | duration | `1s` | Base backoff duration |
| `retryBackoffMax` | duration | `30s` | Max backoff duration (cap) |
| `retryOn5xx` | bool | `true` | Retry on 5xx server errors |
| `retryOn429` | bool | `true` | Retry on 429 Too Many Requests |
| `retryOnNetworkErr` | bool | `true` | Retry on network/timeout errors |

### Payload Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `usePayloadAfter` | bool | `true` | Use `Payload.After` field for request body |

### Kafka Response Publishing

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `kafkaEnabled` | bool | `false` | Enable Kafka producer for HTTP responses |
| `kafkaBrokers` | string | | Comma-separated list of Kafka brokers (e.g., "localhost:9092,broker2:9092") |
| `kafkaTopic` | string | `http-responses` | Kafka topic to publish responses to |
| `kafkaClientId` | string | `http-connector` | Kafka client ID |
| `kafkaCompression` | string | `snappy` | Compression: `none`, `gzip`, `snappy`, `lz4`, `zstd` |
| `kafkaEnableIdempotence` | bool | `true` | Enable idempotent producer for exactly-once delivery |
| `kafkaSaslEnabled` | bool | `false` | Enable SASL authentication |
| `kafkaSaslMechanism` | string | `PLAIN` | SASL mechanism: `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512` |
| `kafkaSaslUsername` | string | | SASL username (from environment) |
| `kafkaSaslPassword` | string | | SASL password (from environment) |
| `kafkaTlsEnabled` | bool | `false` | Enable TLS for Kafka connections |

## Authentication Examples

### No Authentication

```yaml
settings:
  url: "https://webhook.site/unique-url"
  authType: "none"
```

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

The connector automatically manages OAuth2 tokens:
- Requests new token on first use
- Caches token in memory
- Automatically renews when expired
- Thread-safe token access

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

### Static Headers

Configure headers directly in the pipeline:

```yaml
settings:
  url: "https://api.example.com/data"
  staticHeaders:
    Content-Type: "application/json"
    X-Client-Id: "my-app"
    X-API-Version: "v1"
```

### Environment Variable Headers

Load headers from environment variables using the `HTTP_HEADER_` prefix:

```bash
# Set environment variables
export HTTP_HEADER_X_API_KEY="secret123"
export HTTP_HEADER_X_TENANT_ID="tenant-456"
export HTTP_HEADER_USER_AGENT="MyApp/1.0"
```

These become HTTP headers:
```
X-Api-Key: secret123
X-Tenant-Id: tenant-456
User-Agent: MyApp/1.0
```

**Note**: Underscores in environment variable names are converted to hyphens in HTTP headers.

## Kafka Response Publishing

The connector can publish HTTP responses to Kafka for downstream processing, event streaming, or analytics.

### How It Works

1. **Destination sends HTTP request** to the configured endpoint
2. **Captures response** (status code, headers, body)
3. **Publishes to Kafka** topic as a JSON message
4. **Downstream consumers** can process responses asynchronously

### Response Message Format

Each HTTP response is published as a JSON message with accompanying Kafka headers:

**JSON Message Body:**
```json
{
  "status_code": 200,
  "response_headers": {
    "Content-Type": "application/json",
    "X-Request-Id": "abc123"
  },
  "body": "{\"success\":true,\"id\":\"12345\"}",
  "request_url": "https://api.example.com/webhook",
  "request_method": "POST",
  "timestamp": "2025-12-02T10:30:00Z"
}
```

**Kafka Record Headers:**
Original OpenCDC record metadata is published as native Kafka record headers:
```
conduit.source.connector.id: kafka-source
conduit.source.plugin.name: builtin:kafka
kafka.topic: api-requests
kafka.partition: 0
kafka.offset: 12345
```

**Field Descriptions:**
- `status_code`: HTTP response status code (e.g., 200, 404, 500)
- `response_headers`: HTTP response headers from the API
- `body`: HTTP response body as a string
- `request_url`: The URL that was called
- `request_method`: HTTP method used (POST, PUT, PATCH)
- `timestamp`: When the response was captured

**Why Separate Headers?**
Record headers are stored as Kafka record headers (not in JSON) for:
- **Efficient filtering**: Consumers can filter by headers without parsing JSON
- **Routing**: Kafka Connect and Stream processors can route based on headers
- **Performance**: Headers are indexed and accessible without deserialization
- **Clean separation**: HTTP response data in body, metadata in headers

### Kafka Configuration Examples

#### Basic Kafka (No Authentication)

```yaml
settings:
  url: "https://api.example.com/data"
  kafkaEnabled: true
  kafkaBrokers: "localhost:9092"
  kafkaTopic: "http-responses"
```

#### Kafka with SASL/PLAIN Authentication

```yaml
settings:
  url: "https://api.example.com/data"
  kafkaEnabled: true
  kafkaBrokers: "broker1:9092,broker2:9092,broker3:9092"
  kafkaTopic: "http-responses"
  kafkaSaslEnabled: true
  kafkaSaslMechanism: "PLAIN"
  kafkaSaslUsername: "${KAFKA_USERNAME}"
  kafkaSaslPassword: "${KAFKA_PASSWORD}"
```

#### Kafka with SASL/SCRAM-SHA-256 and TLS

```yaml
settings:
  url: "https://api.example.com/data"
  kafkaEnabled: true
  kafkaBrokers: "secure-broker1:9093,secure-broker2:9093"
  kafkaTopic: "http-responses-prod"
  kafkaClientId: "http-connector-prod"
  kafkaCompression: "zstd"
  kafkaEnableIdempotence: true
  kafkaSaslEnabled: true
  kafkaSaslMechanism: "SCRAM-SHA-256"
  kafkaSaslUsername: "${KAFKA_USERNAME}"
  kafkaSaslPassword: "${KAFKA_PASSWORD}"
  kafkaTlsEnabled: true
```

#### Environment Variables for Kafka

```bash
# Kafka authentication
export KAFKA_USERNAME="your-kafka-username"
export KAFKA_PASSWORD="your-kafka-password"
```

### Use Cases

**API Response Monitoring**
- Track API response times and status codes
- Alert on errors or degraded performance
- Analyze API usage patterns

**Event-Driven Architectures**
- Trigger downstream workflows based on HTTP responses
- Build event sourcing systems
- Create audit trails

**Data Analytics**
- Stream responses to data lake/warehouse
- Real-time analytics on API interactions
- ML feature engineering from API data

**Webhook Relay**
- Receive webhooks via HTTP
- Publish to Kafka for reliable processing
- Decouple webhook ingestion from processing

### Kafka Performance Tuning

**High Throughput Configuration**:
```yaml
kafkaCompression: "zstd"           # Best compression ratio
kafkaEnableIdempotence: true       # Exactly-once delivery
```

**Low Latency Configuration**:
```yaml
kafkaCompression: "none"           # Skip compression overhead
kafkaEnableIdempotence: false      # Skip idempotence checks
```

**Balanced Configuration** (Recommended):
```yaml
kafkaCompression: "snappy"         # Good balance of speed/compression
kafkaEnableIdempotence: true       # Reliability
```

## Error Handling

### Success Criteria

A request is considered successful if:
- HTTP response status is 2xx (200-299)
- No network errors occurred

### Failure Handling

On failure, the connector:
1. Returns error to Conduit with the record index
2. Conduit will retry the failed record according to pipeline retry policy
3. Pipeline may route to DLQ (Dead Letter Queue) after max retries

### Retryable vs Non-Retryable Errors

**Retryable** (will be retried automatically):
- 5xx server errors (if `retryOn5xx=true`)
- 429 Too Many Requests (if `retryOn429=true`)
- Network timeouts (if `retryOnNetworkErr=true`)
- Connection failures (if `retryOnNetworkErr=true`)

**Non-Retryable** (fail immediately):
- 4xx client errors (except 429)
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- Invalid payload

## Retry Behavior

The connector implements exponential backoff:

### Backoff Calculation

```
backoff = 2^attempt × retryBackoffBase
backoff = min(backoff, retryBackoffMax)
```

### Example with Defaults

With `retryBackoffBase=1s`, `retryBackoffMax=30s`, `maxRetries=3`:

| Attempt | Backoff | Total Time |
|---------|---------|------------|
| 1       | 2s      | 2s         |
| 2       | 4s      | 6s         |
| 3       | 8s      | 14s        |

### Retry Configuration Examples

**Fast Retry** (for real-time APIs):
```yaml
maxRetries: 2
retryBackoffBase: 500ms
retryBackoffMax: 5s
```

**Conservative Retry** (for rate-limited APIs):
```yaml
maxRetries: 5
retryBackoffBase: 2s
retryBackoffMax: 60s
retryOn429: true
```

**No Retry** (fail fast):
```yaml
maxRetries: 0
```

## Connection Pooling

The connector maintains an HTTP connection pool for efficiency:

```yaml
settings:
  maxIdleConns: 100        # Total idle connections across all hosts
  maxConnsPerHost: 10      # Max connections per target host
  timeout: 30s             # Per-request timeout
```

**Recommendations**:
- **Low throughput** (<10 req/s): Use defaults
- **Medium throughput** (10-100 req/s): Increase `maxConnsPerHost` to 20-50
- **High throughput** (>100 req/s): Increase both to 100-200

## Development

### Project Structure

```
conduit-connector-http/
├── cmd/connector/         # Main entry point
├── destination/           # Destination implementation
│   ├── config.go         # Configuration structure
│   └── destination.go    # Core logic
├── internal/
│   ├── auth/             # Authentication (Basic, Bearer, OAuth2)
│   ├── http/             # HTTP client and retry logic
│   └── schema/           # Schema validation (future)
├── examples/             # Example configurations
├── connector.go          # Connector registration
├── go.mod                # Go module definition
├── Makefile              # Build automation
└── README.md             # This file
```

### Building

```bash
# Build for current platform
go build -o conduit-connector-http

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o conduit-connector-http-linux-amd64

# Install dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./destination -run TestOpen
```

## Examples

See the `examples/` directory for complete pipeline configurations:

- `01-basic-no-auth.yaml` - Simple webhook without authentication
- `02-basic-auth.yaml` - Basic authentication example
- `03-bearer-token.yaml` - Bearer token authentication
- `04-oauth2-client-credentials.yaml` - OAuth2 with automatic token management
- `05-custom-headers.yaml` - Environment variable headers
- `06-full-featured.yaml` - All configuration options

## Logging

The connector uses structured logging (zerolog) with these levels:

- **Info**: Configuration, successful operations
- **Debug**: Request details, retry attempts
- **Warn**: Non-2xx responses, retry exhausted
- **Error**: Authentication failures, network errors

Example log output:
```
INFO Opening HTTP destination url=https://api.example.com authType=bearer
DEBUG HTTP request successful status=200
WARN HTTP request returned non-2xx status status=429
ERROR HTTP request failed after retries error="connection timeout"
```

## Performance Considerations

### Throughput

Expected throughput depends on:
- Target API response time (e.g., 100ms = ~10 req/s per connection)
- Connection pool size (`maxConnsPerHost`)
- Network latency

**Example**: With 10ms API latency and 10 connections → ~1000 req/s theoretical max

### Latency

Request latency includes:
- Network RTT
- Authentication overhead (OAuth2 token request: ~100ms on first request)
- Target API processing time
- Retry delays (if applicable)

### Memory Usage

- Base: ~10MB
- Per connection: ~4KB
- OAuth2 token cache: ~1KB

## Limitations

1. **HTTP Methods**: Only POST, PUT, PATCH (no GET, DELETE)
2. **OAuth2 Flows**: Only Client Credentials (no Authorization Code, PKCE)
3. **Request Transformation**: Uses raw payload (no templating yet)
4. **Response Publishing**: Requires Kafka for response capture (no in-memory queue)
5. **Schema Validation**: Not yet implemented

## Roadmap

Potential future enhancements:

- [x] Kafka response publishing with SASL/TLS
- [ ] Schema validation (JSON Schema, Avro)
- [ ] URL templating (dynamic endpoints)
- [ ] Request body templates
- [ ] DELETE method support
- [ ] Request/response transformation
- [ ] More OAuth2 flows (Authorization Code, PKCE)
- [ ] Metrics exporter (Prometheus)
- [ ] Health check endpoint
- [ ] Batch request support

## Troubleshooting

### Authentication Failures

**Problem**: `401 Unauthorized` or `403 Forbidden`

**Solutions**:
- Verify credentials are set in environment variables
- Check OAuth2 token endpoint URL is correct
- Ensure OAuth2 scopes are appropriate
- Test authentication manually with `curl`

### Connection Timeouts

**Problem**: `context deadline exceeded` or network timeouts

**Solutions**:
- Increase `timeout` (e.g., `60s`)
- Check target API is reachable
- Verify network connectivity
- Increase `maxRetries` for transient issues

### Too Many Retries

**Problem**: Requests retry too aggressively

**Solutions**:
- Reduce `maxRetries`
- Increase `retryBackoffBase`
- Disable retry for specific errors:
  ```yaml
  retryOn5xx: false
  retryOn429: false
  ```

### OAuth2 Token Issues

**Problem**: Token not refreshing or authentication loops

**Solutions**:
- Verify OAuth2 credentials and token URL
- Check OAuth2 server is returning valid tokens
- Review connector logs for token request errors

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `go test ./...` and `go build`
5. Submit a pull request

## License

Apache 2.0 License - see LICENSE file for details.

## Support

- **GitHub Issues**: [https://github.com/conduitio/conduit-connector-http/issues](https://github.com/conduitio/conduit-connector-http/issues)
- **Conduit Docs**: [https://conduit.io/docs](https://conduit.io/docs)
- **Connector SDK**: [https://github.com/ConduitIO/conduit-connector-sdk](https://github.com/ConduitIO/conduit-connector-sdk)
- **Examples**: See `examples/` directory in this repository

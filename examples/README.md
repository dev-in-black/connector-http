# Conduit HTTP Sink Connector - Configuration Examples

This directory contains example Conduit pipeline configurations demonstrating various use cases of the HTTP sink connector.

## Quick Start

1. **Choose an example** that matches your use case
2. **Set required environment variables** (documented in each file)
3. **Run Conduit** with the configuration:
   ```bash
   conduit run -c <example-file>.yaml
   ```

## Examples Overview

### [01-basic-no-auth.yaml](01-basic-no-auth.yaml)
**Simplest configuration** - Send records to HTTP endpoint without authentication.

**Use case**: Testing, webhooks that don't require authentication

**Key features**:
- No authentication
- Basic error handling
- Response file writing

### [02-basic-auth.yaml](02-basic-auth.yaml)
**Basic Authentication** - Username and password authentication.

**Use case**: Internal APIs, legacy systems

**Environment variables required**:
```bash
export BASIC_AUTH_USERNAME="your-username"
export BASIC_AUTH_PASSWORD="your-password"
```

**Key features**:
- HTTP Basic Authentication
- Configurable retry logic
- Error handling with exponential backoff

### [03-bearer-token.yaml](03-bearer-token.yaml)
**Bearer Token Authentication** - Static API token authentication.

**Use case**: Most modern REST APIs

**Environment variables required**:
```bash
export BEARER_TOKEN="your-bearer-token"
```

**Key features**:
- Bearer Token (API Key) authentication
- Connection pooling configuration
- Full response metadata capture

### [04-oauth2-client-credentials.yaml](04-oauth2-client-credentials.yaml)
**OAuth2 Client Credentials** - Automatic token management.

**Use case**: Enterprise APIs, secured services

**Environment variables required**:
```bash
export OAUTH2_CLIENT_ID="your-client-id"
export OAUTH2_CLIENT_SECRET="your-client-secret"
```

**Key features**:
- OAuth2 Client Credentials flow
- Automatic token acquisition and caching
- Token expiration handling
- Thread-safe token management
- Production-grade retry configuration

**How it works**:
1. Connector requests token from OAuth2 server on first use
2. Token is cached in memory
3. Token is automatically renewed when expired
4. All requests use the cached token

### [05-custom-headers.yaml](05-custom-headers.yaml)
**Custom HTTP Headers** - Add headers from environment variables.

**Use case**: APIs requiring custom headers (API keys, tenant IDs, etc.)

**Environment variables for headers**:
```bash
export HTTP_HEADER_X_API_KEY="secret-key"
export HTTP_HEADER_X_TENANT_ID="tenant-123"
export HTTP_HEADER_USER_AGENT="MyApp/1.0"
```

**Key features**:
- Dynamic header loading from environment
- Automatic header name conversion (underscore → hyphen)
- Combined with authentication

**Header conversion**:
- `HTTP_HEADER_X_API_KEY` → `X-Api-Key` header
- `HTTP_HEADER_USER_AGENT` → `User-Agent` header

### [06-full-featured.yaml](06-full-featured.yaml)
**Complete Configuration** - All available options demonstrated.

**Use case**: Production deployments, reference configuration

**Key features**:
- OAuth2 authentication
- Custom headers
- Advanced retry configuration
- Connection pooling tuning
- Complete response handling
- All configurable parameters

### [07-response-pipeline.yaml](07-response-pipeline.yaml)
**Request-Response Pipeline** - Complete bidirectional flow.

**Use case**: When you need to process HTTP responses in Kafka

**Key features**:
- Three pipelines:
  1. **Request Pipeline**: Kafka → HTTP → Response Files
  2. **Success Pipeline**: Success Files → Kafka Success Topic
  3. **Error Pipeline**: Error Files → Kafka Error Topic
- Complete response routing
- Separate success and error handling

**Data Flow**:
```
Kafka (requests)
    ↓
HTTP Endpoint
    ↓
Response Files (success.ndjson / error.ndjson)
    ↓
Kafka (responses-success / responses-errors)
```

## Configuration Parameters Reference

### Core HTTP Settings

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `url` | string | *required* | HTTP endpoint URL |
| `method` | string | `POST` | HTTP method (POST, PUT, PATCH) |
| `timeout` | duration | `30s` | Request timeout |
| `maxIdleConns` | int | `100` | Maximum idle connections |
| `maxConnsPerHost` | int | `10` | Maximum connections per host |

### Authentication

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `authType` | string | `none` | Auth type: none, basic, bearer, oauth2 |
| `basicUsername` | string | | Basic auth username |
| `basicPassword` | string | | Basic auth password |
| `bearerToken` | string | | Bearer token |
| `oauth2ClientId` | string | | OAuth2 client ID |
| `oauth2ClientSecret` | string | | OAuth2 client secret |
| `oauth2TokenUrl` | string | | OAuth2 token endpoint |
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
| `responseOutputPath` | string | `/tmp/conduit-http-responses` | Response files directory |
| `successFile` | string | `success.ndjson` | Success responses filename |
| `errorFile` | string | `error.ndjson` | Error responses filename |
| `includeResponseHeaders` | bool | `true` | Include HTTP response headers |
| `includeRequestMetadata` | bool | `true` | Include original record metadata |

## Common Use Cases

### Use Case 1: Webhook Delivery

Send Kafka events to external webhooks:

```yaml
url: "https://webhook.site/unique-url"
authType: "none"
maxRetries: "3"
```

**Files**: 01-basic-no-auth.yaml

### Use Case 2: API Integration

Integrate with REST APIs:

```yaml
url: "https://api.example.com/v1/events"
authType: "bearer"
bearerToken: "${API_TOKEN}"
maxRetries: "5"
```

**Files**: 03-bearer-token.yaml

### Use Case 3: Enterprise Integration

Connect to OAuth2-secured enterprise APIs:

```yaml
url: "https://enterprise.example.com/api"
authType: "oauth2"
oauth2ClientId: "${CLIENT_ID}"
oauth2ClientSecret: "${CLIENT_SECRET}"
oauth2TokenUrl: "https://auth.example.com/token"
```

**Files**: 04-oauth2-client-credentials.yaml

### Use Case 4: Multi-Tenant SaaS

Send events with tenant-specific headers:

```yaml
authType: "bearer"
bearerToken: "${TENANT_API_TOKEN}"
# Environment: HTTP_HEADER_X_TENANT_ID=tenant-123
```

**Files**: 05-custom-headers.yaml

### Use Case 5: Event Processing Pipeline

Full request-response cycle with Kafka:

- Send events to HTTP endpoints
- Capture all responses
- Route successes and failures to different Kafka topics
- Enable downstream processing

**Files**: 07-response-pipeline.yaml

## Retry Behavior

The connector implements intelligent retry logic:

### Retryable Errors
- **5xx errors**: Server errors (if `retryOn5xx=true`)
- **429 Too Many Requests**: Rate limiting (if `retryOn429=true`)
- **Network errors**: Timeouts, connection failures (if `retryOnNetworkErr=true`)

### Non-Retryable Errors
- **4xx errors** (except 429): Client errors
- **Authentication failures**: Invalid credentials

### Backoff Calculation
- Formula: `2^attempt × retryBackoffBase`
- Capped at `retryBackoffMax`
- Example with defaults (base=1s, max=30s):
  - Attempt 1: 2s wait
  - Attempt 2: 4s wait
  - Attempt 3: 8s wait
  - Attempt 4+: 30s wait (capped)

## Response File Format

### Success File (success.ndjson)

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

### Error File (error.ndjson)

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

## Testing Tips

### 1. Test with webhook.site

Use [webhook.site](https://webhook.site) for quick testing:

```yaml
url: "https://webhook.site/your-unique-url"
authType: "none"
```

View all received requests in the webhook.site UI.

### 2. Test with httpbin

Use [httpbin.org](https://httpbin.org) for testing different scenarios:

```yaml
# Test successful response
url: "https://httpbin.org/post"

# Test 500 error
url: "https://httpbin.org/status/500"

# Test timeout
url: "https://httpbin.org/delay/10"
```

### 3. Monitor Response Files

Watch response files in real-time:

```bash
# Watch success responses
tail -f /tmp/conduit-http-responses/success.ndjson | jq .

# Watch error responses
tail -f /tmp/conduit-http-responses/error.ndjson | jq .
```

### 4. Test Authentication

Verify authentication headers:

```yaml
url: "https://httpbin.org/headers"
authType: "bearer"
bearerToken: "test-token"
```

Check the response to see if the Authorization header is present.

## Troubleshooting

### Issue: Authentication failures

**Solution**: Verify environment variables are set:
```bash
echo $BEARER_TOKEN
echo $OAUTH2_CLIENT_ID
```

### Issue: Connection timeouts

**Solution**: Increase timeout:
```yaml
timeout: "60s"
maxRetries: "5"
retryBackoffMax: "60s"
```

### Issue: Too many retries

**Solution**: Reduce retry count for non-retryable errors:
```yaml
maxRetries: "1"
retryOn5xx: "false"
```

### Issue: Response files growing too large

**Solution**: Set up file rotation or use the response pipeline (example 07) to consume responses immediately.

## Next Steps

1. **Choose an example** matching your use case
2. **Copy and customize** the configuration
3. **Set environment variables**
4. **Test with a safe endpoint** (webhook.site)
5. **Monitor response files**
6. **Deploy to production**

## Additional Resources

- [Conduit Documentation](https://conduit.io/docs)
- [Connector SDK](https://github.com/ConduitIO/conduit-connector-sdk)
- [Full README](../README.md)

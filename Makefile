.PHONY: build test install clean lint

# Build the connector
build:
	@echo "Building HTTP connector..."
	go build -o conduit-connector-http cmd/connector/main.go

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f conduit-connector-http
	rm -rf /tmp/conduit-http-responses

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run all checks
check: fmt lint test

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o bin/conduit-connector-http-linux-amd64 cmd/connector/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/conduit-connector-http-darwin-amd64 cmd/connector/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/conduit-connector-http-darwin-arm64 cmd/connector/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/conduit-connector-http-windows-amd64.exe cmd/connector/main.go

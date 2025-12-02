package http

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/dev-in-black/connector-http/destination"
)

// Connector is the main entry point for the HTTP connector
var Connector = sdk.Connector{
	NewSpecification: Specification,
	NewSource:        nil, // HTTP destination only - responses published to Kafka
	NewDestination:   destination.NewDestination,
}

// Specification returns the connector specification
func Specification() sdk.Specification {
	return sdk.Specification{
		Name:    "http",
		Summary: "HTTP destination connector with Kafka response publishing",
		Description: "The HTTP connector sends records to HTTP endpoints with enterprise-grade features. " +
			"Supports multiple authentication methods (Basic, Bearer, OAuth2), custom headers, " +
			"configurable retry logic, and connection pooling. " +
			"HTTP responses can be published to Kafka topics for downstream processing, " +
			"enabling event-driven architectures and response analytics.",
		Version: "v0.3.0",
		Author:  "Conduit",
	}
}

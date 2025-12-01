package http

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-http/destination"
)

// Connector is the main entry point for the HTTP connector
var Connector = sdk.Connector{
	NewSpecification: Specification,
	NewSource:        nil, // HTTP sink only, no source
	NewDestination:   destination.NewDestination,
}

// Specification returns the connector specification
func Specification() sdk.Specification {
	return sdk.Specification{
		Name:    "http",
		Summary: "HTTP sink connector for Conduit",
		Description: "The HTTP connector allows you to send records from Conduit to HTTP endpoints. " +
			"It supports multiple authentication methods (Basic, Bearer, OAuth2), custom headers, " +
			"configurable retry logic, and writes responses to files for further processing.",
		Version: "v0.1.0",
		Author:  "Conduit",
	}
}

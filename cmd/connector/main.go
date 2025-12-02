package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
	http "github.com/dev-in-black/connector-http"
)

func main() {
	sdk.Serve(http.Connector)
}

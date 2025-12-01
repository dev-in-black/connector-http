package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
	http "github.com/conduitio/conduit-connector-http"
)

func main() {
	sdk.Serve(http.Connector)
}

package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"

	rabbitmq "github.com/alarbada/conduit-connector-rabbitmq"
)

func main() {
	sdk.Serve(rabbitmq.Connector)
}

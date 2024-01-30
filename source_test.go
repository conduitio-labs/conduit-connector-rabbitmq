package rabbitmq_test

import (
	"context"
	"testing"

	rabbitmq "github.com/alarbada/conduit-connector-rabbitmq"
	"github.com/matryer/is"
)

func TestTeardownSource_NoOpen(t *testing.T) {
	is := is.New(t)
	con := rabbitmq.NewSource()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}

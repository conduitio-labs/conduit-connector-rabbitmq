package rabbitmq_test

import (
	"context"
	"testing"

	rabbitmq "github.com/alarbada/conduit-connector-rabbitmq"
	"github.com/matryer/is"
)

func TestTeardown_NoOpen(t *testing.T) {
	is := is.New(t)
	con := rabbitmq.NewDestination()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}

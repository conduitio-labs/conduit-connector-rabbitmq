package rabbitmq

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestTeardown_NoOpen(t *testing.T) {
	is := is.New(t)
	con := NewDestination()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}

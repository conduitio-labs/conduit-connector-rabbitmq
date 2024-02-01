package rabbitmq

import (
	"context"
	"fmt"
	"testing"

	"github.com/alarbada/conduit-connector-rabbitmq/test"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/rabbitmq/amqp091-go"
)

func TestTeardownDestination_NoOpen(t *testing.T) {
	is := is.New(t)
	con := NewDestination()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}

func newDestinationCfg(queueName string) map[string]string {
	return map[string]string{
		"url":       test.URL,
		"queueName": queueName,
	}
}

func TestDestination_Integration(t *testing.T) {
	// ctx := test.CtxWithLogger()
	ctx := context.Background()
	is := is.New(t)

	queueName := test.SetupQueueName(t, is)

	{
		destination := NewDestination()
		cfg := newDestinationCfg(queueName)

		err := destination.Configure(ctx, cfg)
		is.NoErr(err)

		err = destination.Open(ctx)
		is.NoErr(err)

		defer test.TeardownResource(is, ctx, destination)

		recsToWrite := generate3Records(queueName)
		writtenTotal, err := destination.Write(ctx, recsToWrite)
		is.Equal(writtenTotal, len(recsToWrite))
		is.NoErr(err)
	}

	{
		conn, err := amqp091.Dial(test.URL)
		is.NoErr(err)

		defer test.CloseResource(is, conn)

		ch, err := conn.Channel()
		is.NoErr(err)

		defer test.CloseResource(is, ch)

		recs, err := ch.Consume(queueName, "", true, false, false, false, nil)
		is.NoErr(err)

		rec1 := <-recs
		is.Equal(string(rec1.Body), "example message 0")

		rec2 := <-recs
		is.Equal(string(rec2.Body), "example message 1")

		rec3 := <-recs
		is.Equal(string(rec3.Body), "example message 2")
	}
}

func generate3Records(queueName string) []sdk.Record {
	recs := []sdk.Record{}

	for i := 0; i < 3; i++ {
		exampleMessage := fmt.Sprintf("example message %d", i)

		rec := sdk.Util.Source.NewRecordCreate(
			[]byte(uuid.NewString()),
			sdk.Metadata{"rabbitmq.queue": queueName},
			sdk.RawData("test-key"),
			sdk.RawData(exampleMessage),
		)

		recs = append(recs, rec)
	}

	return recs
}

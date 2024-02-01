package test

import (
	"context"
	"os"
	"testing"

	"github.com/matryer/is"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const URL = "amqp://guest:guest@localhost:5672"

func CtxWithLogger() context.Context {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return logger.WithContext(ctx)
}

// SetupQueueName creates a new topic name for the test and deletes it if it
// exists, so that the test can start from a clean slate.
func SetupQueueName(t *testing.T, is *is.I) string {
	queueName := "rabbitmq.queue." + t.Name()
	DeleteQueue(is, queueName)

	return queueName
}

func DeleteQueue(is *is.I, queueName string) {
	conn, err := amqp091.Dial(URL)
	is.NoErr(err)

	ch, err := conn.Channel()
	is.NoErr(err)

	// force queue delete
	_, err = ch.QueueDelete(queueName, false, false, false)
	is.NoErr(err)
}

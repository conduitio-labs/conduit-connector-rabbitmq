package rabbitmq

//go:generate paramgen -output=paramgen_dest.go DestinationConfig

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/rabbitmq/amqp091-go"
)

type Destination struct {
	sdk.UnimplementedDestination

	conn *amqp091.Connection
	ch   *amqp091.Channel

	config DestinationConfig
}

func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{}, sdk.DefaultDestinationMiddleware()...)
}

func (d *Destination) Parameters() map[string]sdk.Parameter {
	return d.config.Parameters()
}

func (d *Destination) Configure(ctx context.Context, cfg map[string]string) error {
	sdk.Logger(ctx).Info().Msg("Configuring Destination...")
	err := sdk.Util.ParseConfig(cfg, &d.config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

func (d *Destination) Open(ctx context.Context) (err error) {
	d.conn, err = amqp091.Dial(d.config.URL)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	d.ch, err = d.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	return nil
}

func (d *Destination) Write(ctx context.Context, records []sdk.Record) (int, error) {
	for _, record := range records {
		msg := amqp091.Publishing{
			ContentType: "text/plain",
			Body:        record.Payload.After.Bytes(),
		}

		err := d.ch.PublishWithContext(ctx, "", d.config.QueueName, false, false, msg)
		if err != nil {
			return 0, fmt.Errorf("failed to publish: %w", err)
		}
	}

	return len(records), nil
}

func (d *Destination) Teardown(ctx context.Context) error {
	errch := d.ch.Close()
	errconn := d.conn.Close()
	return errors.Join(errch, errconn)
}

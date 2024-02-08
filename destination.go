// Copyright © 2024 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func (d *Destination) Configure(ctx context.Context, cfg map[string]string) (err error) {
	sdk.Logger(ctx).Info().Msg("Configuring Destination...")
	d.config, err = newDestinationConfig(cfg)

	return err
}

func (d *Destination) Open(ctx context.Context) (err error) {
	d.conn, err = amqp091.Dial(d.config.URL)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("connected to RabbitMQ")

	d.ch, err = d.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	sdk.Logger(ctx).Debug().Msgf("opened channel")

	_, err = d.ch.QueueDeclare(d.config.QueueName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	sdk.Logger(ctx).Debug().Msgf("declared queue %s", d.config.QueueName)

	return nil
}

func (d *Destination) Write(ctx context.Context, records []sdk.Record) (int, error) {
	for _, record := range records {
		msgId := string(record.Key.Bytes())
		msg := amqp091.Publishing{
			MessageId:   msgId,
			ContentType: d.config.ContentType,
			Body:        record.Payload.After.Bytes(),
		}

		err := d.ch.PublishWithContext(ctx, "", d.config.QueueName, false, false, msg)
		if err != nil {
			return 0, fmt.Errorf("failed to publish: %w", err)
		}

		sdk.Logger(ctx).Debug().Msgf("published message %s on %s", msgId, d.config.QueueName)
	}

	return len(records), nil
}

func (d *Destination) Teardown(_ context.Context) error {
	errs := make([]error, 0, 2)
	if d.ch != nil {
		if err := d.ch.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	sdk.Logger(context.Background()).Debug().Msg("destination teardown complete")

	return nil
}

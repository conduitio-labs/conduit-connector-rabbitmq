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

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/rabbitmq/amqp091-go"
)

type Destination struct {
	sdk.UnimplementedDestination

	conn *amqp091.Connection
	ch   *amqp091.Channel

	config    DestinationConfig
	tlsConfig *tls.Config

	exchange, routingKey string
}

func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{}, sdk.DefaultDestinationMiddleware()...)
}

func (d *Destination) Parameters() map[string]sdk.Parameter {
	return d.config.Parameters()
}

func (d *Destination) Configure(ctx context.Context, cfg map[string]string) (err error) {
	d.config, err = newDestinationConfig(cfg)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if shouldParseTLSConfig(ctx, d.config.Config) {
		d.tlsConfig, err = parseTLSConfig(ctx, d.config.Config)
		if err != nil {
			return fmt.Errorf("failed to parse TLS config: %w", err)
		}

		sdk.Logger(ctx).Debug().Msg("source configured with TLS")
		return nil
	}

	d.exchange = d.config.ExchangeName
	d.routingKey = d.config.RoutingKey
	if d.exchange == "" {
		d.routingKey = d.config.QueueName
	}

	sdk.Logger(ctx).Debug().Msg("destination configured")
	return nil
}

func (d *Destination) Open(ctx context.Context) (err error) {
	d.conn, err = ampqDial(d.config.URL, d.tlsConfig)
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

	if d.config.ExchangeName != "" {
		err = d.ch.ExchangeDeclare(d.config.ExchangeName, d.config.ExchangeType, false, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
		sdk.Logger(ctx).Debug().Msgf("declared exchange %s", d.config.ExchangeName)

		err = d.ch.QueueBind(d.config.QueueName, d.config.RoutingKey, d.config.ExchangeName, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
		sdk.Logger(ctx).Debug().Msgf("bound queue %s to exchange %s with routing key %s", d.config.QueueName, d.config.ExchangeName, d.config.RoutingKey)
	}

	return nil
}

func (d *Destination) Write(ctx context.Context, records []sdk.Record) (int, error) {
	for _, record := range records {
		msgID := string(record.Position)
		msg := amqp091.Publishing{
			MessageId:   msgID,
			ContentType: d.config.ContentType,
			Body:        record.Bytes(),
		}

		err := d.ch.PublishWithContext(ctx, d.exchange, d.routingKey, false, false, msg)
		if err != nil {
			return 0, fmt.Errorf("failed to publish: %w", err)
		}

		sdk.Logger(ctx).Trace().Str("MessageID", msgID).Str("queue", d.config.QueueName).Msg("published message")
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

	sdk.Logger(context.Background()).Info().Msg("destination teardown complete")

	return nil
}

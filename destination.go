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

	d.exchange = d.config.Exchange.Name
	d.routingKey = d.config.RoutingKey
	if d.exchange == "" {
		d.routingKey = d.config.QueueName
	}

	if shouldParseTLSConfig(ctx, d.config.Config) {
		d.tlsConfig, err = parseTLSConfig(ctx, d.config.Config)
		if err != nil {
			return fmt.Errorf("failed to parse TLS config: %w", err)
		}

		sdk.Logger(ctx).Debug().Msg("source configured with TLS")
		return nil
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

	_, err = d.ch.QueueDeclare(
		d.config.QueueName,
		d.config.Queue.Durable,
		d.config.Queue.AutoDelete,
		d.config.Queue.Exclusive,
		d.config.Queue.NoWait,
		nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	sdk.Logger(ctx).Debug().Msgf("declared queue %s", d.config.QueueName)

	if d.config.Exchange.Name != "" {
		err = d.ch.ExchangeDeclare(
			d.config.Exchange.Name,
			d.config.Exchange.Type,
			d.config.Exchange.Durable,
			d.config.Exchange.AutoDelete,
			d.config.Exchange.Internal,
			d.config.Exchange.NoWait,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
		sdk.Logger(ctx).Debug().Msgf("declared exchange %s", d.config.Exchange.Name)

		err = d.ch.QueueBind(d.config.QueueName, d.config.RoutingKey, d.config.Exchange.Name, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
		sdk.Logger(ctx).Debug().Msgf(
			"bound queue %s to exchange %s with routing key %s",
			d.config.QueueName, d.config.Exchange.Name, d.config.RoutingKey,
		)
	}

	return nil
}

func (d *Destination) Write(ctx context.Context, records []sdk.Record) (int, error) {
	for _, record := range records {
		msgID := string(record.Position)
		msg := amqp091.Publishing{
			ContentType:     d.config.Delivery.ContentType,
			ContentEncoding: d.config.Delivery.ContentEncoding,
			DeliveryMode:    d.config.Delivery.DeliveryMode,
			Priority:        d.config.Delivery.Priority,
			CorrelationId:   d.config.Delivery.CorrelationID,
			ReplyTo:         d.config.Delivery.ReplyTo,

			MessageId: msgID,
			Type:      d.config.Delivery.MessageTypeName,
			UserId:    d.config.Delivery.UserID,
			AppId:     d.config.Delivery.AppID,
			Body:      record.Bytes(),

			// TODO: study this field, I'm not sure what's supposed to do
			Expiration: "",
		}

		if createdAt, err := record.Metadata.GetCreatedAt(); err != nil {
			msg.Timestamp = createdAt
		}

		err := d.ch.PublishWithContext(
			ctx, d.exchange, d.routingKey,
			d.config.Delivery.Mandatory,
			d.config.Delivery.Immediate,
			msg,
		)
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

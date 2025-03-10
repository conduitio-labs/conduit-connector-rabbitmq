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
	"errors"
	"fmt"

	"github.com/conduitio/conduit-commons/opencdc"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/rabbitmq/amqp091-go"
)

type Destination struct {
	sdk.UnimplementedDestination

	config DestinationConfig

	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func (d *Destination) Config() sdk.DestinationConfig {
	return &d.config
}

func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{})
}

func (d *Destination) Open(ctx context.Context) (err error) {
	d.conn, err = ampqDial(ctx, d.config.Config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	sdk.Logger(ctx).Debug().Str("url", d.config.URL).Msg("connected to RabbitMQ")

	d.ch, err = d.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	sdk.Logger(ctx).Debug().Msgf("opened channel")

	_, err = d.ch.QueueDeclare(
		d.config.Queue.Name,
		d.config.Queue.Durable,
		d.config.Queue.AutoDelete,
		d.config.Queue.Exclusive,
		d.config.Queue.NoWait,
		nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	sdk.Logger(ctx).Debug().Any("queueConfig", d.config.Queue).Msgf("declared queue")

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
		sdk.Logger(ctx).Debug().Any("exchange config", d.config.Exchange).Msgf("declared exchange")

		err = d.ch.QueueBind(d.config.Queue.Name, d.config.RoutingKey, d.config.Exchange.Name, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
		sdk.Logger(ctx).Debug().Msgf(
			"bound queue %s to exchange %s with routing key %s",
			d.config.Queue.Name, d.config.Exchange.Name, d.config.RoutingKey,
		)
	}

	return nil
}

func (d *Destination) Write(ctx context.Context, records []opencdc.Record) (int, error) {
	for i, record := range records {
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

			Expiration: d.config.Delivery.Expiration,
		}

		if createdAt, err := record.Metadata.GetCreatedAt(); err != nil {
			msg.Timestamp = createdAt
		}

		err := d.ch.PublishWithContext(
			ctx,
			d.config.Exchange.Name,
			d.config.RoutingKey,
			d.config.Delivery.Mandatory,
			d.config.Delivery.Immediate,
			msg,
		)
		if err != nil {
			return i, fmt.Errorf("failed to publish: %w", err)
		}

		sdk.Logger(ctx).Trace().
			Str("messageID", msgID).
			Str("routingKey", d.config.RoutingKey).
			Bool("mandatoryDelivery", d.config.Delivery.Mandatory).
			Bool("immediateDelivery", d.config.Delivery.Immediate).
			Msg("published message")
	}

	return len(records), nil
}

func (d *Destination) Teardown(ctx context.Context) error {
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

	sdk.Logger(ctx).Info().Msg("destination teardown complete")

	return nil
}

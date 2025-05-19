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

	routingKey    string
	getRoutingKey RoutingKeyFn
}

func (d *Destination) Config() sdk.DestinationConfig {
	return &d.config
}

func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{})
}

func (d *Destination) Open(ctx context.Context) (err error) {
	routingKey, routingKeyFn, err := d.config.ParseRoutingKey()
	if err != nil {
		// Unlikely to happen, as the topic is validated in the config.
		return fmt.Errorf("failed to parse routingKey: %w", err)
	}

	d.routingKey = routingKey
	d.getRoutingKey = routingKeyFn

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

	// We setup the channel in confirm mode, so that we ensure the "written"
	// total of records that we return is correct.
	noWait := false
	if err := d.ch.Confirm(noWait); err != nil {
		return fmt.Errorf("failed to set channel on confirm mode: %w", err)
	}

	err = d.declareQueue(ctx)
	if err != nil {
		return err
	}

	err = d.declareExchange(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (d *Destination) createMessage(record opencdc.Record, msgID string) amqp091.Publishing {
	return amqp091.Publishing{
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
		Headers:   metadataToHeaders(record.Metadata),
		Body:      record.Bytes(),

		Expiration: d.config.Delivery.Expiration,
	}
}

func (d *Destination) createConfirmation(
	ctx context.Context,
	confirmation *amqp091.DeferredConfirmation,
	msgID string,
	routingKey string,
	msg amqp091.Publishing,
) func() error {
	return func() error {
		confirmed, err := confirmation.WaitContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait for publish confirmation: %w", err)
		}

		if !confirmed {
			return fmt.Errorf(
				"message was not confirmed: id=%s type=%s userId=%s appId=%s",
				msgID, msg.Type, msg.UserId, msg.AppId,
			)
		}

		sdk.Logger(ctx).Trace().
			Str("messageID", msgID).
			Str("routingKey", routingKey).
			Bool("mandatoryDelivery", d.config.Delivery.Mandatory).
			Bool("immediateDelivery", d.config.Delivery.Immediate).
			Msg("published message")

		return nil
	}
}

func (d *Destination) Write(ctx context.Context, records []opencdc.Record) (int, error) {
	confirmations := make([]func() error, 0, len(records))

	for _, record := range records {
		msgID := string(record.Position)
		msg := d.createMessage(record, msgID)

		if createdAt, err := record.Metadata.GetCreatedAt(); err == nil {
			msg.Timestamp = createdAt
		}

		routingKey := d.routingKey
		if d.getRoutingKey != nil {
			var err error
			routingKey, err = d.getRoutingKey(record)
			if err != nil {
				return 0, fmt.Errorf("could not get routingKey: %w", err)
			}
		}

		confirmation, err := d.ch.PublishWithDeferredConfirm(
			d.config.Exchange.Name,
			routingKey,
			d.config.Delivery.Mandatory,
			d.config.Delivery.Immediate,
			msg,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to publish: %w", err)
		}

		confirmationFn := d.createConfirmation(ctx, confirmation, msgID, routingKey, msg)
		confirmations = append(confirmations, confirmationFn)
	}

	for i, confirmation := range confirmations {
		if err := confirmation(); err != nil {
			return i, err
		}
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

func (d *Destination) declareQueue(ctx context.Context) (err error) {
	if d.config.Queue.SkipDeclare {
		sdk.Logger(ctx).Debug().Str("queue.name", d.config.Queue.Name).Msg("skipping queue declare")
		return nil
	}

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
	sdk.Logger(ctx).Debug().Any("queueConfig", d.config.Queue).Msg("queue declared")

	return nil
}

func (d *Destination) declareExchange(ctx context.Context) (err error) {
	if d.config.Exchange.SkipDeclare {
		sdk.Logger(ctx).Debug().Str("exchange.name", d.config.Exchange.Name).Msg("skipping exchange declare")
		return nil
	}

	if d.config.Exchange.Name == "" {
		// Skip declaring exchange with an empty name.
		return nil
	}

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
	sdk.Logger(ctx).Debug().Any("exchangeConfig", d.config.Exchange).Msgf("exchange declared")

	if d.getRoutingKey == nil {
		err = d.ch.QueueBind(d.config.Queue.Name, d.routingKey, d.config.Exchange.Name, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
		sdk.Logger(ctx).Debug().Msgf(
			"bound queue %s to exchange %s with routing key %s",
			d.config.Queue.Name, d.config.Exchange.Name, d.routingKey,
		)
	}

	return nil
}

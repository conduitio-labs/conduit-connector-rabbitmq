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

type Source struct {
	sdk.UnimplementedSource

	conn  *amqp091.Connection
	ch    *amqp091.Channel
	queue amqp091.Queue
	msgs  <-chan amqp091.Delivery

	config SourceConfig
}

func (s *Source) Config() sdk.SourceConfig {
	return &s.config
}

func NewSource() sdk.Source {
	return sdk.SourceWithMiddleware(&Source{})
}

func (s *Source) Open(ctx context.Context, sdkPos opencdc.Position) (err error) {
	s.conn, err = ampqDial(ctx, s.config.Config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("connected to RabbitMQ")

	s.ch, err = s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("opened channel")

	if sdkPos != nil {
		pos, err := parsePosition(sdkPos)
		if err != nil {
			return fmt.Errorf("failed to parse position: %w", err)
		}

		if s.config.Queue.Name != "" && s.config.Queue.Name != pos.QueueName {
			return fmt.Errorf(
				"the old position contains a different queue name than the connector configuration (%q vs %q), please check if the configured queue name changed since the last run",
				pos.QueueName, s.config.Queue.Name,
			)
		}

		sdk.Logger(ctx).Debug().Msg("got queue name from given position")
		s.config.Queue.Name = pos.QueueName
	}

	s.queue, err = s.ch.QueueDeclare(
		s.config.Queue.Name,
		s.config.Queue.Durable,
		s.config.Queue.AutoDelete,
		s.config.Queue.Exclusive,
		s.config.Queue.NoWait,
		nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	sdk.Logger(ctx).Debug().Str("queueName", s.queue.Name).Msg("declared queue")

	s.msgs, err = s.ch.Consume(
		s.queue.Name,
		s.config.Consumer.Name,
		s.config.Consumer.AutoAck,
		s.config.Consumer.Exclusive,
		s.config.Consumer.NoLocal,
		s.config.Consumer.NoWait,
		nil)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}
	sdk.Logger(ctx).Debug().Str("queueName", s.queue.Name).Msg("subscribed to queue")

	return nil
}

func (s *Source) Read(ctx context.Context) (opencdc.Record, error) {
	var rec opencdc.Record

	select {
	case <-ctx.Done():
		err := ctx.Err()
		if err != nil {
			return rec, fmt.Errorf("context error: %w", err)
		}
		return rec, nil
	case msg, ok := <-s.msgs:
		if !ok {
			return rec, errors.New("source message channel closed")
		}

		var (
			pos = Position{
				DeliveryTag: msg.DeliveryTag,
				QueueName:   s.queue.Name,
			}
			sdkPos   = pos.ToSdkPosition()
			metadata = metadataFromMessage(msg)
			key      = opencdc.RawData(msg.MessageId)
			payload  = opencdc.RawData(msg.Body)
		)

		rec = sdk.Util.Source.NewRecordCreate(sdkPos, metadata, key, payload)

		sdk.Logger(ctx).Trace().Msgf("read message %s from %s", msg.MessageId, s.queue.Name)

		return rec, nil
	}
}

func (s *Source) Ack(_ context.Context, position opencdc.Position) error {
	pos, err := parsePosition(position)
	if err != nil {
		return fmt.Errorf("failed to parse position: %w", err)
	}

	if err := s.ch.Ack(pos.DeliveryTag, false); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	return nil
}

func (s *Source) Teardown(ctx context.Context) error {
	errs := make([]error, 0, 2)
	if s.ch != nil {
		if err := s.ch.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	sdk.Logger(ctx).Debug().Msg("source teardown complete")

	return nil
}

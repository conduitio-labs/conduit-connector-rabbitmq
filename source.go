package rabbitmq

//go:generate paramgen -output=paramgen_src.go SourceConfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

func NewSource() sdk.Source {
	return sdk.SourceWithMiddleware(&Source{}, sdk.DefaultSourceMiddleware()...)
}

func (s *Source) Parameters() map[string]sdk.Parameter {
	return s.config.Parameters()
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sdk.Logger(ctx).Info().Msg("Configuring Source...")
	err := sdk.Util.ParseConfig(cfg, &s.config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

func (s *Source) Open(ctx context.Context, pos sdk.Position) (err error) {
	s.conn, err = amqp091.Dial(s.config.URL)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	s.ch, err = s.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	s.queue, err = s.ch.QueueDeclare(s.config.QueueName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	s.msgs, err = s.ch.Consume(s.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	return nil
}

func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	var rec sdk.Record

	msg, ok := <-s.msgs
	if !ok {
		return rec, errors.New("source message channel closed")
	}

	pos := Position{DeliveryTag: msg.DeliveryTag}.ToSdkPosition()
	var metadata sdk.Metadata
	var key sdk.Data

	var payload sdk.Data = sdk.RawData(msg.Body)

	rec = sdk.Util.Source.NewRecordCreate(pos, metadata, key, payload)

	return rec, nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	pos, err := parseSdkPosition(position)
	if err != nil {
		return fmt.Errorf("failed to parse position: %w", err)
	}

	if err := s.ch.Ack(pos.DeliveryTag, false); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	return nil
}

func (s *Source) Teardown(ctx context.Context) error {
	chErr := s.ch.Close()
	connErr := s.conn.Close()

	return errors.Join(chErr, connErr)
}

type Position struct{ DeliveryTag uint64 }

func (p Position) ToSdkPosition() sdk.Position {
	bs, err := json.Marshal(p)
	if err != nil {
		// this error should not be possible
		panic(fmt.Errorf("error marshaling position to JSON: %w", err))
	}

	return sdk.Position(bs)
}

func parseSdkPosition(pos sdk.Position) (Position, error) {
	var p Position
	err := json.Unmarshal([]byte(pos), &p)
	if err != nil {
		return p, fmt.Errorf("error unmarshaling position from JSON: %w", err)
	}

	return p, nil
}

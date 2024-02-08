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
	"encoding/json"
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/rabbitmq/amqp091-go"
)

type Position struct {
	DeliveryTag uint64
	QueueName   string
}

func (p Position) ToSdkPosition() sdk.Position {
	bs, err := json.Marshal(p)
	if err != nil {
		// this error should not be possible
		panic(fmt.Errorf("error marshaling position to JSON: %w", err))
	}

	return sdk.Position(bs)
}

func parsePosition(pos sdk.Position) (Position, error) {
	var p Position
	err := json.Unmarshal([]byte(pos), &p)
	if err != nil {
		return p, fmt.Errorf("error unmarshaling position from JSON: %w", err)
	}

	return p, nil
}

func metadataFromMessage(msg amqp091.Delivery) sdk.Metadata {
	metadata := sdk.Metadata{}

	setKey := func(key string, v any) {
		if s, ok := v.(string); ok {
			metadata[key] = s
			return
		}

		metadata[key] = fmt.Sprintf("%v", v)
	}

	setKey("rabbitmq.queueName", msg.MessageCount)
	setKey("rabbitmq.contentType", msg.ContentType)
	setKey("rabbitmq.contentEncoding", msg.ContentEncoding)
	setKey("rabbitmq.deliveryMode", msg.DeliveryMode)
	setKey("rabbitmq.priority", msg.Priority)
	setKey("rabbitmq.correlationId", msg.CorrelationId)
	setKey("rabbitmq.replyTo", msg.ReplyTo)
	setKey("rabbitmq.expiration", msg.Expiration)
	setKey("rabbitmq.timestamp", msg.Timestamp)
	setKey("rabbitmq.type", msg.Type)
	setKey("rabbitmq.userId", msg.UserId)
	setKey("rabbitmq.appId", msg.AppId)
	setKey("rabbitmq.consumerTag", msg.ConsumerTag)
	setKey("rabbitmq.messageCount", msg.MessageCount)
	setKey("rabbitmq.deliveryTag", msg.DeliveryTag)
	setKey("rabbitmq.redelivered", msg.Redelivered)
	setKey("rabbitmq.exchange", msg.Exchange)
	setKey("rabbitmq.routingKey", msg.RoutingKey)

	return metadata
}

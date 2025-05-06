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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/conduitio/conduit-commons/opencdc"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
	"github.com/rabbitmq/amqp091-go"
)

func generate3Records(queueName, routingKey string) []opencdc.Record {
	var recs []opencdc.Record

	for i := range 3 {
		pos := fmt.Sprintf(`{"deliveryTag":%d,"queueName":"%s"}`, i, queueName)
		payload := fmt.Sprintf("example message %d", i)

		position := []byte(pos)

		rec := sdk.Util.Source.NewRecordCreate(
			position,
			opencdc.Metadata{
				"rabbitmq.queue":                            queueName,
				MetadataRabbitmqRoutingKey:                  routingKey,
				MetadataRabbitmqHeaderPrefix + "queue_name": queueName,
			},
			opencdc.StructuredData{"id": i},
			opencdc.RawData(payload),
		)

		recs = append(recs, rec)
	}

	return recs
}

func testExchange(is *is.I, queueName, exchangeName, exchangeType, routingKey, routingKeyTemplate string) {
	ctx := context.Background()

	sourceCfg := map[string]string{
		"url":        testURL,
		"queue.name": queueName,
	}

	routingKeyCfg := routingKey
	if routingKeyTemplate != "" {
		routingKeyCfg = routingKeyTemplate
	}

	destCfg := map[string]string{
		"url":                  testURL,
		"queue.name":           queueName,
		"delivery.contentType": "text/plain",
		"exchange.name":        exchangeName,
		"exchange.type":        exchangeType,
		"routingKey":           routingKeyCfg,
	}

	dest := NewDestination()
	err := sdk.Util.ParseConfig(ctx, destCfg, dest.Config(), Connector.NewSpecification().DestinationParams)
	is.NoErr(err)

	err = dest.Open(ctx)
	is.NoErr(err)
	defer teardownResource(ctx, is, dest)

	recs := generate3Records(queueName, routingKey)
	_, err = dest.Write(ctx, recs)
	is.NoErr(err)

	src := NewSource()
	err = sdk.Util.ParseConfig(ctx, sourceCfg, src.Config(), Connector.NewSpecification().SourceParams)
	is.NoErr(err)

	err = src.Open(ctx, nil)
	is.NoErr(err)
	defer teardownResource(ctx, is, src)

	assertNextPayloadIs := func(expectedPayload string) {
		readRec, err := src.Read(ctx)
		is.NoErr(err)

		is.Equal(readRec.Metadata[MetadataRabbitmqRoutingKey], routingKey)
		is.Equal(readRec.Metadata[MetadataRabbitmqHeaderPrefix+"queue_name"], queueName)

		var rec struct {
			Payload struct {
				After string `json:"after"`
			} `json:"payload"`
		}
		err = json.Unmarshal(readRec.Payload.After.Bytes(), &rec)
		is.NoErr(err)

		body, err := base64.StdEncoding.DecodeString(rec.Payload.After)
		is.NoErr(err)

		is.Equal(string(body), expectedPayload)
	}

	assertNextPayloadIs("example message 0")
	assertNextPayloadIs("example message 1")
	assertNextPayloadIs("example message 2")
}

func TestDestination_ExchangeWorks(t *testing.T) {
	is := is.New(t)
	testExchange(is, "testDirectQueue", "testDirectExchange", "direct", "specificRoutingKey", "")
	testExchange(is, "testFanoutQueue", "testFanoutExchange", "fanout", "testFanoutQueue", "")
	testExchange(is, "testTopicQueue", "testTopicExchange", "topic", "specificRoutingKey", "")
}

func TestDestination_MetadataRoutingKeyWorks(t *testing.T) {
	is := is.New(t)

	exchangeName := "testTopicExchangeWithMetadataRoutingKey"
	queueName := "testTopicQueueWithMetadataRoutingKey"
	routingKey := "specificRoutingKey"

	// Create queue to exchange binding
	conn, err := amqp091.Dial(testURL)
	is.NoErr(err)
	ch, err := conn.Channel()
	is.NoErr(err)

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	is.NoErr(err)

	err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
	is.NoErr(err)
	err = ch.QueueBind(queueName, routingKey, exchangeName, false, nil)
	is.NoErr(err)
	conn.Close()

	testExchange(is,
		queueName,
		exchangeName,
		"topic",
		routingKey,
		`{{ index .Metadata "rabbitmq.routingKey" }}`,
	)
}

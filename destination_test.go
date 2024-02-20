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

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func generate3Records(queueName string) []sdk.Record {
	recs := []sdk.Record{}

	for i := 0; i < 3; i++ {
		exampleMessage := fmt.Sprintf("example message %d", i)

		rec := sdk.Util.Source.NewRecordCreate(
			[]byte(uuid.NewString()),
			sdk.Metadata{"rabbitmq.queue": queueName},
			sdk.RawData("test-key"),
			sdk.RawData(exampleMessage),
		)

		recs = append(recs, rec)
	}

	return recs
}

func testExchange(is *is.I, queueName, exchangeName, exchangeType, routingKey string) {
	ctx := context.Background()
	sharedCfg := Config{URL: testURL}

	dest := NewDestination()
	destCfg := cfgToMap(DestinationConfig{
		Config: sharedCfg,
		Delivery: DeliveryConfig{
			ContentType: "text/plain",
		},
		Queue: QueueConfig{
			Name:       queueName,
			Durable:    false,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
		},
		Exchange: ExchangeConfig{
			Name:       exchangeName,
			Type:       exchangeType,
			Durable:    false,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
		},
		RoutingKey: routingKey,
	})
	err := dest.Configure(ctx, destCfg)
	is.NoErr(err)

	err = dest.Open(ctx)
	is.NoErr(err)
	defer teardownResource(ctx, is, dest)

	recs := generate3Records(queueName)
	_, err = dest.Write(ctx, recs)
	is.NoErr(err)

	src := NewSource().(*Source)
	srcCfg := cfgToMap(SourceConfig{Config: sharedCfg})
	err = src.Configure(ctx, srcCfg)
	is.NoErr(err)

	err = src.Open(ctx, nil)
	is.NoErr(err)
	defer teardownResource(ctx, is, src)

	assertNextPayloadIs := func(expectedPayload string) {
		readRec, err := src.Read(ctx)
		is.NoErr(err)

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
	testExchange(is, "testQueue", "testDirectExchange", "direct", "specificRoutingKey")
	testExchange(is, "testQueue", "testFanoutExchange", "fanout", "")
	testExchange(is, "testQueue", "testTopicExchange", "topic", "specificRoutingKey")
}

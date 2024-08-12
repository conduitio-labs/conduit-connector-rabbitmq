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
	"fmt"
	"testing"
	"time"

	"github.com/conduitio/conduit-commons/config"
	"github.com/conduitio/conduit-commons/opencdc"
	"github.com/matryer/is"
	"github.com/rabbitmq/amqp091-go"
)

func TestTeardownSource_NoOpen(t *testing.T) {
	is := is.New(t)
	con := NewSource()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}

func TestSource_Integration_RestartFull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	is := is.New(t)

	queueName := setupQueueName(t, is)
	sourceCfg := config.Config{
		SourceConfigUrl:          testURL,
		SourceConfigQueueName:    queueName,
		SourceConfigQueueDurable: "false",
	}

	recs1 := generateRabbitmqMsgs(1, 3)
	produceRabbitmqMsgs(ctx, is, queueName, recs1)
	lastPosition := testSourceIntegrationRead(ctx, is, sourceCfg, nil, recs1, false)

	recs2 := generateRabbitmqMsgs(4, 6)
	produceRabbitmqMsgs(ctx, is, queueName, recs2)

	testSourceIntegrationRead(ctx, is, sourceCfg, lastPosition, recs2, false)
}

func TestSource_Integration_RestartPartial(t *testing.T) {
	t.Parallel()

	is := is.New(t)
	ctx := context.Background()
	queueName := setupQueueName(t, is)

	sourceCfg := config.Config{
		SourceConfigUrl:          testURL,
		SourceConfigQueueName:    queueName,
		SourceConfigQueueDurable: "false",
	}

	recs1 := generateRabbitmqMsgs(1, 3)
	produceRabbitmqMsgs(ctx, is, queueName, recs1)

	lastPosition := testSourceIntegrationRead(ctx, is, sourceCfg, nil, recs1, true)

	// only first record was acked, produce more records and expect to resume
	// from last acked record
	recs2 := generateRabbitmqMsgs(4, 6)
	produceRabbitmqMsgs(ctx, is, queueName, recs2)

	var wantRecs []amqp091.Publishing
	wantRecs = append(wantRecs, recs1[1:]...)
	wantRecs = append(wantRecs, recs2...)

	testSourceIntegrationRead(ctx, is, sourceCfg, lastPosition, wantRecs, false)
}

const testAppID = "id-1234"

func generateRabbitmqMsgs(from, to int) []amqp091.Publishing {
	var msgs []amqp091.Publishing

	for i := from; i <= to; i++ {
		msg := amqp091.Publishing{
			MessageId:   fmt.Sprintf("test-msg-id-%d", i),
			ContentType: "text/plain",
			// setting testAppId asserts that the metadata is being set
			AppId: testAppID,

			Body: []byte(fmt.Sprintf("test-payload-%d", i)),
		}

		msgs = append(msgs, msg)
	}

	return msgs
}

func produceRabbitmqMsgs(ctx context.Context, is *is.I, queueName string, msgs []amqp091.Publishing) {
	conn, err := amqp091.Dial(testURL)
	is.NoErr(err)

	defer conn.Close()

	ch, err := conn.Channel()
	is.NoErr(err)

	q, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	is.NoErr(err)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, msg := range msgs {
		err = ch.PublishWithContext(ctx, "", q.Name, false, false, msg)
		is.NoErr(err)
	}
}

// testSourceIntegrationRead reads and acks messages in range [from,to].
// If ackFirstOnly is true, only the first message will be acknowledged.
// Returns the position of the last message read.
func testSourceIntegrationRead(
	ctx context.Context,
	is *is.I,
	cfg config.Config,
	startFrom opencdc.Position,
	wantRecords []amqp091.Publishing,
	ackFirstOnly bool,
) opencdc.Position {
	is.Helper()
	underTest := NewSource()
	defer func() {
		err := underTest.Teardown(ctx)
		is.NoErr(err)
	}()

	err := underTest.Configure(ctx, cfg)
	is.NoErr(err)
	err = underTest.Open(ctx, startFrom)
	is.NoErr(err)

	var positions []opencdc.Position
	for _, wantRecord := range wantRecords {
		rec, err := underTest.Read(ctx)
		is.NoErr(err)

		recPayload := string(rec.Payload.After.Bytes())
		wantPayload := string(wantRecord.Body)
		is.Equal(wantPayload, recPayload)

		is.Equal(wantRecord.MessageId, string(rec.Key.Bytes()))
		is.Equal(testAppID, rec.Metadata["rabbitmq.appId"])

		positions = append(positions, rec.Position)
	}

	for i, p := range positions {
		if i > 0 && ackFirstOnly {
			break
		}
		err = underTest.Ack(ctx, p)
		is.NoErr(err)
	}

	return positions[len(positions)-1]
}

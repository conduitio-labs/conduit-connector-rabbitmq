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
	"bytes"
	"encoding/base64"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/alarbada/conduit-connector-rabbitmq/test"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
)

func init() {
	// Uncomment this up a logger for tests to use. By default sdk.Logger log calls won't
	// output anything
	// log := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// zerolog.DefaultContextLogger = &log
}


func TestAcceptance(t *testing.T) {
	cfg := map[string]string{
		"url":       test.URL,
		"queueName": "test-queue2",
	}
	is := is.New(t)

	sdk.AcceptanceTest(t, AcceptanceTestDriver{
		sdk.ConfigurableAcceptanceTestDriver{
			Config: sdk.ConfigurableAcceptanceTestDriverConfig{
				Connector:         Connector,
				SourceConfig:      cfg,
				DestinationConfig: cfg,
				BeforeTest: func(t *testing.T) {
					queueName := test.SetupQueueName(t, is)
					cfg["queueName"] = queueName
				},
				Skip: []string{
					"TestSource_Configure_RequiredParams",
					"TestDestination_Configure_RequiredParams",
				},
				WriteTimeout: time.Second * 10,
				ReadTimeout:  time.Second * 10,
			},
		},
	})
}

type AcceptanceTestDriver struct {
	sdk.ConfigurableAcceptanceTestDriver
}

func encodeData(bs sdk.Data, t *testing.T) sdk.Data {
	keyBuff := bytes.NewBuffer([]byte{})
	_, err := base64.NewEncoder(base64.StdEncoding, keyBuff).Write(bs.Bytes())
	if err != nil {
		t.Fatalf("failed to encode record: %v", err)
	}

	return sdk.RawData(keyBuff.Bytes())
}

// GenerateRecord overrides the sdk.ConfigurableAcceptanceTestDriver method
// to generate a record with a simpler data randomizer.
// The reason is that there's an unexpected behaviour with how rabbitmq handles message ids.
func (d AcceptanceTestDriver) GenerateRecord(t *testing.T, op sdk.Operation) sdk.Record {
	rec := d.ConfigurableAcceptanceTestDriver.GenerateRecord(t, op)
	keyBuff := bytes.NewBuffer([]byte{})
	_, err := base64.NewEncoder(base64.StdEncoding, keyBuff).Write(rec.Bytes())
	if err != nil {
		t.Fatalf("failed to encode record: %v", err)
	}

	rec.Key = sdk.RawData(keyBuff.Bytes())

	return sdk.Record{
		Position:  sdk.Position(randBytes()),
		Operation: op,
		Metadata: map[string]string{
			randString(): randString(),
			randString(): randString(),
		},
		Key: randData(),
		Payload: sdk.Change{
			Before: nil,
			After:  randData(),
		},
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz1234567890"

func randBytes() []byte {
	max := 32
	var sb strings.Builder
	sb.Grow(max)
	sb.WriteString("test-")
	for i := 0; i < max; i++ {
		r := rand.Intn(len(charset))
		if r == 0 {
			r = 1
		}
		runeChar := charset[r-1 : r]

		sb.WriteString(runeChar)
	}

	return []byte(sb.String())
}

func randData() sdk.Data { return sdk.RawData(randBytes()) }
func randString() string { return string(randBytes()) }

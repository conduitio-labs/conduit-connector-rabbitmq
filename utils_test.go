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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/matryer/is"
	"github.com/rabbitmq/amqp091-go"
)

func init() {
	// Uncomment this to set up a logger for tests to use. By default
	// sdk.Logger log calls won't output anything
	// log := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// zerolog.DefaultContextLogger = &log
}

// cfgToMap converts a config struct to a map. This is useful for more type
// safety on tests.
func cfgToMap(cfg any) map[string]string {
	bs, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	mAny := map[string]any{}
	err = json.Unmarshal(bs, &mAny)
	if err != nil {
		panic(err)
	}

	m := map[string]string{}
	for k, v := range mAny {
		switch v := v.(type) {
		case string:
			m[k] = v
		case bool:
			m[k] = fmt.Sprintf("%t", v)
		case map[string]any:
			parsed := cfgToMap(v)
			for k2, v := range parsed {
				m[k+"."+k2] = v
			}
		default:
			panic(fmt.Errorf("unsupported type: %T", v))
		}
	}

	return m
}

const (
	testURL    = "amqp://guest:guest@localhost:5672"
	testURLTLS = "amqps://guest:guest@localhost:5671"
)

// setupQueueName creates a new topic name for the test and deletes it if it
// exists, so that the test can start from a clean slate.
func setupQueueName(t *testing.T, is *is.I) string {
	queueName := "rabbitmq.queue." + t.Name()
	deleteQueue(is, queueName, nil)

	return queueName
}

// setupQueueNameTLS does the same as setupQueueName but for TLS connections.
func setupQueueNameTLS(t *testing.T, is *is.I, tlsConfig *tls.Config) string {
	queueName := "rabbitmq.queue." + t.Name()
	deleteQueue(is, queueName, tlsConfig)

	return queueName
}

func deleteQueue(is *is.I, queueName string, tlsConfig *tls.Config) {
	var conn *amqp091.Connection
	var err error
	if tlsConfig != nil {
		conn, err = amqp091.DialTLS(testURLTLS, tlsConfig)
		is.NoErr(err)
	} else {
		conn, err = amqp091.Dial(testURL)
		is.NoErr(err)
	}
	defer closeResource(is, conn)

	ch, err := conn.Channel()
	is.NoErr(err)
	defer closeResource(is, ch)

	// force queue delete
	_, err = ch.QueueDelete(queueName, false, false, false)
	is.NoErr(err)
}

type closable interface {
	Close() error
}

func closeResource(is *is.I, c closable) {
	err := c.Close()
	is.NoErr(err)
}

type teardownable interface {
	Teardown(context.Context) error
}

func teardownResource(ctx context.Context, is *is.I, t teardownable) {
	err := t.Teardown(ctx)
	is.NoErr(err)
}

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

package test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/rabbitmq/amqp091-go"
)

const URL = "amqp://guest:guest@localhost:5672"

// SetupQueueName creates a new topic name for the test and deletes it if it
// exists, so that the test can start from a clean slate.
func SetupQueueName(t *testing.T, is *is.I) string {
	queueName := "rabbitmq.queue." + t.Name()
	DeleteQueue(is, queueName)

	return queueName
}

func DeleteQueue(is *is.I, queueName string) {
	conn, err := amqp091.Dial(URL)
	is.NoErr(err)
	defer CloseResource(is, conn)

	ch, err := conn.Channel()
	is.NoErr(err)
	defer CloseResource(is, ch)

	// force queue delete
	_, err = ch.QueueDelete(queueName, false, false, false)
	is.NoErr(err)
}

type closable interface {
	Close() error
}

func CloseResource(is *is.I, c closable) {
	err := c.Close()
	is.NoErr(err)
}

type teardownable interface {
	Teardown(context.Context) error
}

func TeardownResource(ctx context.Context, is *is.I, t teardownable) {
	err := t.Teardown(ctx)
	is.NoErr(err)
}

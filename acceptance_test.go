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
	"testing"
	"time"

	"github.com/alarbada/conduit-connector-rabbitmq/test"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
)

func init() {
	// Uncomment this to set up a logger for tests to use. By default
	// sdk.Logger log calls won't output anything
	// log := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// zerolog.DefaultContextLogger = &log
}

func TestAcceptance(t *testing.T) {
	cfg := map[string]string{
		"url":       test.URL,
		"queueName": "test-queue",
	}
	is := is.New(t)

	driver := sdk.ConfigurableAcceptanceTestDriver{
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
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, driver)
}

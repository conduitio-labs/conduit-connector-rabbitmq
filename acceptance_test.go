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
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
)

func TestAcceptance(t *testing.T) {
	cfg := map[string]string{
		"url":        testURL,
		"queue.name": "test-queue",
	}
	is := is.New(t)

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      cfg,
			DestinationConfig: cfg,
			BeforeTest: func(t *testing.T) {
				queueName := setupQueueName(t, is)
				cfg["queue.name"] = queueName
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

func TestAcceptance_TLS(t *testing.T) {
	is := is.New(t)

	cfg := Config{
		URL: testURLTLS,
		TLS: TLSConfig{
			Enabled:    true,
			ClientCert: "./test/certs/client.cert.pem",
			ClientKey:  "./test/certs/client.key.pem",
			CACert:     "./test/certs/ca.cert.pem",
		},
	}

	sourceCfg := SourceConfig{Config: cfg}.toMap()
	destCfg := DestinationConfig{Config: cfg}.toMap()

	ctx := context.Background()

	tlsConfig, err := parseTLSConfig(ctx, cfg)
	is.NoErr(err)

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      sourceCfg,
			DestinationConfig: destCfg,
			BeforeTest: func(t *testing.T) {
				queueName := setupQueueNameTLS(t, is, tlsConfig)
				sourceCfg["queue.name"] = queueName
				destCfg["queue.name"] = queueName
				// TODO: I can see why this is necessary, but then, why does
				// the previous acceptance test not need it?
				destCfg["routingKey"] = queueName
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

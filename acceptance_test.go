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
	"os"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
)

func TestAcceptance(t *testing.T) {
	cfg := map[string]string{
		"url":       testURL,
		"queueName": "test-queue",
	}
	is := is.New(t)

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      cfg,
			DestinationConfig: cfg,
			BeforeTest: func(t *testing.T) {
				queueName := setupQueueName(t, is)
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

func TestAcceptance_TLS(t *testing.T) {
	if os.Getenv("RABBITMQ_TLS") != "true" {
		t.Skip()
	}

	is := is.New(t)

// 	_, err := os.ReadFile("./test/client.cert.pem")
// 	is.NoErr(err)
// 	_, err = os.ReadFile("./test/client.key.pem")
// 	is.NoErr(err)
// 	_, err = os.ReadFile("./test/ca.cert.pem")
// 	is.NoErr(err)

	// return

	sharedCfg := Config{
		URL:        testURLTLS,
		QueueName:  "test-queue",
		ClientCert: "./test/client.cert.pem",
		ClientKey:  "./test/client.key.pem",
		CACert:     "./test/ca.cert.pem",
	}
	cfg := cfgToMap(sharedCfg)
	ctx := context.Background()

	tlsConfig, err := parseTLSConfig(ctx, sharedCfg)
	is.NoErr(err)

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      cfg,
			DestinationConfig: cfg,
			BeforeTest: func(t *testing.T) {
				queueName := setupQueueNameTLS(t, is, tlsConfig)
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

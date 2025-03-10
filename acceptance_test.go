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

	"github.com/conduitio/conduit-commons/config"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
)

func TestAcceptance(t *testing.T) {
	sourceCfg := config.Config{"url": testURL}
	destCfg := config.Config{"url": testURL}
	is := is.New(t)

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      sourceCfg,
			DestinationConfig: destCfg,
			BeforeTest: func(t *testing.T) {
				queueName := setupQueueName(t, is)
				sourceCfg["queue.name"] = queueName
				destCfg["queue.name"] = queueName
			},
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, driver)
}

func TestAcceptance_TLS(t *testing.T) {
	is := is.New(t)

	sourceCfg := config.Config{
		"url":            testURLTLS,
		"queue.name":     "random-queue",
		"tls.enabled":    "true",
		"tls.clientCert": "./test/certs/client.cert.pem",
		"tls.clientKey":  "./test/certs/client.key.pem",
		"tls.caCert":     "./test/certs/ca.cert.pem",
	}
	destCfg := config.Config{
		"url":            testURLTLS,
		"queue.name":     "random-queue",
		"tls.enabled":    "true",
		"tls.clientCert": "./test/certs/client.cert.pem",
		"tls.clientKey":  "./test/certs/client.key.pem",
		"tls.caCert":     "./test/certs/ca.cert.pem",
	}

	ctx := context.Background()
	var cfg Config
	err := sdk.Util.ParseConfig(ctx, sourceCfg, &cfg, Connector.NewSpecification().SourceParams)
	is.NoErr(err)

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
				destCfg["routingKey"] = queueName
			},
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, driver)
}

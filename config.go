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
	"fmt"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

//go:generate paramgen -output=paramgen_src.go SourceConfig
//go:generate paramgen -output=paramgen_dest.go DestinationConfig

type Config struct {
	// URL is the RabbitMQ server URL
	URL string `json:"url" validate:"required"`

	// QueueName is the name of the queue to consume from / publish to
	QueueName string `json:"queueName" validate:"required"`

	ClientCert string `json:"clientCert"`
	ClientKey  string `json:"clientKey"`
	CACert     string `json:"caCert"`

	TLSInsecureSkipVerify bool `json:"tlsInsecureSkipVerify"`
}

type SourceConfig struct {
	Config
}

type DestinationConfig struct {
	Config

	// ContentType is the MIME content type of the messages written to rabbitmq
	ContentType string `json:"contentType" default:"text/plain"`

	// ExchangeName is the name of the exchange to publish to
	ExchangeName string `json:"exchangeName" default:""`

	// ExchangeType is the type of the exchange to publish to
	ExchangeType string `json:"exchangeType" default:"direct"`

	// RoutingKey is the routing key to use when publishing to an exchange
	RoutingKey string `json:"routingKey" default:""`
}

func newDestinationConfig(cfg map[string]string) (DestinationConfig, error) {
	var destCfg DestinationConfig
	err := sdk.Util.ParseConfig(cfg, &destCfg)
	if err != nil {
		return destCfg, fmt.Errorf("invalid config: %w", err)
	}

	if destCfg.ContentType == "" {
		destCfg.ContentType = "text/plain"
	}

	return destCfg, nil
}

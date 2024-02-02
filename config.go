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

type Config struct {
	URL       string `json:"url" validate:"required"`
	QueueName string `json:"queueName" validate:"required"`
}

type SourceConfig struct {
	Config
}

type DestinationConfig struct {
	Config

	ContentType string `json:"contentType" validate:"required"`
}

func newDestinationConfig(cfg map[string]string) (DestinationConfig, error) {
	var destCfg DestinationConfig
	err := sdk.Util.ParseConfig(cfg, cfg)
	if err != nil {
		return destCfg, fmt.Errorf("invalid config: %w", err)
	}

	if destCfg.ContentType == "" {
		destCfg.ContentType = "text/plain"
	}

	return destCfg, nil
}

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

	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Config struct {
	// URL is the RabbitMQ server URL
	URL string `json:"url" validate:"required"`

	TLS   TLSConfig   `json:"tls"`
	Queue QueueConfig `json:"queue"`
}

type TLSConfig struct {
	// Enabled indicates if TLS should be used
	Enabled bool `json:"enabled" default:"false"`

	// ClientCert is the path to the client certificate to use for TLS
	ClientCert string `json:"clientCert"`

	// ClientKey is the path to the client key to use for TLS
	ClientKey string `json:"clientKey"`

	// CACert is the path to the CA certificate to use for TLS
	CACert string `json:"caCert"`
}

type QueueConfig struct {
	// Name is the name of the queue to consume from / publish to
	Name string `json:"name" validate:"required"`

	// Durable indicates if the queue will survive broker restarts.
	Durable bool `json:"durable" default:"true"`

	// AutoDelete indicates if the queue will be deleted when there are no more consumers.
	AutoDelete bool `json:"autoDelete" default:"false"`

	// Exclusive indicates if the queue can be accessed by other connections.
	Exclusive bool `json:"exclusive" default:"false"`

	// NoWait indicates if the queue should be declared without waiting for server confirmation.
	NoWait bool `json:"noWait" default:"false"`
}

// to use with ampq.Channel Consume method.
type ConsumerConfig struct {
	// Name is the name of the consumer
	Name string `json:"name"`

	// AutoAck indicates if the server should consider messages acknowledged once delivered.
	AutoAck bool `json:"autoAck" default:"false"`

	// Exclusive indicates if the consumer should be exclusive.
	Exclusive bool `json:"exclusive" default:"false"`

	// NoLocal indicates if the server should not deliver messages published by the same connection.
	NoLocal bool `json:"noLocal" default:"false"`

	// NoWait indicates if the consumer should be declared without waiting for server confirmation.
	NoWait bool `json:"noWait" default:"false"`
}

type SourceConfig struct {
	sdk.DefaultSourceMiddleware
	Config

	Consumer ConsumerConfig `json:"consumer"`
}

type ExchangeConfig struct {
	// Name is the name of the exchange.
	Name string `json:"name"`

	// Type is the type of the exchange (e.g., direct, fanout, topic, headers).
	Type string `json:"type"`

	// Durable indicates if the exchange will survive broker restarts.
	Durable bool `json:"durable" default:"true"`

	// AutoDelete indicates if the exchange will be deleted when the last queue is unbound from it.
	AutoDelete bool `json:"autoDelete" default:"false"`

	// Internal indicates if the exchange is used for internal purposes and cannot be directly published to by a client.
	Internal bool `json:"internal" default:"false"`

	// NoWait indicates if the exchange should be declared without waiting for server confirmation.
	NoWait bool `json:"noWait" default:"false"`
}

type DeliveryConfig struct {
	// ContentEncoding specifies the encoding of the message content.
	ContentEncoding string `json:"contentEncoding"`

	// ContentType specifies the MIME type of the message content. Defaults to "application/json".
	ContentType string `json:"contentType" default:"application/json"`

	// DeliveryMode indicates the message delivery mode. Non-persistent (1) or persistent (2). Default is 2 (persistent).
	DeliveryMode uint8 `json:"deliveryMode" default:"2" validation:"inclusion=1|2"`

	// Priority specifies the message priority. Ranges from 0 to 9. Default is 0.
	Priority uint8 `json:"priority" default:"0" validate:"greater-than=-1,less-than=10"`

	// CorrelationID is used to correlate RPC responses with requests.
	CorrelationID string `json:"correlationID" default:""`

	// ReplyTo specifies the address to reply to.
	ReplyTo string `json:"replyTo" default:""`

	// MessageTypeName specifies the message type name.
	MessageTypeName string `json:"messageTypeName" default:""`

	// UserID specifies the user who created the message. Useful for publishers.
	UserID string `json:"userID" default:""`

	// AppID specifies the application that created the message.
	AppID string `json:"appID" default:""`

	// Mandatory indicates if the message is mandatory. If true, tells the server to return the message if it cannot be routed to a queue.
	Mandatory bool `json:"mandatory" default:"false"`

	// Immediate indicates if the message should be treated as immediate. If true, the message is not queued if no consumers are on the matching queue.
	Immediate bool `json:"immediate" default:"false"`

	// Expiration specifies the message expiration time, if any.
	Expiration string `json:"expiration" default:""`
}

type DestinationConfig struct {
	sdk.DefaultDestinationMiddleware
	Config

	Delivery DeliveryConfig `json:"delivery"`
	Exchange ExchangeConfig `json:"exchange"`

	// RoutingKey is the routing key to use when publishing to an exchange
	RoutingKey string `json:"routingKey" default:""`
}

func (config *DestinationConfig) Validate(context.Context) error {
	if config.Delivery.ContentType == "" {
		config.Delivery.ContentType = "text/plain"
	}

	if config.RoutingKey == "" {
		config.RoutingKey = config.Queue.Name
	}

	return nil
}

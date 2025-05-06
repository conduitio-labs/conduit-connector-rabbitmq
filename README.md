# Conduit Connector for RabbitMQ

The RabbitMQ connector is one of [Conduit](https://github.com/ConduitIO/conduit) standalone plugins. It provides both a source and a destination connector for [RabbitMQ](https://rabbitmq.com/).

It uses the [AMQP 0-9-1 Model](https://www.rabbitmq.com/tutorials/amqp-concepts) to connect to RabbitMQ.

## What data does the OpenCDC record consist of?

| Field                   | Description                                                                               |
| ----------------------- | ----------------------------------------------------------------------------------------- |
| `record.Position`       | json object with the delivery tag and the queue name from where the record was read from. |
| `record.Operation`      | currently fixed as "create".                                                              |
| `record.Metadata`       | a string to string map, with keys prefixed as `rabbitmq.{DELIVERY_PROPERTY}`.             |
| `record.Key`            | the message id from the read message.                                                     |
| `record.Payload.Before` | <empty>                                                                                   |
| `record.Payload.After`  | the message body                                                                          |

## How to Build?

Run `make build` to compile the connector.

## Testing

Execute `make test` to perform all non-tls tests. Execute `make test-tls` for the TLS tests. Both commands use docker files located at `test/docker-compose.yml` and `test/docker-compose-tls.yml` respectively.
Tests require docker-compose v2.

## Source Configuration Parameters

<!-- readmegen:source.parameters.yaml -->
```yaml
version: 2.2
pipelines:
  - id: example
    status: running
    connectors:
      - id: example
        plugin: "rabbitmq"
        settings:
          # The name of the queue to consume from / publish to
          # Type: string
          # Required: yes
          queue.name: ""
          # The RabbitMQ server URL
          # Type: string
          # Required: yes
          url: ""
          # Indicates if the server should consider messages acknowledged once
          # delivered.
          # Type: bool
          # Required: no
          consumer.autoAck: "false"
          # Indicates if the consumer should be exclusive.
          # Type: bool
          # Required: no
          consumer.exclusive: "false"
          # The name of the consumer
          # Type: string
          # Required: no
          consumer.name: ""
          # Indicates if the server should not deliver messages published by the
          # same connection.
          # Type: bool
          # Required: no
          consumer.noLocal: "false"
          # Indicates if the consumer should be declared without waiting for
          # server confirmation.
          # Type: bool
          # Required: no
          consumer.noWait: "false"
          # Indicates if the queue will be deleted when there are no more
          # consumers.
          # Type: bool
          # Required: no
          queue.autoDelete: "false"
          # Indicates if the queue will survive broker restarts.
          # Type: bool
          # Required: no
          queue.durable: "true"
          # Indicates if the queue can be accessed by other connections.
          # Type: bool
          # Required: no
          queue.exclusive: "false"
          # Indicates if the queue should be declared without waiting for server
          # confirmation.
          # Type: bool
          # Required: no
          queue.noWait: "false"
          # The path to the CA certificate to use for TLS
          # Type: string
          # Required: no
          tls.caCert: ""
          # The path to the client certificate to use for TLS
          # Type: string
          # Required: no
          tls.clientCert: ""
          # The path to the client key to use for TLS
          # Type: string
          # Required: no
          tls.clientKey: ""
          # Indicates if TLS should be used
          # Type: bool
          # Required: no
          tls.enabled: "false"
          # Maximum delay before an incomplete batch is read from the source.
          # Type: duration
          # Required: no
          sdk.batch.delay: "0"
          # Maximum size of batch before it gets read from the source.
          # Type: int
          # Required: no
          sdk.batch.size: "0"
          # Specifies whether to use a schema context name. If set to false, no
          # schema context name will be used, and schemas will be saved with the
          # subject name specified in the connector (not safe because of name
          # conflicts).
          # Type: bool
          # Required: no
          sdk.schema.context.enabled: "true"
          # Schema context name to be used. Used as a prefix for all schema
          # subject names. If empty, defaults to the connector ID.
          # Type: string
          # Required: no
          sdk.schema.context.name: ""
          # Whether to extract and encode the record key with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.key.enabled: "true"
          # The subject of the key schema. If the record metadata contains the
          # field "opencdc.collection" it is prepended to the subject name and
          # separated with a dot.
          # Type: string
          # Required: no
          sdk.schema.extract.key.subject: "key"
          # Whether to extract and encode the record payload with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.payload.enabled: "true"
          # The subject of the payload schema. If the record metadata contains
          # the field "opencdc.collection" it is prepended to the subject name
          # and separated with a dot.
          # Type: string
          # Required: no
          sdk.schema.extract.payload.subject: "payload"
          # The type of the payload schema.
          # Type: string
          # Required: no
          sdk.schema.extract.type: "avro"
```
<!-- /readmegen:source.parameters.yaml -->

## Destination Configuration Parameters

<!-- readmegen:destination.parameters.yaml -->
```yaml
version: 2.2
pipelines:
  - id: example
    status: running
    connectors:
      - id: example
        plugin: "rabbitmq"
        settings:
          # The name of the queue to consume from / publish to
          # Type: string
          # Required: yes
          queue.name: ""
          # The RabbitMQ server URL
          # Type: string
          # Required: yes
          url: ""
          # The application that created the message.
          # Type: string
          # Required: no
          delivery.appID: ""
          # The encoding of the message content.
          # Type: string
          # Required: no
          delivery.contentEncoding: ""
          # The MIME type of the message content. Defaults to
          # "application/json".
          # Type: string
          # Required: no
          delivery.contentType: "application/json"
          # The correlation ID used to correlate RPC responses with requests.
          # Type: string
          # Required: no
          delivery.correlationID: ""
          # The message delivery mode. Non-persistent (1) or persistent (2).
          # Default is 2 (persistent).
          # Type: int
          # Required: no
          delivery.deliveryMode: "2"
          # The message expiration time, if any.
          # Type: string
          # Required: no
          delivery.expiration: ""
          # Indicates if the message should be treated as immediate. If true,
          # the message is not queued if no consumers are on the matching queue.
          # Type: bool
          # Required: no
          delivery.immediate: "false"
          # Indicates if the message is mandatory. If true, tells the server to
          # return the message if it cannot be routed to a queue.
          # Type: bool
          # Required: no
          delivery.mandatory: "false"
          # The message type name.
          # Type: string
          # Required: no
          delivery.messageTypeName: ""
          # The message priority. Ranges from 0 to 9. Default is 0.
          # Type: int
          # Required: no
          delivery.priority: "0"
          # The address to reply to.
          # Type: string
          # Required: no
          delivery.replyTo: ""
          # The user who created the message. Useful for publishers.
          # Type: string
          # Required: no
          delivery.userID: ""
          # Indicates if the exchange will be deleted when the last queue is
          # unbound from it.
          # Type: bool
          # Required: no
          exchange.autoDelete: "false"
          # Indicates if the exchange will survive broker restarts.
          # Type: bool
          # Required: no
          exchange.durable: "true"
          # Indicates if the exchange is used for internal purposes and cannot
          # be directly published to by a client.
          # Type: bool
          # Required: no
          exchange.internal: "false"
          # The name of the exchange.
          # Type: string
          # Required: no
          exchange.name: ""
          # Indicates if the exchange should be declared without waiting for
          # server confirmation.
          # Type: bool
          # Required: no
          exchange.noWait: "false"
          # The type of the exchange (e.g., direct, fanout, topic, headers).
          # Type: string
          # Required: no
          exchange.type: ""
          # Indicates if the queue will be deleted when there are no more
          # consumers.
          # Type: bool
          # Required: no
          queue.autoDelete: "false"
          # Indicates if the queue will survive broker restarts.
          # Type: bool
          # Required: no
          queue.durable: "true"
          # Indicates if the queue can be accessed by other connections.
          # Type: bool
          # Required: no
          queue.exclusive: "false"
          # Indicates if the queue should be declared without waiting for server
          # confirmation.
          # Type: bool
          # Required: no
          queue.noWait: "false"
          # The routing key to use when publishing to an exchange
          # Type: string
          # Required: no
          routingKey: "{{ index .Metadata "rabbitmq.routingKey" }}"
          # The path to the CA certificate to use for TLS
          # Type: string
          # Required: no
          tls.caCert: ""
          # The path to the client certificate to use for TLS
          # Type: string
          # Required: no
          tls.clientCert: ""
          # The path to the client key to use for TLS
          # Type: string
          # Required: no
          tls.clientKey: ""
          # Indicates if TLS should be used
          # Type: bool
          # Required: no
          tls.enabled: "false"
          # Maximum delay before an incomplete batch is written to the
          # destination.
          # Type: duration
          # Required: no
          sdk.batch.delay: "0"
          # Maximum size of batch before it gets written to the destination.
          # Type: int
          # Required: no
          sdk.batch.size: "0"
          # Allow bursts of at most X records (0 or less means that bursts are
          # not limited). Only takes effect if a rate limit per second is set.
          # Note that if `sdk.batch.size` is bigger than `sdk.rate.burst`, the
          # effective batch size will be equal to `sdk.rate.burst`.
          # Type: int
          # Required: no
          sdk.rate.burst: "0"
          # Maximum number of records written per second (0 means no rate
          # limit).
          # Type: float
          # Required: no
          sdk.rate.perSecond: "0"
          # The format of the output record. See the Conduit documentation for a
          # full list of supported formats
          # (https://conduit.io/docs/using/connectors/configuration-parameters/output-format).
          # Type: string
          # Required: no
          sdk.record.format: "opencdc/json"
          # Options to configure the chosen output record format. Options are
          # normally key=value pairs separated with comma (e.g.
          # opt1=val2,opt2=val2), except for the `template` record format, where
          # options are a Go template.
          # Type: string
          # Required: no
          sdk.record.format.options: ""
          # Whether to extract and decode the record key with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.key.enabled: "true"
          # Whether to extract and decode the record payload with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.payload.enabled: "true"
```
<!-- /readmegen:destination.parameters.yaml -->

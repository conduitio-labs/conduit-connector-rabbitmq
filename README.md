# Conduit Connector for RabbitMQ
The RabbitMQ connector is one of [Conduit](https://github.com/ConduitIO/conduit) builtin plugins. It provides both, a
source and a destination connector for [RabbitMQ](https://rabbitmq.com/).

## How to Build?

Run `make build` to compile the connector.

## Testing

Execute `make test` to perform all non-tls tests. Execute `make test-tls` for the TLS tests. Both commands use docker files located at `test/docker-compose.yml` and `test/docker-compose-tls.yml` respectively.
Tests require docker-compose v2.

## Shared Configuration Parameters

Shared between source and destination connectors.

| Name             | Description                                         | Required | Default Value |
|------------------|-----------------------------------------------------|----------|---------------|
| `url`            | The RabbitMQ server's URL.                          | Yes      |               |
| `tls.clientCert` | Path to the client certificate for TLS.             | No       |               |
| `tls.clientKey`  | Path to the client's key for TLS.                   | No       |               |
| `tls.caCert`     | Path to the CA (Certificate Authority) certificate for TLS. | No   |               |

## Source Connector

The source connector extracts data from RabbitMQ and sends it to downstream systems via Conduit.

### Source Configuration Parameters

| Name                   | Description                                                                 | Required | Default Value |
|------------------------|-----------------------------------------------------------------------------|----------|---------------|
| `queue.name`           | The name of the RabbitMQ queue to consume messages from.                    | Yes      |               |
| `queue.durable`        | Specifies whether the queue is durable.                                     | No       | `true`        |
| `queue.autoDelete`     | If the queue will auto-delete.                                              | No       | `false`       |
| `queue.exclusive`      | If the queue is exclusive.                                                  | No       | `false`       |
| `queue.noWait`         | If the queue is declared without waiting for server reply.                  | No       | `false`       |
| `consumer.name`        | The name of the consumer.                                                   | No       |               |
| `consumer.autoAck`     | If the server should consider messages acknowledged once delivered.         | No       | `false`       |
| `consumer.exclusive`   | If the consumer should be exclusive.                                        | No       | `false`       |
| `consumer.noLocal`     | If the server should not deliver messages published by the same connection. | No       | `false`       |
| `consumer.noWait`      | If the consumer should be declared without waiting for server confirmation. | No       | `false`       |

## Destination Connector

The destination connector sends data from upstream systems to RabbitMQ via Conduit.

### Destination Configuration Parameters

| Name                       | Description                                                         | Required | Default Value |
|----------------------------|---------------------------------------------------------------------|----------|---------------|
| `queue.name`               | The name of the RabbitMQ queue where messages will be published to. | Yes      |               |
| `queue.durable`            | Specifies whether the queue is durable.                             | No       | `true`        |
| `queue.autoDelete`         | If the queue will auto-delete.                                      | No       | `false`       |
| `queue.exclusive`          | If the queue is exclusive.                                          | No       | `false`       |
| `queue.noWait`             | If the queue is declared without waiting for server reply.          | No       | `false`       |
| `contentType`              | The MIME content type of the messages written to RabbitMQ.          | No       | `text/plain`  |
| `delivery.contentEncoding` | The content encoding for the message.                               | No       |               |
| `delivery.deliveryMode`    | Delivery mode of the message. Non-persistent (1) or persistent (2). | No       | `2`           |
| `delivery.priority`        | The priority of the message.                                        | No       | `0`           |
| `delivery.correlationID`   | The correlation id associated with the message.                     | No       |               |
| `delivery.replyTo`         | Address to reply to.                                                | No       |               |
| `delivery.messageTypeName` | The type name of the message.                                       | No       |               |
| `delivery.userID`          | The user id associated with the message.                            | No       |               |
| `delivery.appID`           | The application id associated with the message.                     | No       |               |
| `delivery.mandatory`       | Indicates if this message is mandatory.                             | No       | `false`       |
| `delivery.immediate`       | Indicates if this message should be treated as immediate.           | No       | `false`       |
| `delivery.expiration`      | Indicates the message expiration time, if any.                      | No       |               |
| `exchange.name`            | The name of the exchange to publish to.                             | No       |               |
| `exchange.type`            | The type of the exchange to publish to.                             | No       | `direct`      |
| `exchange.durable`         | Specifies whether the exchange is durable.                          | No       | `true`        |
| `exchange.autoDelete`      | If the exchange will auto-delete.                                   | No       | `false`       |
| `exchange.internal`        | If the exchange is internal.                                        | No       | `false`       |
| `exchange.noWait`          | If the exchange is declared without waiting for server reply.       | No       | `false`       |
| `routingKey`               | The routing key to use when publishing to an exchange.              | No       |               |

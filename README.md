# Conduit Connector for RabbitMQ

The RabbitMQ connector is one of [Conduit](https://github.com/ConduitIO/conduit) standalone plugins. It provides both a source and a destination connector for [RabbitMQ](https://rabbitmq.com/).

It uses the [AMQP 0-9-1 Model](https://www.rabbitmq.com/tutorials/amqp-concepts) to connect to RabbitMQ.


## What data does the OpenCDC record consist of?

| Field                   | Description                                                                             |
|-------------------------|-----------------------------------------------------------------------------------------|
| `record.Position`       | json object with the delivery tag and the queue name from where the record was read from.|
| `record.Operation`      | currently fixed as "create".                                                            |
| `record.Metadata`       | a string to string map, with keys prefixed as `rabbitmq.{DELIVERY_PROPERTY}`.           |
| `record.Key`            | the message id from the read message.                                                    |
| `record.Payload.Before` | <empty>                                                                                 |
| `record.Payload.After`  | the message body                                                                        |

## How to Build?

Run `make build` to compile the connector.

## Testing

Execute `make test` to perform all non-tls tests. Execute `make test-tls` for the TLS tests. Both commands use docker files located at `test/docker-compose.yml` and `test/docker-compose-tls.yml` respectively.
Tests require docker-compose v2.

## Source Configuration Parameters

| Name                   | Description                                                                 | Required | Default Value |
|------------------------|-----------------------------------------------------------------------------|----------|---------------|
| `url`            | The RabbitMQ server's URL.                          | Yes      |               |
| `tls.enabled` | Flag to enable or disable TLS. | false | `false` |
| `tls.clientCert` | Path to the client certificate for TLS.             | No       |               |
| `tls.clientKey`  | Path to the client's key for TLS.                   | No       |               |
| `tls.caCert`     | Path to the CA (Certificate Authority) certificate for TLS. | No   |               |
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

## Destination Configuration Parameters

| Name                   | Description                                                                 | Required | Default Value |
|------------------------|-----------------------------------------------------------------------------|----------|---------------|
| `url`            | The RabbitMQ server's URL.                          | Yes      |               |
| `tls.enabled` | Flag to enable or disable TLS. | false | `false` |
| `tls.clientCert` | Path to the client certificate for TLS.             | No       |               |
| `tls.clientKey`  | Path to the client's key for TLS.                   | No       |               |
| `tls.caCert`     | Path to the CA (Certificate Authority) certificate for TLS. | No   |               |
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


## Example pipeline.yml file

Here's an example of a `pipeline.yml` file using `file to RabbitMQ` and `RabbitMQ to file` pipelines: 

```yaml
version: 2.0
pipelines:
  - id: file-to-rabbitmq
    status: running
    connectors:
      - id: file.in
        type: source
        plugin: builtin:file
        name: file-destination
        settings:
          path: ./file.in
      - id: rabbitmq.out
        type: destination
        plugin: standalone:rabbitmq
        name: rabbitmq-source
        settings:
          url: amqp://guest:guest@localhost:5672/
          queue.name: demo-queue
          sdk.record.format: template
          sdk.record.format.options: '{{ printf "%s" .Payload.After }}'

  - id: rabbitmq-to-file
    status: running
    connectors:
      - id: rabbitmq.in
        type: source
        plugin: standalone:rabbitmq
        name: rabbitmq-source
        settings:
          url: amqp://guest:guest@localhost:5672/
          queue.name: demo-queue

      - id: file.out
        type: destination
        plugin: builtin:file
        name: file-destination
        settings:
          path: ./file.out
          sdk.record.format: template
          sdk.record.format.options: '{{ printf "%s" .Payload.After }}'
```

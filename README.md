# Conduit Connector for RabbitMQ
Integration of [Conduit](https://conduit.io) with RabbitMQ.

## How to Build?
Run `make build` to compile the connector.

## Testing
Execute `make test` to perform all non-tls tests. Execute `make test-tls` for the tls tests.

Use the Docker Compose file located at `test/docker-compose.yml` for running the required resources locally.

## Source Connector
The source connector extracts data from RabbitMQ and sends it to downstream systems via Conduit.

### Configuration Parameters

| Name                   | Description                                                  | Required | Default Value |
|------------------------|--------------------------------------------------------------|----------|---------------|
| `url`                  | The RabbitMQ server's URL.                                   | Yes      |               |
| `queueName`            | The name of the RabbitMQ queue to consume messages from.     | Yes      |               |
| `clientCert`           | Path to the client certificate for TLS.                      | No       |               |
| `clientKey`            | Path to the client's key for TLS.                            | No       |               |
| `caCert`               | Path to the CA (Certificate Authority) certificate for TLS.  | No       |               |


## Destination Connector
The destination connector sends data from upstream systems to RabbitMQ via Conduit.

### Configuration Parameters

| Name                   | Description                                                         | Required | Default Value |
|------------------------|---------------------------------------------------------------------|----------|---------------|
| `url`                  | The RabbitMQ server's URL.                                          | Yes      |               |
| `queueName`            | The name of the RabbitMQ queue where messages will be published to. | Yes      |               |
| `clientCert`           | Path to the client certificate for TLS.                             | No       |               |
| `clientKey`            | Path to the client's key for TLS.                                   | No       |               |
| `caCert`               | Path to the CA (Certificate Authority) certificate for TLS.         | No       |               |
| `contentType`          | The MIME content type of the messages written to RabbitMQ.          | No       | `text/plain`  |
| `exchangeName`         | The name of the exchange to publish to.                             | No       |               |
| `exchangeType`         | The type of the exchange to publish to.                             | No       | `direct`      |
| `routingKey`           | The routing key to use when publishing to an exchange.              | No       |               |


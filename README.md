# Conduit Connector for RabbitMQ
Integration of [Conduit](https://conduit.io) with RabbitMQ.

## How to Build?
Run `make build` to compile the connector.

## Testing
Execute `make test` to perform all tests.

Use the Docker Compose file located at `test/docker-compose.yml` for running the required resources locally.

## Source Connector
The source connector extracts data from RabbitMQ and sends it to downstream systems via Conduit.

### Configuration Parameters

| Name        | Description                                             | Required | Default Value |
|-------------|---------------------------------------------------------|----------|---------------|
| `url`       | The RabbitMQ server's URL.                              | Yes      |               |
| `queueName` | The name of the RabbitMQ queue to consume messages from.| Yes      |               |

## Destination Connector
The destination connector sends data from upstream systems to RabbitMQ via Conduit.

### Configuration Parameters

| Name          | Description                                                          | Required | Default Value |
|---------------|----------------------------------------------------------------------|----------|---------------|
| `url`         | The RabbitMQ server's URL.                                           | Yes      |               |
| `queueName`   | The name of the RabbitMQ queue where messages will be published to.  | Yes      |               |
| `contentType` | The MIME content type of the messages written to RabbitMQ.           | No       | `text/plain`  |


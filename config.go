package rabbitmq

type Config struct {
	URL string `json:"url" validate:"required"`

	QueueName string `json:"queueName" validate:"required"`
}

type SourceConfig struct {
	Config
}

type DestinationConfig struct {
	Config
}

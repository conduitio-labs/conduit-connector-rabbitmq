package rabbitmq

type Config struct {
	URL string `json:"url" validate:"required"`
}

type SourceConfig struct {
	Config
	
	QueueName string `json:"queueName" validate:"required"`
}

type DestinationConfig struct {
	Config
}

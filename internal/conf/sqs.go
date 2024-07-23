package conf

type SqsConfig struct {
	SqsQueue string `yaml:"sqs_queue" env:"SQS_QUEUE"`
	Region   string `yaml:"region" env:"SQS_REGION"`
}

func DefaultSqsConfig() SqsConfig {
	return SqsConfig{
		Region: "us-east-1",
	}
}

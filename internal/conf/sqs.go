package conf

type SqsConfig struct {
	SqsQueue string `yaml:"sqs_queue" env:"QUEUE"`
	Region   string `yaml:"region" env:"REGION"`
}

func DefaultSqsConfig() SqsConfig {
	return SqsConfig{
		Region: "us-east-1",
	}
}

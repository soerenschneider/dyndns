package hybrid

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// AwsCredentialsConfig holds optional static AWS credentials, e.g. the access
// key of an IAM user. Without them, the AWS SDK's default credential chain is
// used: environment variables, the shared config and credentials files, and
// the role of an EC2 instance, ECS task or EKS pod.
type AwsCredentialsConfig struct {
	AccessKeyId         string `yaml:"access_key_id" env:"ACCESS_KEY_ID"`
	SecretAccessKey     string `yaml:"secret_access_key" env:"SECRET_ACCESS_KEY"`
	SecretAccessKeyFile string `yaml:"secret_access_key_file" env:"SECRET_ACCESS_KEY_FILE" validate:"omitempty,filepath"`
}

func (conf *AwsCredentialsConfig) IsConfigured() bool {
	return conf.AccessKeyId != "" || conf.SecretAccessKey != "" || conf.SecretAccessKeyFile != ""
}

// CredentialsProvider returns a provider for the static credentials, or nil if
// none are configured, so the default credential chain is used.
func (conf *AwsCredentialsConfig) CredentialsProvider() (aws.CredentialsProvider, error) {
	if !conf.IsConfigured() {
		return nil, nil
	}

	secret := conf.SecretAccessKey
	if conf.SecretAccessKeyFile != "" {
		data, err := os.ReadFile(conf.SecretAccessKeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not read aws secret access key file: %w", err)
		}
		secret = strings.TrimSpace(string(data))
	}

	return credentials.NewStaticCredentialsProvider(conf.AccessKeyId, secret, ""), nil
}

// LoadAwsConfig loads the AWS config for region. A nil provider keeps the
// default credential chain.
func LoadAwsConfig(ctx context.Context, region string, provider aws.CredentialsProvider) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if provider != nil {
		opts = append(opts, config.WithCredentialsProvider(aws.NewCredentialsCache(provider)))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading aws config: %w", err)
	}

	return awsCfg, nil
}

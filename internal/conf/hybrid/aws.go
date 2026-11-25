package hybrid

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

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

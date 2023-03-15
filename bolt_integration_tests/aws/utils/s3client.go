package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

// GetS3Client returns a S3 client
func GetS3Client(t *testing.T, ctx context.Context) *s3.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           SidekickURL,
				SigningRegion: region,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(customResolver))
	require.NoError(t, err)

	// Local cli config may increase this to 5 or 6.
	// However, we want to test with the SDK default of 3.
	cfg.RetryMaxAttempts = retry.DefaultMaxAttempts

	s3c := s3.NewFromConfig(cfg, func(options *s3.Options) {
		// options.UsePathStyle = true
	})

	return s3c
}

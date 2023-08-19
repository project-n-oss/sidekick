package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

// GetSidekickS3Client returns a S3 client connected to bolt through sidekick
func GetSidekickS3Client(t *testing.T, ctx context.Context, region string) *s3.Client {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, signRegion string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           SidekickURL,
				SigningRegion: signRegion,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(customResolver))
	require.NoError(t, err)

	// Local cli config may increase this to 5 or 6.
	// However, we want to test with the SDK default of 3.
	cfg.RetryMaxAttempts = retry.DefaultMaxAttempts

	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if region == "" {
			o.Region = awsRegion(t, ctx, cfg)
		} else {
			o.Region = region
		}
		o.UsePathStyle = true
	})

	return s3c
}

// GetAwsS3Client returns a default aws S3 client
func GetAwsS3Client(t *testing.T, ctx context.Context, region string) *s3.Client {
	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(t, err)

	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if region == "" {
			o.Region = awsRegion(t, ctx, cfg)
		} else {
			o.Region = region
		}
	})
	return s3c
}

func awsRegion(t *testing.T, ctx context.Context, cfg aws.Config) string {
	client := imds.NewFromConfig(cfg)

	output, err := client.GetRegion(ctx, &imds.GetRegionInput{})
	require.NoError(t, err)
	return output.Region
}

func GetRegionForBucket(t *testing.T, ctx context.Context, bucket string) string {
	region, err := manager.GetBucketRegion(ctx, AwsS3c, bucket)
	require.NoError(t, err)
	return region
}

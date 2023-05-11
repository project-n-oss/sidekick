package boltrouter

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// awsCredentialsMap is used to cache aws credentials for a given region.
var awsCredentialsMap = sync.Map{}

// GetAwsCredentialsFromRegion returns the aws credentials for the given region.
func getAwsCredentialsFromRegion(ctx context.Context, region string) (aws.Credentials, error) {
	if awsCred, ok := awsCredentialsMap.Load(region); ok {
		return awsCred.(aws.Credentials), nil
	}

	return newAwsCredentialsFromRegion(ctx, region)
}

// newAwsCredentialsFromRegion creates a new aws credentials from the given region.
func newAwsCredentialsFromRegion(ctx context.Context, region string) (aws.Credentials, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return aws.Credentials{}, err
	}

	cred, err := awsConfig.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("could not retrieve aws credentials: %w", err)
	}
	awsCredentialsMap.Store(awsConfig.Region, cred)
	return cred, nil
}

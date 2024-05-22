package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"
)

// awsCredentialsMap is used to cache aws credentials for a given region.
var awsCredentialsMap = sync.Map{}

// GetAwsCredentialsFromRegion returns the aws credentials for the given region.
func getCredentialsFromRegion(ctx context.Context, region string) (aws.Credentials, error) {
	if awsCred, ok := awsCredentialsMap.Load(region); ok {
		return awsCred.(aws.Credentials), nil
	}
	return newCredentialsFromRegion(ctx, region)
}

// newAwsCredentialsFromRegion creates a new aws credentials from the given region.
func newCredentialsFromRegion(ctx context.Context, region string) (aws.Credentials, error) {
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

func refreshAWSCredentials(ctx context.Context, logger *zap.Logger) {
	awsCredentialsMap.Range(func(key, value interface{}) bool {
		region := key.(string)
		cred := value.(aws.Credentials)

		if cred.CanExpire {
			// if credential can expire, get new credentials for the region
			refreshedCreds, err := newCredentialsFromRegion(ctx, region)
			if err != nil {
				logger.Error(fmt.Sprintf("aws credential refresh failed for region %s", region), zap.Error(err))
				return true
			}
			awsCredentialsMap.Store(region, refreshedCreds)
		}
		return true
	})
}

func RefreshCredentialsPeriodically(ctx context.Context, logger *zap.Logger) {
	refreshAWSCredentials(ctx, logger)
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				refreshAWSCredentials(ctx, logger)
			}
		}
	}()
}

package boltrouter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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

func (br *BoltRouter) RefreshAWSCredentials(ctx context.Context) error {
	var errs []error

	awsCredentialsMap.Range(func(key, value interface{}) bool {
		region := key.(string)
		cred := value.(aws.Credentials)

		if cred.CanExpire {
			// if credential can expire, get new credentials for the region
			refreshedCreds, err := newAwsCredentialsFromRegion(ctx, region)
			if err != nil {
				errs = append(errs, fmt.Errorf("aws credential refresh failed for region %s, %w", region, err))
				return true
			}
			awsCredentialsMap.Store(region, refreshedCreds)
		}
		return true
	})

	if len(errs) > 0 {
		errStrs := make([]string, len(errs))
		for i, err := range errs {
			errStrs[i] = err.Error()
		}
		return errors.New(strings.Join(errStrs, "; "))
	}

	return nil
}

func (br *BoltRouter) RefreshAWSCredentialsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				br.RefreshAWSCredentials(ctx)
			}
		}
	}()
}

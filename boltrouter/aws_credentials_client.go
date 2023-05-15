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
var refreshRoutineScheduled bool = false

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

	if !refreshRoutineScheduled {
		refreshRoutineScheduled = true
		go scheduleRefreshAWSCredentials(ctx)
	}

	awsCredentialsMap.Store(awsConfig.Region, cred)
	return cred, nil
}

func refreshAWSCredentials(ctx context.Context) error {
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

func scheduleRefreshAWSCredentials(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, return to stop the goroutine
			return
		case <-ticker.C:
			// Get new credentials here
			if err := refreshAWSCredentials(ctx); err != nil {
				// handle error, possibly with a log message
			}
		}
	}
}

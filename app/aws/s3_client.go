package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

var s3ClientMap = sync.Map{}

func GetS3ClientFromRegion(ctx context.Context, region string) (*s3.Client, error) {
	if client, ok := s3ClientMap.Load(region); ok {
		return client.(*s3.Client), nil
	}
	return newS3ClientFromRegion(ctx, region)
}

func newS3ClientFromRegion(ctx context.Context, region string) (*s3.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig)
	return client, nil
}

func refreshS3Client(ctx context.Context, logger *zap.Logger) {
	s3ClientMap.Range(func(key, value interface{}) bool {
		region := key.(string)
		awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			logger.Error(fmt.Sprintf("failed to load aws config for region %s", region), zap.Error(err))
			return true
		}

		client := s3.NewFromConfig(awsConfig)
		s3ClientMap.Store(region, client)
		return true
	})
}

func RefreshS3ClientPeriodically(ctx context.Context, logger *zap.Logger) {
	refreshS3Client(ctx, logger)
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				refreshS3Client(ctx, logger)
			}
		}
	}()

}

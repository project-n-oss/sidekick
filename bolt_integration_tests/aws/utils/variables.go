package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	// Input params
	Bucket string
	// Region string
	// ZoneId string

	// Global state
	S3c         *s3.Client
	Port        int
	SidekickURL string
)

func InitVariables(t *testing.T, ctx context.Context) {
	Bucket = GetEnvStr("BUCKET", "sidekick-test")
	// Region = GetEnvStr("AWS_REGION", "us-east-2")
	// ZoneId = GetEnvStr("AWS_ZONE_ID", "use1-az1")

	S3c = GetS3Client(t, ctx)
}

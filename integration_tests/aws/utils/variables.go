package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

var (
	// Input params
	Bucket                   string
	FailoverBucket           string
	FailoverBucketDiffRegion string

	// Global state
	SidekickS3c *s3.Client
	AwsS3c      *s3.Client
	SidekickURL string
)

func InitVariables(t *testing.T, ctx context.Context) {
	Bucket = GetEnvStr("BUCKET", "")
	FailoverBucket = GetEnvStr("FAILOVER_BUCKET", "")
	FailoverBucketDiffRegion = GetEnvStr("FAILOVER_BUCKET_DIFF_REGION", "")
	require.NotEmpty(t, Bucket)
	SidekickS3c = GetSidekickS3Client(t, ctx, "")
	AwsS3c = GetAwsS3Client(t, ctx, "")
}

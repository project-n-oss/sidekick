package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

var (
	// Input params
	Bucket string

	// Global state
	S3c         *s3.Client
	SidekickURL string
)

func InitVariables(t *testing.T, ctx context.Context) {
	Bucket = GetEnvStr("BUCKET", "")
	require.NotEmpty(t, Bucket)
	S3c = GetS3Client(t, ctx)
}

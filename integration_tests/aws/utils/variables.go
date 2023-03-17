package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	// Input params
	Bucket string

	// Global state
	S3c         *s3.Client
	Port        int
	SidekickURL string
)

func InitVariables(t *testing.T, ctx context.Context) {
	Bucket = GetEnvStr("BUCKET", "sidekick-test")

	S3c = GetS3Client(t, ctx)
}

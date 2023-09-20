package utils

import (
	"context"
	"testing"

	"cloud.google.com/go/storage"
	integrationUtils "github.com/project-n-oss/sidekick/integration_tests/utils"
	"github.com/stretchr/testify/require"
)

var (
	// Input params
	Bucket                   string
	FailoverBucket           string
	FailoverBucketDiffRegion string

	// Global state
	SidekickGcsClient *storage.Client
	GcpGcsClient      *storage.Client
	SidekickURL       string
)

func InitVariables(t *testing.T, ctx context.Context) {
	Bucket = integrationUtils.GetEnvStr("BUCKET", "")
	FailoverBucket = integrationUtils.GetEnvStr("FAILOVER_BUCKET", "")
	FailoverBucketDiffRegion = integrationUtils.GetEnvStr("FAILOVER_BUCKET_DIFF_REGION", "")
	require.NotEmpty(t, Bucket)
	SidekickGcsClient = GetSidekickGcsClient(t, ctx)
	GcpGcsClient = GetGoogleGcsClient(t, ctx)
}

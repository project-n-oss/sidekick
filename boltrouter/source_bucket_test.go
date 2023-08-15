package boltrouter

import (
	"context"
	"fmt"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestExtractSourceBucket(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	testCase := []struct {
		requestStyle s3RequestStyle
		region       string
	}{
		{requestStyle: pathStyle, region: "us-east-1"},
		{requestStyle: pathStyle, region: "us-east-2"},
		{requestStyle: pathStyle, region: "us-west-1"},
		{requestStyle: pathStyle, region: "us-west-2"},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("%v_%v", tc.region, tc.requestStyle), func(t *testing.T) {
			bucketName := randomdata.SillyName()

			testS3Client := NewTestS3Client(t, ctx, tc.requestStyle, tc.region)
			// This populates testS3Client.req
			testS3Client.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket: aws.String(bucketName),
			})

			req := testS3Client.GetRequest(t, ctx)
			sourceBucket, err := extractSourceBucket(ctx, logger, req, "foo")
			assert.NoError(t, err)
			assert.Equal(t, bucketName, sourceBucket.Bucket)
			assert.Equal(t, tc.requestStyle, sourceBucket.Style)
			assert.Equal(t, tc.region, sourceBucket.Region)
		})
	}

	t.Run("DefaultRegionFallback", func(t *testing.T) {
		bucketName := randomdata.SillyName()

		testS3Client := NewTestS3Client(t, ctx, pathStyle, "")
		// This populates testS3Client.req
		testS3Client.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})

		req := testS3Client.GetRequest(t, ctx)
		sourceBucket, err := extractSourceBucket(ctx, logger, req, "foo")
		assert.NoError(t, err)
		assert.Equal(t, bucketName, sourceBucket.Bucket)
		assert.Equal(t, pathStyle, sourceBucket.Style)
		assert.Equal(t, "foo", sourceBucket.Region)
	})
}

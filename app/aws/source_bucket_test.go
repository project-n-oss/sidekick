package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAws_SourceBucket(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		requestStyle s3RequestStyle
		region       string
	}{
		{requestStyle: PathStyle, region: "us-east-1"},
		{requestStyle: PathStyle, region: "us-east-2"},
		{requestStyle: PathStyle, region: "us-west-1"},
		{requestStyle: PathStyle, region: "us-west-2"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v_%v", tc.region, tc.requestStyle), func(t *testing.T) {
			bucketName := randomdata.SillyName()
			testS3Client := NewTestS3Client(t, ctx, tc.requestStyle, tc.region)
			testS3Client.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket: aws.String(bucketName),
			})

			req := testS3Client.GetRequest(t, ctx)
			sourceBucket, err := ExtractSourceBucket(req)
			assert.Equal(t, bucketName, sourceBucket.Bucket)
			assert.Equal(t, tc.requestStyle, sourceBucket.Style)
			assert.Equal(t, tc.region, sourceBucket.Region)
			require.NoError(t, err)
		})
	}
}

package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *AwsSuite) TestListObjectsV2() {
	s.Run("List", s.listObjectsV2)
	s.Run("ListDiffRegion", s.listObjectsV2DiffRegion)
}

func (s *AwsSuite) listObjectsV2() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.Bucket)

	expectedKeys := testDataKeys()

	utils.AssertAwsClients(t, ctx, "ListObjectsV2",
		&s3.ListObjectsV2Input{
			Bucket: awsBucket,
		},
		func(t *testing.T, v reflect.Value) reflect.Value {
			resp := v.Interface().(*s3.ListObjectsV2Output)
			keys := make([]string, len(resp.Contents))
			for i, obj := range resp.Contents {
				keys[i] = *obj.Key
				fmt.Println(*obj.Key)
			}
			assert.ElementsMatch(t, expectedKeys, keys)
			return reflect.ValueOf(keys)
		},
	)
}

func (s *AwsSuite) listObjectsV2DiffRegion() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.FailoverBucketDiffRegion)
	region := utils.GetRegionForBucket(t, ctx, utils.FailoverBucketDiffRegion)
	s3c := utils.GetAwsS3Client(t, ctx, region)
	s3cSidekick := utils.GetSidekickS3Client(t, ctx, region)

	input := &s3.ListObjectsV2Input{
		Bucket: awsBucket,
	}

	awsResp, err := s3c.ListObjectsV2(ctx, input)
	require.NoError(t, err)
	sidekickResp, err := s3cSidekick.ListObjectsV2(ctx, input)
	require.NoError(t, err)

	require.ElementsMatch(t, awsResp.Contents, sidekickResp.Contents)
}

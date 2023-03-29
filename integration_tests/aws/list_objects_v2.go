package aws

import (
	"reflect"
	"testing"

	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *AwsSuite) TestListObjectsV2() {
	s.Run("List", s.listObjectsV2)
}

func (s *AwsSuite) listObjectsV2() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.Bucket)

	utils.AssertAwsClients(t, ctx, "ListObjectsV2", &s3.ListObjectsV2Input{Bucket: awsBucket}, func(t *testing.T, v reflect.Value) reflect.Value {
		resp := v.Interface().(*s3.ListObjectsV2Output)
		return reflect.ValueOf(resp.Contents)
	})

	// awsResp, err := utils.AwsS3c.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
	// 	Bucket: awsBucket,
	// })
	// require.NoError(t, err)
	// require.NotNil(t, awsResp)

	// boltResp, err := utils.BoltS3c.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
	// 	Bucket: awsBucket,
	// })
	// require.NoError(t, err)
	// require.NotNil(t, boltResp)

	// require.Equal(t, awsResp.Contents, boltResp.Contents)
}

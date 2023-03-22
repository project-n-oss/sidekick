package aws

import (
	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

func (s *AwsSuite) TestListObjectsV2() {
	s.Run("List", s.listObjectsV2)
}

func (s *AwsSuite) listObjectsV2() {
	t := s.T()
	awsBucket := aws.String(utils.Bucket)

	awsResp, err := utils.AwsS3c.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
		Bucket: awsBucket,
	})
	require.NoError(t, err)
	require.NotNil(t, awsResp)

	boltResp, err := utils.BoltS3c.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
		Bucket: awsBucket,
	})
	require.NoError(t, err)
	require.NotNil(t, boltResp)

	require.Equal(t, awsResp.Contents, boltResp.Contents)
}

package aws

import (
	"sidekick/bolt_integration_tests/aws/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *AwsSuite) TestListObjectsV2() {
	// s.Run("List", s.listObjectsV2)
}

func (s *AwsSuite) listObjectsV2() {
	t := s.T()
	require.True(t, true)
	awsBucket := aws.String(utils.Bucket)

	resp, err := utils.S3c.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
		Bucket: awsBucket,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	for _, o := range resp.Contents {
		t.Logf("%s", *o.Key)
	}
}

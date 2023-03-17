package aws

import (
	"sidekick/integration_tests/aws/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

func (s *AwsSuite) TestGetObject() {
	s.Run("Get", s.getObject)
}

func (s *AwsSuite) getObject() {
	t := s.T()
	require.True(t, true)
	awsBucket := aws.String(utils.Bucket)

	resp, err := utils.S3c.GetObject(s.ctx, &s3.GetObjectInput{
		Bucket: awsBucket,
		Key:    aws.String("data.csv"),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// b, err := io.ReadAll(resp.Body)
	// require.NoError(t, err)
	// t.Logf("BODY: %s", string(b))
}

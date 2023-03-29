package aws

import (
	"io"
	"reflect"
	"testing"

	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *AwsSuite) TestGetObject() {
	s.Run("Get", s.getObject)
}

func testDataKeys() []string {
	return []string{
		"animals/1.csv",
		"animals/2.csv",
		"cities/1.json",
		"cities/2.json",
	}
}

func (s *AwsSuite) getObject() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.Bucket)

	for _, key := range testDataKeys() {
		t.Run(key, func(t *testing.T) {
			utils.AssertAwsClients(t, ctx, "GetObject", &s3.GetObjectInput{
				Bucket: awsBucket,
				Key:    aws.String(key),
			},
				func(t *testing.T, v reflect.Value) reflect.Value {
					resp := v.Interface().(*s3.GetObjectOutput)
					buf, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					return reflect.ValueOf(buf)
				},
			)

			// awsResp, err := utils.AwsS3c.GetObject(s.ctx, &s3.GetObjectInput{
			// 	Bucket: awsBucket,
			// 	Key:    aws.String(key),
			// })
			// require.NoError(t, err)
			// require.NotNil(t, awsResp)
			// awsBuf, err := io.ReadAll(awsResp.Body)
			// awsResp.Body.Close()
			// require.NoError(t, err)
			// awsBody := string(awsBuf)

			// boltResp, err := utils.SidekickS3c.GetObject(s.ctx, &s3.GetObjectInput{
			// 	Bucket: awsBucket,
			// 	Key:    aws.String(key),
			// })
			// require.NoError(t, err)
			// require.NotNil(t, boltResp)
			// boltBuf, err := io.ReadAll(boltResp.Body)
			// boltResp.Body.Close()
			// require.NoError(t, err)
			// boltBody := string(boltBuf)

			// require.Equal(t, awsBody, boltBody)
		})
	}
}

package aws

import (
	"io"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *AwsSuite) TestPutObject() {
	s.Run("Put", s.putObject)
}

// func testDataKeys() []string {
// 	return []string{
// 		"animals/1.csv",
// 		"animals/2.csv",
// 		"cities/1.json",
// 		"cities/2.json",
// 	}
// }

func (s *AwsSuite) putObject() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.Bucket)
	region := utils.GetRegionForBucket(t, ctx, utils.Bucket)
	s3c := utils.GetAwsS3Client(t, ctx, region)
	s3cSidekick := utils.GetSidekickS3Client(t, ctx, region)

	data := randomdata.Paragraph()
	reader := strings.NewReader(data)

	input := &s3.PutObjectInput{
		Bucket: awsBucket,
		Key:    aws.String("foo.txt"),
		Body:   reader,
	}
	awsResp, err := s3c.PutObject(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, awsResp)
	// awsResp, err = s3cSidekick.PutObject(ctx, input)
	// require.NoError(t, err)
	// assert.NotNil(t, awsResp)
	t.Cleanup(func() {
		_, err := s3cSidekick.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: input.Bucket,
			Key:    input.Key,
		})
		require.NoError(t, err)
		// getObjResp, err := s3cSidekick.GetObject(ctx, &s3.GetObjectInput{
		// 	Bucket: input.Bucket,
		// 	Key:    input.Key,
		// })
		// require.NoError(t, err)
		// assert.NotNil(t, getObjResp)
	})

	testCases := []struct {
		name string
		s3c  *s3.Client
	}{
		{name: "sidekick", s3c: s3cSidekick},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			getObjResp, err := tt.s3c.GetObject(ctx, &s3.GetObjectInput{
				Bucket: input.Bucket,
				Key:    input.Key,
			})
			require.NoError(t, err)
			defer getObjResp.Body.Close()

			assert.NotNil(t, getObjResp)

			retData, err := io.ReadAll(getObjResp.Body)
			require.NoError(t, err)
			assert.Equal(t, data, string(retData))
		})
	}
}

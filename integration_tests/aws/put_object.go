package aws

import (
	"crypto/md5"
	_ "embed"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *AwsSuite) TestPutObject() {
	s.Run("Put", s.putObject)
	s.Run("PutWrongMD5", s.putObjectWrongMD5)
}

//go:embed put_data.txt
var putData string

func (s *AwsSuite) putObject() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.Bucket)
	awsKey := aws.String("put_data.txt")
	region := utils.GetRegionForBucket(t, ctx, utils.Bucket)
	s3cSidekick := utils.GetSidekickS3Client(t, ctx, region)

	md5Hash := md5.Sum([]byte(putData))
	md5HashStr := base64.StdEncoding.EncodeToString(md5Hash[:])

	reader := strings.NewReader(putData)
	input := &s3.PutObjectInput{
		Bucket:      awsBucket,
		Key:         awsKey,
		Body:        reader,
		ContentType: aws.String("text/plain"),
		ContentMD5:  aws.String(md5HashStr),
	}
	awsResp, err := s3cSidekick.PutObject(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, awsResp)

	t.Cleanup(func() {
		_, err := s3cSidekick.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: awsBucket,
			Key:    awsKey,
		})
		require.NoError(t, err)
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
				Bucket: awsBucket,
				Key:    awsKey,
			})
			require.NoError(t, err)
			defer getObjResp.Body.Close()

			assert.NotNil(t, getObjResp)

			retData, err := io.ReadAll(getObjResp.Body)
			require.NoError(t, err)
			assert.Equal(t, putData, string(retData))
		})
	}
}

func (s *AwsSuite) putObjectWrongMD5() {
	ctx := s.ctx
	t := s.T()
	awsBucket := aws.String(utils.Bucket)
	awsKey := aws.String("put_data.txt")
	region := utils.GetRegionForBucket(t, ctx, utils.Bucket)
	s3cSidekick := utils.GetSidekickS3Client(t, ctx, region)

	md5Hash := md5.Sum([]byte(putData + "wrong"))
	md5HashStr := base64.StdEncoding.EncodeToString(md5Hash[:])

	reader := strings.NewReader(putData)
	input := &s3.PutObjectInput{
		Bucket:      awsBucket,
		Key:         awsKey,
		Body:        reader,
		ContentType: aws.String("text/plain"),
		ContentMD5:  aws.String(md5HashStr),
	}
	awsResp, err := s3cSidekick.PutObject(ctx, input)
	require.Error(t, err)
	assert.Nil(t, awsResp)
}

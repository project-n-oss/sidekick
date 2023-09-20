package gcp

import (
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/project-n-oss/sidekick/integration_tests/gcp/utils"
	"github.com/stretchr/testify/require"
)

func (s *GcpSuite) TestGetObject() {
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

func (s *GcpSuite) getObject() {
	ctx := s.ctx
	t := s.T()

	testData := map[string][]byte{}
	for _, key := range testDataKeys() {
		r, err := os.Open(filepath.Join("test_data", key))
		require.NoError(t, err)
		buf, err := io.ReadAll(r)
		require.NoError(t, err)
		testData[key] = buf
	}

	testCases := []struct {
		name   string
		bucket string
		gscs   *storage.Client
	}{
		{name: "SidekickNormalBucket", bucket: utils.Bucket, gscs: utils.SidekickGcsClient},
		{name: "SidekickFailoverBucket", bucket: utils.FailoverBucket, gscs: utils.SidekickGcsClient},
		{name: "GcpNormalBucket", bucket: utils.Bucket, gscs: utils.GcpGcsClient},
	}

	for _, tt := range testCases {
		for _, key := range testDataKeys() {
			r, err := tt.gscs.Bucket(tt.bucket).Object(key).NewReader(ctx)
			require.NoError(t, err)
			buf, err := io.ReadAll(r)
			require.NoError(t, err)
			require.NoError(t, r.Close())
			require.NotNil(t, buf)
			require.Equal(t, testData[key], buf)
		}
	}
}

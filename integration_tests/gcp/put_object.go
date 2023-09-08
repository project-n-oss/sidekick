package gcp

import (
	_ "embed"

	"cloud.google.com/go/storage"
	"github.com/project-n-oss/sidekick/integration_tests/gcp/utils"
)

func (s *GcpSuite) TestPutObject() {
	s.Run("Put", s.putObject)
}

//go:embed put_data.txt
var putData string

func (s *GcpSuite) putObject() {
	ctx := s.ctx
	t := s.T()

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
		w := tt.gscs.Bucket(tt.bucket).Object("put_data.txt").NewWriter(ctx)
		_, err := w.Write([]byte(putData))
		if err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

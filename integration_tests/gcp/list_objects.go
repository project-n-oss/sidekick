package gcp

import (
	"cloud.google.com/go/storage"
	"github.com/project-n-oss/sidekick/integration_tests/gcp/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
)

func (s *GcpSuite) TestListObjects() {
	s.Run("List", s.listObjects)
}

func (s *GcpSuite) listObjects() {
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
		objectKeys := []string{}
		it := tt.gscs.Bucket(tt.bucket).Objects(ctx, nil)
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			objectKeys = append(objectKeys, attrs.Name)
		}
		assert.Subset(t, objectKeys, testDataKeys())
	}
}

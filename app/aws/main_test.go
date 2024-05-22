package aws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Setenv("AWS_ACCESS_KEY_ID", "foobar_key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "foobar_secret")
	exitVal := m.Run()
	os.Exit(exitVal)
}

type TestS3Client struct {
	req  *http.Request
	lock sync.RWMutex

	S3Client *s3.Client
}

func NewTestS3Client(t *testing.T, ctx context.Context, requestStyle s3RequestStyle, region string) *TestS3Client {
	ret := &TestS3Client{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ret.lock.Lock()
		ret.req = r
		ret.lock.Unlock()

		sc := http.StatusOK
		w.WriteHeader(sc)
	}))

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, signRegion string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           server.URL,
				SigningRegion: signRegion,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(customResolver),
	)
	require.NoError(t, err)
	ret.S3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		if requestStyle == PathStyle {
			o.UsePathStyle = true
		} else {
			o.UsePathStyle = false
		}
		o.Region = region
	})

	return ret
}

func (c *TestS3Client) GetRequest(t *testing.T, ctx context.Context) *http.Request {
	c.lock.RLock()
	defer c.lock.RUnlock()
	require.NotNil(t, c.req)
	return c.req.Clone(ctx)
}

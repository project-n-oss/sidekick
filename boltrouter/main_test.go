package boltrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ZONE_ID", "use1-az1")
	os.Setenv("GRANICA_CUSTOM_DOMAIN", "test.bolt.projectn.co")

	exitVal := m.Run()
	os.Exit(exitVal)
}

var (
	mainWriteEndpoints     = []interface{}{"0.0.0.1", "0.0.0.2", "0.0.0.3"}
	failoverWriteEndpoints = []interface{}{"1.0.0.1", "1.0.0.2", "1.0.0.3"}
	mainReadEndpoints      = []interface{}{"2.0.0.1", "2.0.0.2", "2.0.0.3"}
	failoverReadEndpoints  = []interface{}{"3.0.0.1", "3.0.0.2", "3.0.0.3"}

	defaultClientBehaviorParams = map[string]interface{}{
		"cleaner_on":             "true",
		"crunch_traffic_percent": "20.0",
	}

	quicksilverResponse = map[string]interface{}{
		"main_write_endpoints":     mainWriteEndpoints,
		"failover_write_endpoints": failoverWriteEndpoints,
		"main_read_endpoints":      mainReadEndpoints,
		"failover_read_endpoints":  failoverReadEndpoints,
	}
)

func NewQuicksilverMock(t *testing.T, clusterHealthy bool, clientBehaviorParams map[string]interface{}, intelligentQS bool) *httptest.Server {
	// Reset quicksilverResponse to default
	delete(quicksilverResponse, "client_behavior_params")
	delete(quicksilverResponse, "cluster_healthy")

	if intelligentQS {
		quicksilverResponse["cluster_healthy"] = clusterHealthy
		quicksilverResponse["client_behavior_params"] = defaultClientBehaviorParams
		for k, v := range clientBehaviorParams {
			quicksilverResponse["client_behavior_params"].(map[string]interface{})[k] = v
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc := http.StatusOK
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(sc)
		err := json.NewEncoder(w).Encode(quicksilverResponse)
		require.NoError(t, err)
	}))

	return server
}

func SetupQuickSilverMock(t *testing.T, ctx context.Context, clusterHealthy bool, clientBehaviorParams map[string]interface{}, intelligentQS bool, logger *zap.Logger) {
	quicksilver := NewQuicksilverMock(t, clusterHealthy, clientBehaviorParams, intelligentQS)
	boltVars, err := GetBoltVars(ctx, logger)
	require.NoError(t, err)
	boltVars.QuicksilverURL.Set(quicksilver.URL)
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
		if requestStyle == pathStyle {
			o.UsePathStyle = true
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

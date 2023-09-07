package boltrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	os.Setenv("AWS_ZONE_ID", "use1-az1")
	os.Setenv("GRANICA_CUSTOM_DOMAIN", "test.bolt.projectn.co")
	os.Setenv("GRANICA_REGION", "us-east-1")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/mock-service-account.json")

	// Mock Google Service Account data
	mockGcpSaData := map[string]interface{}{
		"type":                        "service_account",
		"project_id":                  "mock-project-id",
		"private_key_id":              "mock-private-key-id",
		"private_key":                 "-----BEGIN PRIVATE KEY-----\nMIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBALZOXYx1pQWaMBpjCYTCARVF9WrA\nk7hj7tXGbhSpBEo25ykZTBQGj0VK+5AAKmcnAewrZaGK5KlWS6eGZLHlBvgpVIfZFPa8cFSwDe3X\nZLbVWSySLJrcTtBH3XNkZXFnIHiBIFWG9JVd4/35GiAsvOKSL9UNdcpPhMi7IwvV5RDZAgMBAAEC\ngYAaSNgyDTA6y41N8KOJsZMIZyrINnXV6wqfZdmvPuMwdBQGF/ChHoT/n5z/mRaEAtrDG0qu7OCl\nDZ0gzT6ta3ECjlw+xPj8wTlVAvayo8SSToSGMiOcoi7nZ0eAn4+eIYkXpZGKDoUpa8dh3I2z4xfR\n+ldc3xotF04fSIq8CF7ckQJBANxOsvW3BkOnQxSsurh24PRdbTuAd+usjf4FKcjFFcyoESjcIqyY\ng/HCH8jzZQ6JWkcRgL1fzTW7aooXrG5gCfUCQQDT1421nSCOS3XJxTyy9V2AyAj1d/Y+DO8+d+vI\npeuSCdU+EActgGavYcokIkaY6Z8MCYQsWWPumElylNpnJKjVAkAlHJzJB6vmeaazNOW/bUc34wUj\noOCSst64i+YeDBVABI/fcjXlHUwczbbNAzNi34B1uF0XiavoAUpROOuzLDqBAkEAnCKASL5Bk38c\nlpUv0rqzqspEiB9dt4gzATjD6MQZpy5mI/MOR0Qe6t7JbO5yWBvAZM/Swhk0ZVOKts/tVR4Y7QJB\nAILwR9prGyUKn7eCN/khkfhGQggvPpbNqPe4CnTjfCBdatNyfhd4oCnvpp8Q2SWAfJt3zULGxRsa\nvYYUFMGnfAI=\n-----END PRIVATE KEY-----",
		"client_email":                "mock-service-account-email@mock-project-id.iam.gserviceaccount.com",
		"client_id":                   "mock-client-id",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/mock-service-account-email%40mock-project-id.iam.gserviceaccount.com",
		"universe_domain":             "googleapis.com",
	}

	// Convert the map to JSON
	jsonMockGcpSaData, err := json.MarshalIndent(mockGcpSaData, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	// Write the JSON data to a file
	err = ioutil.WriteFile("/tmp/mock-service-account.json", jsonMockGcpSaData, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

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
		"crunch_traffic_percent": "20",
	}

	quicksilverResponse = map[string]interface{}{
		"main_write_endpoints":     mainWriteEndpoints,
		"failover_write_endpoints": failoverWriteEndpoints,
		"main_read_endpoints":      mainReadEndpoints,
		"failover_read_endpoints":  failoverReadEndpoints,
	}
)

func NewQuicksilverMock(t *testing.T, clusterHealthy bool, clientBehaviorParams map[string]interface{}, intelligentQS bool) string {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

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

	// This URL will where sidekick will call to get the quicksilver response
	url := "https://quicksilver.us-west-2.com/services/bolt/"
	httpmock.RegisterResponder("GET", url,
		func(req *http.Request) (*http.Response, error) {
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(quicksilverResponse)
			require.NoError(t, err)

			resp := httpmock.NewStringResponse(http.StatusOK, buf.String())
			resp.Header.Set("Content-Type", "application/json")

			return resp, nil
		})

	return url
}

func SetupQuickSilverMock(t *testing.T, ctx context.Context, clusterHealthy bool, clientBehaviorParams map[string]interface{}, intelligentQS bool, logger *zap.Logger) {
	quicksilverURL := NewQuicksilverMock(t, clusterHealthy, clientBehaviorParams, intelligentQS)
	boltVars, err := GetBoltVars(ctx, logger, AwsCloudPlatform)
	require.NoError(t, err)
	boltVars.QuicksilverURL.Set(quicksilverURL)
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

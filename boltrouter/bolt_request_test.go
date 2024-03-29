package boltrouter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestBoltRequest(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, ctx, true, make(map[string]interface{}), true, logger)

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
	}{
		{name: "GetAws", cfg: Config{CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodGet},
		{name: "PutAws", cfg: Config{CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodPut},
		{name: "PostAws", cfg: Config{CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodPost},
		{name: "DeleteAws", cfg: Config{CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodDelete},
		{name: "PassthroughGetAws", cfg: Config{CloudPlatform: AwsCloudPlatform, Passthrough: true}, httpMethod: http.MethodGet},
		{name: "GetGcp", cfg: Config{CloudPlatform: GcpCloudPlatform, GcpReplicasEnabled: true}, httpMethod: http.MethodGet},
		{name: "PutGcp", cfg: Config{CloudPlatform: GcpCloudPlatform, GcpReplicasEnabled: true}, httpMethod: http.MethodPut},
		{name: "PostGcp", cfg: Config{CloudPlatform: GcpCloudPlatform, GcpReplicasEnabled: true}, httpMethod: http.MethodPost},
		{name: "DeleteGcp", cfg: Config{CloudPlatform: GcpCloudPlatform, GcpReplicasEnabled: true}, httpMethod: http.MethodDelete},
		{name: "PassthroughGetGcp", cfg: Config{CloudPlatform: GcpCloudPlatform, GcpReplicasEnabled: true, Passthrough: true}, httpMethod: http.MethodGet},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			require.NoError(t, err)
			require.NoError(t, br.RefreshBoltInfo(ctx))
			boltVars := br.boltVars
			body := strings.NewReader(randomdata.Paragraph())

			req, err := http.NewRequest(tt.httpMethod, "test.projectn.co", body)
			require.NoError(t, err)

			if tt.cfg.CloudPlatform == AwsCloudPlatform {
				req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIA3Y7DLM2EYWSYCN5P/20230511/us-west-2/s3/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=6447287d46d333a010e224191d64c31b9738cc37886aadb7753a0a579a30edc6")

			}
			if tt.cfg.CloudPlatform == GcpCloudPlatform {
				req.Header.Set("Authorization", "Bearer ya29.c.b0Aaekm1J-kvFPrZl9fpH2Yw")
			}

			boltReq, err := br.NewBoltRequest(ctx, logger, req)
			require.NoError(t, err)
			require.NotNil(t, boltReq)
			boltHttpRequest := boltReq.Bolt
			require.Equal(t, req.Method, boltHttpRequest.Method)

			endpointOrder := br.getPreferredEndpointOrder(req.Method)
			endpoints := boltVars.BoltInfo.Get()[endpointOrder[0]]
			require.Contains(t, endpoints, boltHttpRequest.URL.Hostname())

			if tt.cfg.CloudPlatform == AwsCloudPlatform {
				require.Equal(t, boltVars.BoltHostname.Get(), boltHttpRequest.Header.Get("Host"))
				require.NotEmpty(t, boltHttpRequest.Header.Get("X-Amz-Date"))
				require.NotEmpty(t, boltHttpRequest.Header.Get("Authorization"))
				require.NotEmpty(t, boltHttpRequest.Header.Get("X-Bolt-Auth-Prefix"))
				require.Len(t, boltHttpRequest.Header.Get("X-Bolt-Auth-Prefix"), 4)
				require.Contains(t, boltHttpRequest.Header.Get("User-Agent"), boltVars.UserAgentPrefix.Get())
			}

			if tt.cfg.CloudPlatform == GcpCloudPlatform {
				require.Equal(t, boltVars.BoltHostname.Get(), boltHttpRequest.Header.Get("Host"))
				require.Contains(t, boltHttpRequest.Header.Get("User-Agent"), boltVars.UserAgentPrefix.Get())
			}

			passthrough := boltHttpRequest.Header.Get("X-Bolt-Passthrough-Read")
			if !tt.cfg.Passthrough {
				require.Equal(t, "disable", passthrough)
			} else {
				require.Equal(t, "", passthrough)
			}
		})
	}
}

// function to test failover handling behavior in BoltRouter.DoRequest
func TestBoltRequestFailover(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	allReadEndpoints := append(mainReadEndpoints, failoverReadEndpoints...)
	// register error responders for all read endpoints
	for _, endpoint := range allReadEndpoints {
		httpmock.RegisterResponder("GET", fmt.Sprintf("https://%s/test.granica.ai123456", endpoint),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, "SERVER ERROR"), fmt.Errorf("s3 error")
			})
	}
	httpmock.RegisterResponder("GET", "https://bolt.s3.us-west-2.amazonaws.com/test.granica.ai123456",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "OK"), nil
		})
	httpmock.RegisterResponder("GET", "https://storage.googleapis.com/test.granica.ai123456",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "OK"), nil
		})
	httpmock.RegisterResponder("POST", "https://oauth2.googleapis.com/token",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "{\"token_type\":\"bearer\",\"access_token\":\"AAAA%2FAAA%3DAAAAAAAAxxxxxx\"}"), nil
		})

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
	}{
		{name: "AwsFailover", cfg: Config{CloudPlatform: AwsCloudPlatform, Failover: false}, httpMethod: http.MethodGet},
		{name: "GcpFailover", cfg: Config{CloudPlatform: GcpCloudPlatform, Failover: false}, httpMethod: http.MethodGet},
	}
	for _, tt := range testCases {
		ctx := context.Background()
		logger := zaptest.NewLogger(t)
		SetupQuickSilverMock(t, ctx, true, make(map[string]interface{}), true, logger)
		br, err := NewBoltRouter(ctx, logger, tt.cfg)
		require.NoError(t, err)
		require.NoError(t, br.RefreshBoltInfo(ctx))
		overrideCrunchTrafficPct(br, "100")

		body := strings.NewReader(randomdata.Paragraph())
		req, err := http.NewRequest(http.MethodGet, "test.granica.ai123456", body)
		require.NoError(t, err)

		if tt.cfg.CloudPlatform == AwsCloudPlatform {
			req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIA3Y7DLM2EYWSYCN5P/20230511/us-west-2/s3/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=6447287d46d333a010e224191d64c31b9738cc37886aadb7753a0a579a30edc6")
		}
		if tt.cfg.CloudPlatform == GcpCloudPlatform {
			req.Header.Set("Authorization", "Bearer ya29.c.b0Aaekm1J-kvFPrZl9fpH2Yw")
		}

		boltReq, err := br.NewBoltRequest(ctx, logger, req)
		require.NoError(t, err)
		require.NotNil(t, boltReq)
		_, failover, _, err := br.DoRequest(logger, boltReq)

		require.Error(t, err, failover)
		require.False(t, failover, err)

		br.config.Failover = true
		boltReq, err = br.NewBoltRequest(ctx, logger, req)
		require.NoError(t, err)
		require.NotNil(t, boltReq)

		_, failover, analytics, err := br.DoRequest(logger, boltReq)
		// failover is enabled, so we should get a successful response by failing over to s3
		require.NoError(t, err)
		require.True(t, failover, err)
		require.NotZero(t, analytics.AwsRequestResponseStatusCode)
	}
}

// function to test panic handling behavior in BoltRouter.DoRequest
func TestBoltRequestPanic(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	allReadEndpoints := append(mainReadEndpoints, failoverReadEndpoints...)
	// register panic responders for all read endpoints
	for _, endpoint := range allReadEndpoints {
		httpmock.RegisterResponder("GET", fmt.Sprintf("https://%s/test.granica.ai123456", endpoint),
			func(req *http.Request) (*http.Response, error) {
				panic("Simulated panic during request")
			})
		httpmock.RegisterResponder("GET", fmt.Sprintf("https://%s:8443/test.granica.ai123456", endpoint),
			func(req *http.Request) (*http.Response, error) {
				panic("Simulated panic during request")
			})
	}

	httpmock.RegisterResponder("GET", "https://bolt.s3.us-west-2.amazonaws.com/test.granica.ai123456",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "OK"), nil
		})
	httpmock.RegisterResponder("GET", "https://storage.googleapis.com/test.granica.ai123456",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "OK"), nil
		})
	httpmock.RegisterResponder("POST", "https://oauth2.googleapis.com/token",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "{\"token_type\":\"bearer\",\"access_token\":\"AAAA%2FAAA%3DAAAAAAAAxxxxxx\"}"), nil
		})

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
	}{
		{name: "AwsPanic", cfg: Config{CloudPlatform: AwsCloudPlatform, Failover: false}, httpMethod: http.MethodGet},
		{name: "GcpPanic", cfg: Config{CloudPlatform: GcpCloudPlatform, Failover: false, GcpReplicasEnabled: true}, httpMethod: http.MethodGet},
	}

	for _, tt := range testCases {
		ctx := context.Background()
		logger := zaptest.NewLogger(t)
		SetupQuickSilverMock(t, ctx, true, make(map[string]interface{}), true, logger)

		br, err := NewBoltRouter(ctx, logger, tt.cfg)
		require.NoError(t, err)
		require.NoError(t, br.RefreshBoltInfo(ctx))

		overrideCrunchTrafficPct(br, "100")

		body := strings.NewReader(randomdata.Paragraph())
		req, err := http.NewRequest(http.MethodGet, "test.granica.ai123456", body)
		require.NoError(t, err)

		if tt.cfg.CloudPlatform == AwsCloudPlatform {
			req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIA3Y7DLM2EYWSYCN5P/20230511/us-west-2/s3/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=6447287d46d333a010e224191d64c31b9738cc37886aadb7753a0a579a30edc6")
		}
		if tt.cfg.CloudPlatform == GcpCloudPlatform {
			req.Header.Set("Authorization", "Bearer ya29.c.b0Aaekm1J-kvFPrZl9fpH2Yw")
		}

		boltReq, err := br.NewBoltRequest(ctx, logger, req)
		require.NoError(t, err)
		require.NotNil(t, boltReq)

		_, _, _, err = br.DoRequest(logger, boltReq)
		require.Error(t, err)
		// no failover in config so we should get an error
		require.Contains(t, err.Error(), "panic")

		br.config.Failover = true
		boltReq, err = br.NewBoltRequest(ctx, logger, req)
		require.NoError(t, err)
		require.NotNil(t, boltReq)

		_, _, _, err = br.DoRequest(logger, boltReq)
		// failover is enabled, so we should get a successful response by failing over to s3/gcs
		require.NoError(t, err)
	}
}

func overrideCrunchTrafficPct(br *BoltRouter, pct string) {
	m := br.boltVars.BoltInfo.Get()
	m["client_behavior_params"].(map[string]interface{})["crunch_traffic_percent"] = pct
}

// function to test endpoint offline behavior in BoltRouter.DoRequest
func TestBoltEndpoint(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, ctx, true, make(map[string]interface{}), true, logger)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	allReadEndpoints := append(mainReadEndpoints, failoverReadEndpoints...)
	// register panic responders for all read endpoints
	for _, endpoint := range allReadEndpoints {
		httpmock.RegisterResponder("GET", fmt.Sprintf("https://%s/test.projectn.co", endpoint),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(500, "SERVER ERROR"), fmt.Errorf("s3 error")
			})
	}

	httpmock.RegisterResponder("GET", "https://bolt.s3.us-west-2.amazonaws.com/test.projectn.co",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "OK"), nil
		})
	httpmock.RegisterResponder("GET", "https://storage.googleapis.com/test.granica.ai123456",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "OK"), nil
		})
	httpmock.RegisterResponder("POST", "https://oauth2.googleapis.com/token",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "{\"token_type\":\"bearer\",\"access_token\":\"AAAA%2FAAA%3DAAAAAAAAxxxxxx\"}"), nil
		})

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
	}{
		{name: "Aws", cfg: Config{CloudPlatform: AwsCloudPlatform, Failover: false}, httpMethod: http.MethodGet},
		{name: "Gcp", cfg: Config{CloudPlatform: GcpCloudPlatform, Failover: false, GcpReplicasEnabled: true}, httpMethod: http.MethodGet},
	}

	for _, tt := range testCases {
		br, err := NewBoltRouter(ctx, logger, tt.cfg)
		require.NoError(t, err)
		require.NoError(t, br.RefreshBoltInfo(ctx))

		overrideCrunchTrafficPct(br, "100")

		doRequest := func() (*BoltRequestAnalytics, error) {
			body := strings.NewReader(randomdata.Paragraph())
			req, err := http.NewRequest(tt.httpMethod, "test.projectn.co", body)
			require.NoError(t, err)

			if tt.cfg.CloudPlatform == AwsCloudPlatform {
				req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIA3Y7DLM2EYWSYCN5P/20230511/us-west-2/s3/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=6447287d46d333a010e224191d64c31b9738cc37886aadb7753a0a579a30edc6")
			}
			if tt.cfg.CloudPlatform == GcpCloudPlatform {
				req.Header.Set("Authorization", "Bearer ya29.c.b0Aaekm1J-kvFPrZl9fpH2Yw")
			}

			boltReq, err := br.NewBoltRequest(ctx, logger, req)
			if err != nil {
				return nil, err
			}
			require.NotNil(t, boltReq)

			_, _, analytics, err := br.DoRequest(logger, boltReq)
			require.Error(t, err)

			return analytics, nil
		}

		for ii := 0; ii < 20; ii++ {
			analytics, err := doRequest()
			// Ensure on error, we have tried many endpoints already
			if err != nil {
				require.Greater(t, ii, 0)
				break
			}
			logger.Info(fmt.Sprintf("resp %+v", analytics))
		}
		// On refresh, all endpoints will become live again
		require.NoError(t, br.RefreshBoltInfo(ctx))
		analytics, err := doRequest()
		require.NoError(t, err)
		logger.Info(fmt.Sprintf("resp %+v", analytics))
	}
}

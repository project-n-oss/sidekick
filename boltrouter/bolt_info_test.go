package boltrouter

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetBoltInfo(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, ctx, true, map[string]interface{}{"crunch_traffic_percent": "50"}, true, logger)

	testCases := []struct {
		name     string
		cfg      Config
		expected map[string]interface{}
	}{
		{name: "NonLocal", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform}, expected: quicksilverResponse},
		{name: "Local", cfg: Config{Local: true, CloudPlatform: AwsCloudPlatform}, expected: map[string]interface{}{}},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			require.NoError(t, err)

			info, err := br.getBoltInfo(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.expected, map[string]interface{}(info))
		})
	}
}

func TestSelectBoltEndpoint(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, ctx, true, map[string]interface{}{"crunch_traffic_percent": "60"}, true, logger)

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
		expected   []interface{}
	}{
		{name: "NonLocalGet", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodGet, expected: mainReadEndpoints},
		{name: "NonLocalHead", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodHead, expected: mainReadEndpoints},

		{name: "NonLocalPut", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodPut, expected: mainWriteEndpoints},
		{name: "NonLocalPost", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodPost, expected: mainWriteEndpoints},
		{name: "NonLocalDelete", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform}, httpMethod: http.MethodDelete, expected: mainWriteEndpoints},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			require.NoError(t, err)
			err = br.RefreshBoltInfo(ctx)
			require.NoError(t, err)

			endpoint, err := br.SelectBoltEndpoint(tt.httpMethod)
			require.NoError(t, err)
			require.Contains(t, tt.expected, endpoint.Hostname())
		})
	}
}

func TestSelectInitialRequestTarget(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		URL                  string
		name                 string
		cfg                  Config
		clusterHealthy       bool
		clientBehaviorParams map[string]interface{}
		expected             int
		reason               string
		intelligentQS        bool
	}{
		// Check with CrunchTrafficSplitByObjectKeyHash
		{URL: "pqr.txt", name: "ClusterUnhealthy", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: false, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "70"}, expected: InitialRequestTargetFallback, reason: "cluster unhealthy", intelligentQS: true},
		{URL: "pqr.txt", name: "ClusterHealthyCrunchTrafficZeroPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "0"}, expected: InitialRequestTargetFallback, reason: "traffic splitting", intelligentQS: true},
		{URL: "pqr.txt", name: "ClusterHealthyCrunchTrafficHundredPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "100"}, expected: InitialRequestTargetBolt, reason: "traffic splitting", intelligentQS: true},
		{URL: "pqr.txt", name: "BackwardsCompat", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: false, clientBehaviorParams: map[string]interface{}{}, expected: InitialRequestTargetBolt, reason: "backwards compatibility", intelligentQS: false},
		{URL: "xyz123.txt", name: "ClusterHealthyCrunchTrafficFiftyPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "50"}, expected: InitialRequestTargetFallback, reason: "traffic splitting", intelligentQS: true},
		{URL: "abc123.txt", name: "ClusterHealthyCrunchTrafficFiftyPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "50"}, expected: InitialRequestTargetBolt, reason: "traffic splitting", intelligentQS: true},
		{URL: "abc/abc123.txt", name: "ClusterHealthyCrunchTrafficFiftyPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByObjectKeyHash}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "50"}, expected: InitialRequestTargetFallback, reason: "traffic splitting", intelligentQS: true},
		// Check with CrunchTrafficSplitByRandomRequest
		{URL: "pqr.txt", name: "ClusterUnhealthy", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByRandomRequest}, clusterHealthy: false, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "70"}, expected: InitialRequestTargetFallback, reason: "cluster unhealthy", intelligentQS: true},
		{URL: "pqr.txt", name: "ClusterHealthyCrunchTrafficZeroPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByRandomRequest}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "0"}, expected: InitialRequestTargetFallback, reason: "traffic splitting", intelligentQS: true},
		{URL: "pqr.txt", name: "ClusterHealthyCrunchTrafficHundredPercent", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByRandomRequest}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": "100"}, expected: InitialRequestTargetBolt, reason: "traffic splitting", intelligentQS: true},
		{URL: "pqr.txt", name: "BackwardsCompat", cfg: Config{Local: false, CloudPlatform: AwsCloudPlatform, CrunchTrafficSplit: CrunchTrafficSplitByRandomRequest}, clusterHealthy: false, clientBehaviorParams: map[string]interface{}{}, expected: InitialRequestTargetBolt, reason: "backwards compatibility", intelligentQS: false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			SetupQuickSilverMock(t, ctx, tt.clusterHealthy, tt.clientBehaviorParams, tt.intelligentQS, logger)
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			require.NoError(t, err)
			err = br.RefreshBoltInfo(ctx)
			require.NoError(t, err)
			body := strings.NewReader(randomdata.Paragraph())
			req, err := http.NewRequest(http.MethodGet, tt.URL, body)
			req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIA3Y7DLM2EYWSYCN5P/20230511/us-west-2/s3/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=6447287d46d333a010e224191d64c31b9738cc37886aadb7753a0a579a30edc6")
			boltReq, err := br.NewBoltRequest(ctx, logger, req)
			target, reason, err := br.SelectInitialRequestTarget(boltReq)
			require.NoError(t, err)
			require.Equal(t, tt.expected, target)
			require.Equal(t, tt.reason, reason)
		})
	}
}

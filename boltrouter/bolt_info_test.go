package boltrouter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetBoltInfo(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, ctx, true, make(map[string]interface{}), logger)

	testCases := []struct {
		name     string
		cfg      Config
		expected map[string]interface{}
	}{
		{name: "NonLocal", cfg: Config{Local: false}, expected: quicksilverResponse},
		{name: "Local", cfg: Config{Local: true}, expected: map[string]interface{}{}},
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
	SetupQuickSilverMock(t, ctx, true, make(map[string]interface{}), logger)

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
		expected   []string
	}{
		{name: "NonLocalGet", cfg: Config{Local: false}, httpMethod: http.MethodGet, expected: mainReadEndpoints},
		{name: "NonLocalHead", cfg: Config{Local: false}, httpMethod: http.MethodHead, expected: mainReadEndpoints},

		{name: "NonLocalPut", cfg: Config{Local: false}, httpMethod: http.MethodPut, expected: mainWriteEndpoints},
		{name: "NonLocalPost", cfg: Config{Local: false}, httpMethod: http.MethodPost, expected: mainWriteEndpoints},
		{name: "NonLocalDelete", cfg: Config{Local: false}, httpMethod: http.MethodDelete, expected: mainWriteEndpoints},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			br.RefreshBoltInfo(ctx)
			require.NoError(t, err)

			endpoint, err := br.SelectBoltEndpoint(ctx, tt.httpMethod)
			require.NoError(t, err)
			require.Contains(t, tt.expected, endpoint.Hostname())
		})
	}
}

func TestSelectInitialRequestTarget(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	testCases := []struct {
		name                 string
		cfg                  Config
		clusterHealthy       bool
		clientBehaviorParams map[string]interface{}
		expected             string
		reason               string
	}{
		{name: "ClusterUnhealthy", cfg: Config{Local: false}, clusterHealthy: false, clientBehaviorParams: map[string]interface{}{}, expected: "s3", reason: "cluster unhealthy"},
		{name: "ClusterHealthyCrunchTrafficZeroPercent", cfg: Config{Local: true}, clusterHealthy: false, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": 0}, expected: "s3", reason: "traffic splitting"},
		{name: "ClusterHealthyCrunchTrafficHundredPercent", cfg: Config{Local: true}, clusterHealthy: true, clientBehaviorParams: map[string]interface{}{"crunch_traffic_percent": 100}, expected: "bolt", reason: "traffic splitting"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			SetupQuickSilverMock(t, ctx, tt.clusterHealthy, tt.clientBehaviorParams, logger)
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			require.NoError(t, err)

			target, reason, err := br.SelectInitialRequestTarget()
			require.NoError(t, err)
			require.Equal(t, tt.expected, target)
			require.Equal(t, tt.reason, reason)
		})
	}
}

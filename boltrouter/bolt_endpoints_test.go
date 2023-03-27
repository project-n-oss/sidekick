package boltrouter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetBoltEndpoints(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, logger)

	testCases := []struct {
		name     string
		cfg      Config
		expected map[string][]string
	}{
		{name: "NonLocal", cfg: Config{Local: false}, expected: quicksilverResponse},
		{name: "Local", cfg: Config{Local: true}, expected: map[string][]string{}},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			require.NoError(t, err)

			endpoints, err := br.getBoltEndpoints(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.expected, map[string][]string(endpoints))
		})
	}
}

func TestSelectBoltEndpoint(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, logger)

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
			br.RefreshEndpoints(ctx)
			require.NoError(t, err)

			endpoint, err := br.SelectBoltEndpoint(ctx, tt.httpMethod)
			require.NoError(t, err)
			require.Contains(t, tt.expected, endpoint.Hostname())
		})
	}
}

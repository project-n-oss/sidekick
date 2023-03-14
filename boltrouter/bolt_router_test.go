package boltrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var (
	mainWriteEndpoints     []string = []string{"0.0.0.1", "0.0.0.2", "0.0.0.3"}
	failoverWriteEndpoints []string = []string{"1.0.0.1", "1.0.0.2", "1.0.0.3"}
	mainReadEndpoints      []string = []string{"2.0.0.1", "2.0.0.2", "2.0.0.3"}
	failoverReadEndpoints  []string = []string{"3.0.0.1", "3.0.0.2", "3.0.0.3"}

	quicksilverResponse map[string][]string = map[string][]string{
		"main_write_endpoints":     mainWriteEndpoints,
		"failover_write_endpoints": failoverWriteEndpoints,
		"main_read_endpoints":      mainReadEndpoints,
		"failover_read_endpoints":  failoverReadEndpoints,
	}
)

func QuicksilverMock(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc := http.StatusOK

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(sc)
		json.NewEncoder(w).Encode(quicksilverResponse)
	}))

	return server
}

func TestGetBoltEndpoints(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	quicksilver := QuicksilverMock(t)
	boltVars, err := GetBoltVars(logger)
	require.NoError(t, err)
	boltVars.QuicksilverURL.Set(quicksilver.URL)

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
	quicksilver := QuicksilverMock(t)
	boltVars, err := GetBoltVars(logger)
	require.NoError(t, err)
	boltVars.QuicksilverURL.Set(quicksilver.URL)

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

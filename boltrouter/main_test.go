package boltrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ZONE_ID", "use1-az1")
	os.Setenv("BOLT_CUSTOM_DOMAIN", "test.bolt.projectn.co")

	exitVal := m.Run()
	os.Exit(exitVal)
}

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

func SetupQuickSilverMock(t *testing.T, ctx context.Context, logger *zap.Logger) {
	quicksilver := QuicksilverMock(t)
	boltVars, err := GetBoltVars(ctx, logger)
	require.NoError(t, err)
	boltVars.QuicksilverURL.Set(quicksilver.URL)
}

package boltrouter

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestBoltRequest(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, logger)

	testCases := []struct {
		name       string
		cfg        Config
		httpMethod string
	}{
		{name: "Get", cfg: Config{}, httpMethod: http.MethodGet},
		{name: "Put", cfg: Config{}, httpMethod: http.MethodPut},
		{name: "Post", cfg: Config{}, httpMethod: http.MethodPost},
		{name: "Delete", cfg: Config{}, httpMethod: http.MethodDelete},
		{name: "PassthroughGet", cfg: Config{Passthrough: true}, httpMethod: http.MethodGet},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			br, err := NewBoltRouter(ctx, logger, tt.cfg)
			br.RefreshEndpoints(ctx)
			boltVars := br.boltVars
			require.NoError(t, err)

			req, err := http.NewRequest(tt.httpMethod, "test.projectn.co", nil)
			require.NoError(t, err)

			boltReq, err := br.NewBoltRequest(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, boltReq)

			require.Equal(t, req.Method, boltReq.Method)

			endpointOrder := br.getPreferredEndpointOrder(req.Method)
			endpoints := boltVars.BoltEndpoints.Get()[endpointOrder[0]]
			require.Contains(t, endpoints, boltReq.URL.Hostname())

			require.Equal(t, boltVars.BoltHostname.Get(), boltReq.Header.Get("Host"))
			require.NotEmpty(t, boltReq.Header.Get("X-Amz-Date"))
			require.NotEmpty(t, boltReq.Header.Get("Authorization"))
			require.NotEmpty(t, boltReq.Header.Get("X-Bolt-Auth-Prefix"))
			require.Len(t, boltReq.Header.Get("X-Bolt-Auth-Prefix"), 4)
			require.NotEmpty(t, boltReq.Header.Get("User-Agent"))
			require.Contains(t, boltReq.Header.Get("User-Agent"), boltVars.UserAgentPrefix.Get())

			passthrough := boltReq.Header.Get("X-Bolt-Passthrough-Read")
			if !tt.cfg.Passthrough {
				require.Equal(t, "disable", passthrough)
			} else {
				require.Equal(t, "", passthrough)
			}
		})
	}

}

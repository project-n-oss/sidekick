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

func TestBoltRequest(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	SetupQuickSilverMock(t, ctx, logger)

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
			require.NoError(t, err)
			br.RefreshBoltInfo(ctx)
			boltVars := br.boltVars
			body := strings.NewReader(randomdata.Paragraph())

			req, err := http.NewRequest(tt.httpMethod, "test.projectn.co", body)
			req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=AKIA3Y7DLM2EYWSYCN5P/20230511/us-west-2/s3/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;host;x-amz-content-sha256;x-amz-date, Signature=6447287d46d333a010e224191d64c31b9738cc37886aadb7753a0a579a30edc6")
			require.NoError(t, err)

			boltReq, err := br.NewBoltRequest(ctx, logger, req)
			require.NoError(t, err)
			require.NotNil(t, boltReq)
			boltHttpRequest := boltReq.Bolt
			require.Equal(t, req.Method, boltHttpRequest.Method)

			endpointOrder := br.getPreferredEndpointOrder(req.Method)
			endpoints := boltVars.BoltInfo.Get()[endpointOrder[0]]
			require.Contains(t, endpoints, boltHttpRequest.URL.Hostname())

			require.Equal(t, boltVars.BoltHostname.Get(), boltHttpRequest.Header.Get("Host"))
			require.NotEmpty(t, boltHttpRequest.Header.Get("X-Amz-Date"))
			require.NotEmpty(t, boltHttpRequest.Header.Get("Authorization"))
			require.NotEmpty(t, boltHttpRequest.Header.Get("X-Bolt-Auth-Prefix"))
			require.Len(t, boltHttpRequest.Header.Get("X-Bolt-Auth-Prefix"), 4)
			require.Contains(t, boltHttpRequest.Header.Get("User-Agent"), boltVars.UserAgentPrefix.Get())

			passthrough := boltHttpRequest.Header.Get("X-Bolt-Passthrough-Read")
			if !tt.cfg.Passthrough {
				require.Equal(t, "disable", passthrough)
			} else {
				require.Equal(t, "", passthrough)
			}
		})
	}
}

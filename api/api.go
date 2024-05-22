package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/project-n-oss/sidekick/app"
	"go.uber.org/zap"
)

type Api struct {
	app *app.App
}

func New(ctx context.Context, cfg Config, app *app.App) (*Api, error) {
	return &Api{
		app: app,
	}, nil
}

// CreateHandler creates the http.Handler for the sidekick api
func (api *Api) CreateHandler() http.Handler {
	handler := http.HandlerFunc(api.routeBase)
	handler = api.healthMiddleware(handler, api.app.Health)
	handler = api.sessionMiddleware(handler)

	return handler
}

func (api *Api) routeBase(w http.ResponseWriter, req *http.Request) {
	sess := CtxSession(req.Context())
	resp, crunched, err := sess.DoRequest(req)
	if sess.Logger().Level() == zap.DebugLevel {
		dumpRequest(sess.Logger(), req)
	}

	if err != nil {
		api.InternalError(sess.Logger(), w, err)
		return
	}
	sess.WithLogger(sess.Logger().With(
		zap.Int("awsStatusCode", resp.StatusCode),
		zap.Bool("crunched", crunched),
	))

	// Convert the response headers to lower case, as Python etc libraries expect lower case.
	for k, values := range resp.Header {
		lowK := strings.ToLower(k)
		if strings.HasPrefix(lowK, "x-amz-meta") {
			w.Header()[lowK] = values
		} else {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		api.InternalError(sess.Logger(), w, err)
		return
	}
}

func dumpRequest(logger *zap.Logger, req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		logger.Error("dumping request", zap.Error(err))
		return
	}

	logger.Debug("Request dump", zap.String("dump", string(dump)))
}

package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/project-n-oss/sidekick/boltrouter"

	"go.uber.org/zap"
)

type Api struct {
	logger *zap.Logger
	config Config

	br *boltrouter.BoltRouter
}

func New(ctx context.Context, logger *zap.Logger, cfg Config) (*Api, error) {
	br, err := boltrouter.NewBoltRouter(ctx, logger, cfg.BoltRouter)
	if err != nil {
		return nil, err
	}

	if cfg.BoltRouter.Local {
		logger.Info("running sidekick locally")
	}

	// force refresh endpoints at the start
	if err := br.RefreshEndpoints(ctx); err != nil {
		return nil, err
	}
	// Refresh endpoints periodically
	br.RefreshEndpointsPeriodically(ctx)

	return &Api{
		logger: logger,
		config: cfg,

		br: br,
	}, nil
}

// CreateHandler creates the http.Handler for the sidekick api
func (a *Api) CreateHandler() http.Handler {
	handler := http.HandlerFunc(a.routeBase)
	handler = a.sessionMiddleware(handler)

	return handler
}

func (a *Api) routeBase(w http.ResponseWriter, req *http.Request) {
	sess := CtxSession(req.Context())
	ctx := req.Context()

	boltReq, err := sess.br.NewBoltRequest(ctx, req.Clone(ctx))
	if err != nil {
		a.InternalError(sess.Logger(), w, err)
		return
	}

	if sess.Logger().Level() == zap.DebugLevel {
		dumpRequest(sess.Logger(), boltReq)
	}

	resp, failover, err := sess.br.DoBoltRequest(sess.Logger(), boltReq)
	if err != nil {
		a.InternalError(sess.Logger(), w, err)
		return
	} else if failover {
		sess.WithLogger(sess.Logger().With(zap.Bool("failover", true)))
	}

	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	if resp.StatusCode <= 200 || resp.StatusCode >= 300 {
		sess.Logger().Warn("Status code is not 2xx in aws response", zap.Int("statusCode", resp.StatusCode))
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		a.InternalError(sess.Logger(), w, err)
		return

	}
}

func dumpRequest(logger *zap.Logger, boltReq *boltrouter.BoltRequest) {
	boltDump, err := httputil.DumpRequest(boltReq.Bolt, true)
	if err != nil {
		logger.Error("bolt dump request", zap.Error(err))
		return
	}

	awsDump, err := httputil.DumpRequest(boltReq.Aws, true)
	if err != nil {
		logger.Error("aws dump request", zap.Error(err))
		return
	}

	logger.Debug("BoltRequest dump", zap.String("bolt", string(boltDump)), zap.String("aws", string(awsDump)))
}

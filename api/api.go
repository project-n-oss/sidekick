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

func (a *Api) CreateHandler() http.Handler {
	handler := http.HandlerFunc(a.routeBase)
	return handler
}

func (a *Api) routeBase(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	boltReq, err := a.br.NewBoltRequest(ctx, req.Clone(ctx))
	if err != nil {
		a.InternalError(w, err)
		return
	}

	if a.logger.Level() == zap.DebugLevel {
		dumpRequest(a.logger, boltReq)
	}

	resp, err := a.br.DoBoltRequest(boltReq)
	if err != nil {
		a.InternalError(w, err)
		return
	}

	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		a.InternalError(w, err)
		return

	}
}

func dumpRequest(logger *zap.Logger, boltReq *boltrouter.BoltRequest) {
	boltDump, err := httputil.DumpRequest(boltReq.Bolt, true)
	if err != nil {
		zap.Error(err)
		return
	}

	awsDump, err := httputil.DumpRequest(boltReq.Aws, true)
	if err != nil {
		zap.Error(err)
		return
	}

	logger.Debug("request dump", zap.String("bolt", string(boltDump)), zap.String("aws", string(awsDump)))
}

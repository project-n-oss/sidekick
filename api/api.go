package api

import (
	"context"
	"net/http"
	"sidekick/boltrouter"

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

	resp, err := a.br.DoBoltRequest(boltReq)
	if err != nil {
		a.InternalError(w, err)
		return
	}
	if err := resp.Write(w); err != nil {
		a.InternalError(w, err)
		return
	}
}

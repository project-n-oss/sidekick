package boltrouter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// BoltRouter is used to find bolt endpoints and route a AWS call to the right endpoint.
type BoltRouter struct {
	config Config

	boltHttpClient     *http.Client
	standardHttpClient *http.Client
	boltVars           *BoltVars
}

// NewBoltRouter creates a new BoltRouter.
func NewBoltRouter(ctx context.Context, logger *zap.Logger, cfg Config) (*BoltRouter, error) {
	boltVars, err := GetBoltVars(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("could not get BoltVars: %w", err)
	}

	boltHttpClient := http.Client{
		Timeout: time.Duration(90) * time.Second,
	}

	// custom transport is needed to allow certificate validation from bolt hostname
	if tp, ok := http.DefaultTransport.(*http.Transport); ok {
		customTransport := tp.Clone()
		customTransport.TLSClientConfig = &tls.Config{
			ServerName: boltVars.BoltHostname.Get(),
		}
		boltHttpClient.Transport = customTransport
	}

	standardHttpClient := http.Client{
		Timeout: time.Duration(90) * time.Second,
	}

	logger.Debug("config", zap.Any("config", cfg))
	br := &BoltRouter{
		config: cfg,

		boltHttpClient:     &boltHttpClient,
		standardHttpClient: &standardHttpClient,
		boltVars:           boltVars,
	}

	return br, nil
}

package boltrouter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"
)

// BoltRouter is used to find bolt endpoints and route a aws call to the right endpoint.
type BoltRouter struct {
	config Config

	boltHttpClient *http.Client
	boltVars       *BoltVars
	awsCred        aws.Credentials
}

// NewBoltRouter creates a new BoltRouter.
func NewBoltRouter(ctx context.Context, logger *zap.Logger, cfg Config) (*BoltRouter, error) {
	boltVars, err := GetBoltVars(logger)
	if err != nil {
		return nil, fmt.Errorf("could not get BoltVars: %w", err)
	}

	// custom transport is needed to allow certificate validation from bolt hostname
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{
		ServerName: boltVars.BoltHostname.Get(),
	}
	boltHttpClient := http.Client{
		Timeout:   time.Duration(90) * time.Second,
		Transport: customTransport,
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not load aws default config: %w", err)
	}
	cred, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve aws credentials: %w", err)
	}

	br := &BoltRouter{
		config: cfg,

		boltHttpClient: &boltHttpClient,
		boltVars:       boltVars,
		awsCred:        cred,
	}

	return br, nil
}

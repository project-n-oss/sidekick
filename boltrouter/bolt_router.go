package boltrouter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// BoltRouter is used to find bolt endpoints and route a aws call to the right endpoint.
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

	// custom transport is needed to allow certificate validation from bolt hostname
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{
		ServerName: boltVars.BoltHostname.Get(),
	}
	boltHttpClient := http.Client{
		Timeout:   time.Duration(90) * time.Second,
		Transport: customTransport,
	}

	standardHttpClient := http.Client{
		Timeout: time.Duration(90) * time.Second,
	}

	br := &BoltRouter{
		config: cfg,

		boltHttpClient:     &boltHttpClient,
		standardHttpClient: &standardHttpClient,
		boltVars:           boltVars,
	}

	if err := br.refreshAWSCredentials(ctx); err != nil {
		return nil, err
	}

	go br.refreshAWSCredentialsPeriodically(ctx)

	return br, nil
}

func (br *BoltRouter) refreshAWSCredentials(ctx context.Context) error {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("could not load aws default config: %w", err)
	}

	cred, err := awsCfg.Credentials.Retrieve(ctx)
	fmt.Printf("Credential can expire %v\n", cred.CanExpire)
	fmt.Printf("Credential expires at %v\n", cred.Expires)
	if err != nil {
		return fmt.Errorf("could not retrieve aws credentials: %w", err)
	}

	br.awsCred = cred
	return nil
}

func (br *BoltRouter) refreshAWSCredentialsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
		case <-ticker.C:
			br.refreshAWSCredentials(ctx)
		}
	}
}

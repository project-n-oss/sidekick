package boltrouter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// BoltRouter is used to find bolt endpoints and route a AWS call to the right endpoint.
type BoltRouter struct {
	config Config

	boltHttpClient     *http.Client
	standardHttpClient *http.Client
	gcpHttpClient      *http.Client
	boltVars           *BoltVars
	logger             *zap.Logger
}

// NewBoltRouter creates a new BoltRouter.
func NewBoltRouter(ctx context.Context, logger *zap.Logger, cfg Config) (*BoltRouter, error) {
	boltVars, err := GetBoltVars(ctx, logger, cfg.CloudPlatform)
	if err != nil {
		return nil, fmt.Errorf("could not get BoltVars: %w", err)
	}

	var boltHttpClient http.Client
	var gcpHttpClient http.Client

	standardHttpClient := http.Client{
		Timeout: time.Duration(90) * time.Second,
	}

	if cfg.CloudPlatform == "aws" {
		boltHttpClient = http.Client{
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
	} else if cfg.CloudPlatform == "gcp" {
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/devstorage.read_write")
		if err != nil {
			return nil, err
		}
		ts := oauth2.TokenSource(creds.TokenSource)
		// custom transport is needed to allow certificate validation from bolt hostname
		if tp, ok := http.DefaultTransport.(*http.Transport); ok {
			customTransport := tp.Clone()
			customTransport.TLSClientConfig = &tls.Config{
				ServerName: "bolt.us-central1.km-aug30-0.bolt.projectn.co", //boltVars.BoltHostname.Get(), TODO: fix this
			}
			boltHttpClient = http.Client{
				Timeout: time.Duration(90) * time.Second,
				Transport: &oauth2.Transport{
					Base:   customTransport,
					Source: ts,
				},
			}
		}

		gcpHttpClient = http.Client{
			Timeout: time.Duration(90) * time.Second,
			Transport: &oauth2.Transport{
				Base:   http.DefaultTransport,
				Source: ts,
			},
		}

	} else {
		return nil, fmt.Errorf("invalid cloud platform: %s", cfg.CloudPlatform)
	}

	logger.Debug("config", zap.Any("config", cfg))
	br := &BoltRouter{
		config: cfg,
		logger: logger,

		boltHttpClient:     &boltHttpClient,
		standardHttpClient: &standardHttpClient,
		gcpHttpClient:      &gcpHttpClient,
		boltVars:           boltVars,
	}

	return br, nil
}

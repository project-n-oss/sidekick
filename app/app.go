package app

import (
	"context"
	"net/http"
	"time"

	sidekickAws "github.com/project-n-oss/sidekick/app/aws"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type App struct {
	cfg    Config
	logger *zap.Logger

	standardHttpClient *http.Client
	gcpHttpClient      *http.Client
}

func New(ctx context.Context, logger *zap.Logger, cfg Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	standardHttpClient := http.Client{
		Timeout: time.Duration(90) * time.Second,
	}

	ret := &App{
		cfg:                cfg,
		logger:             logger,
		standardHttpClient: &standardHttpClient,
	}

	switch cfg.CloudPlatform {
	case AwsCloudPlatform.String():
		sidekickAws.RefreshCredentialsPeriodically(ctx, logger)
		sidekickAws.RefreshS3ClientPeriodically(ctx, logger)

	case GcpCloudPlatform.String():
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/devstorage.read_write")
		if err != nil {
			return nil, err
		}
		ts := oauth2.TokenSource(creds.TokenSource)
		ret.gcpHttpClient = &http.Client{
			Timeout: time.Duration(90) * time.Second,
			Transport: &oauth2.Transport{
				Base:   http.DefaultTransport,
				Source: ts,
			},
		}
	}

	logger.Sugar().Infof("Cloud Platform: %s, CrunchErr: %v", cfg.CloudPlatform, !cfg.NoCrunchErr)
	return ret, nil
}

func (a *App) Close(ctx context.Context) error {
	return nil
}

func (a *App) Health(ctx context.Context) error {
	return nil
}

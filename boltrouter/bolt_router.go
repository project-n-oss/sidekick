package boltrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

// BoltRouter is used to find bolt endpoints and route a aws call to the right endpoint.
type BoltRouter struct {
	logger *zap.Logger
	config Config

	httpClient *retryablehttp.Client

	boltVars *BoltVars
	awsCred  aws.Credentials
}

// NewBoltRouter creates a new BoltRouter.
func NewBoltRouter(ctx context.Context, logger *zap.Logger, cfg Config) (*BoltRouter, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = NewLeveledLogger(logger)

	boltVars, err := GetBoltVars(logger)
	if err != nil {
		return nil, fmt.Errorf("could not get BoltVars: %w", err)
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
		logger: logger,
		config: cfg,

		httpClient: retryClient,
		boltVars:   boltVars,
		awsCred:    cred,
	}

	return br, nil
}

// SelectBoltEndpoint selects a bolt endpoint from BoltVars.BoltEndpoints from the passed in requMethod.
// This method will err if not endpoints were selected.
func (br *BoltRouter) SelectBoltEndpoint(ctx context.Context, reqMethod string) (*url.URL, error) {
	preferredOrder := br.getPreferredEndpointOrder(reqMethod)
	boltEndpoints := br.boltVars.BoltEndpoints.Get()
	for _, key := range preferredOrder {
		availableEndpoints, ok := boltEndpoints[key]
		if ok && len(availableEndpoints) > 0 {
			randomIndex := rand.Intn(len(availableEndpoints))
			selectedEndpoint := availableEndpoints[randomIndex]
			return url.Parse(fmt.Sprintf("https://%s", selectedEndpoint))
		}
	}

	return nil, fmt.Errorf("could not select any bolt endpoints")
}

// getPreferredEndpointOrder returns BoltVars.ReadOrderEndpoints if BoltVars.HttpReadMethodTypes contains
// reqMethod, otherwise return BoltVars.WriteOrderEndpoints.
func (br *BoltRouter) getPreferredEndpointOrder(reqMethod string) []string {
	HttpReadMethodTypes := br.boltVars.HttpReadMethodTypes.Get()
	for _, methodType := range HttpReadMethodTypes {
		if reqMethod == methodType {
			return br.boltVars.ReadOrderEndpoints.Get()
		}
	}

	writeOrderEnpoints := br.boltVars.WriteOrderEndpoints.Get()
	return writeOrderEnpoints
}

// RefreshEndpointsPeriodically starts a goroutine that calls RefreshEndpoints every 2min
func (br *BoltRouter) RefreshEndpointsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(120 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				br.RefreshEndpoints(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// RefreshEndpoints refreshes the BoltVars.BoltEndpoints variable and restarts the refresh interval.
// Call this method to force refresh BoltVars.BoltEndpoints.
func (br *BoltRouter) RefreshEndpoints(ctx context.Context) error {
	endpoints, err := br.getBoltEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("could not refresh bolt endpoints: %w", err)
	}
	br.boltVars.BoltEndpoints.Set(endpoints)
	return nil
}

// getBoltEndpoints queries quicksilver and returns a BoltEndpointsMap.
// It will always return BoltVars.BoltHostname if running in local mode.
func (br *BoltRouter) getBoltEndpoints(ctx context.Context) (BoltEndpointsMap, error) {
	// If running locally, we can't connect to quicksilver, return map of boltEndpoint
	if br.config.Local {
		endpoint := br.boltVars.BoltHostname.Get()
		return BoltEndpointsMap{
			"main_write_endpoints":     {endpoint},
			"failover_write_endpoints": {endpoint},
			"main_read_endpoints":      {endpoint},
			"failover_read_endpoints":  {endpoint},
		}, nil
	}

	if br.boltVars.QuicksilverURL.Get() == "" || br.boltVars.Region.Get() == "" {
		return BoltEndpointsMap{}, fmt.Errorf("quicksilverUrl: '%v' or region '%v' are not set", br.boltVars.QuicksilverURL.Get(), br.boltVars.Region.Get())
	}

	requestURL := br.boltVars.QuicksilverURL.Get()
	r, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return BoltEndpointsMap{}, err
	}

	resp, err := br.httpClient.Do(r)
	if err != nil {
		return BoltEndpointsMap{}, fmt.Errorf("could not get endpoints from quicksilver: %w", err)
	}

	defer resp.Body.Close()
	var endpoints BoltEndpointsMap
	err = json.NewDecoder(resp.Body).Decode(&endpoints)
	if err != nil {
		return BoltEndpointsMap{}, nil
	}

	return endpoints, nil
}

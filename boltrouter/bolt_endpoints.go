package boltrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

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
// It will always return an empty mapif running in local mode.
func (br *BoltRouter) getBoltEndpoints(ctx context.Context) (BoltEndpointsMap, error) {
	// If running locally, we can't connect to quicksilver.
	if br.config.Local {
		return BoltEndpointsMap{}, nil
	}

	if br.boltVars.QuicksilverURL.Get() == "" || br.boltVars.Region.Get() == "" {
		return BoltEndpointsMap{}, fmt.Errorf("quicksilverUrl: '%v' or region '%v' are not set", br.boltVars.QuicksilverURL.Get(), br.boltVars.Region.Get())
	}

	requestURL := br.boltVars.QuicksilverURL.Get()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return BoltEndpointsMap{}, err
	}

	resp, err := br.standardHttpClient.Do(r)
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

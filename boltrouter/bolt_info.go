package boltrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	BoltInfoRefreshInterval = 10 * time.Second
)

// SelectBoltEndpoint selects a bolt endpoint from BoltVars.BoltEndpoints from the passed in requMethod.
// This method will err if not endpoints were selected.
func (br *BoltRouter) SelectBoltEndpoint(ctx context.Context, reqMethod string) (*url.URL, error) {
	preferredOrder := br.getPreferredEndpointOrder(reqMethod)
	boltEndpoints := br.boltVars.BoltInfo.Get()

	if len(boltEndpoints) == 0 {
		return nil, fmt.Errorf("boltInfo is empty")
	}

	for _, key := range preferredOrder {
		availableEndpoints, ok := boltEndpoints[key]
		availableEndpointsSlice, castOk := availableEndpoints.([]interface{})
		if !castOk {
			return nil, fmt.Errorf("could not cast availableEndpoints to []string")
		}
		availableEndpointsStr := make([]string, len(availableEndpointsSlice))
		for i, v := range availableEndpointsSlice {
			availableEndpointsStr[i] = v.(string)
		}
		if ok && len(availableEndpointsStr) > 0 {
			randomIndex := rand.Intn(len(availableEndpointsStr))
			selectedEndpoint := availableEndpointsStr[randomIndex]
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

	writeOrderEndpoints := br.boltVars.WriteOrderEndpoints.Get()
	return writeOrderEndpoints
}

// RefreshBoltInfoPeriodically starts a goroutine that calls RefreshBoltInfo every BoltInfoRefreshInterval seconds
func (br *BoltRouter) RefreshBoltInfoPeriodically(ctx context.Context) {
	ticker := time.NewTicker(BoltInfoRefreshInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := br.RefreshBoltInfo(ctx)
				if err != nil {
					logrus.Errorf(err.Error())
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// RefreshBoltInfo refreshes the BoltVars.BoltInfo variable and restarts the refresh interval.
// Call this method to force refresh BoltVars.BoltInfo.
func (br *BoltRouter) RefreshBoltInfo(ctx context.Context) error {
	info, err := br.getBoltInfo(ctx)
	if err != nil {
		return fmt.Errorf("could not refresh bolt info: %w", err)
	}
	br.boltVars.BoltInfo.Set(info)
	return nil
}

// getBoltInfo queries quicksilver and returns a BoltInfo.
// It will always return an empty mapif running in local mode.
func (br *BoltRouter) getBoltInfo(ctx context.Context) (BoltInfo, error) {
	// If running locally, we can't connect to quicksilver.
	if br.config.Local {
		return BoltInfo{}, nil
	}

	if br.boltVars.QuicksilverURL.Get() == "" || br.boltVars.Region.Get() == "" {
		return BoltInfo{}, fmt.Errorf("quicksilverUrl: '%v' or region '%v' are not set", br.boltVars.QuicksilverURL.Get(), br.boltVars.Region.Get())
	}

	requestURL := br.boltVars.QuicksilverURL.Get()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return BoltInfo{}, err
	}

	resp, err := br.standardHttpClient.Do(r)
	if err != nil {
		return BoltInfo{}, fmt.Errorf("could not get info from quicksilver: %w", err)
	}

	defer resp.Body.Close()
	var info BoltInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return BoltInfo{}, nil
	}

	return info, nil
}

func (br *BoltRouter) GetCleanerStatus() (bool, error) {
	boltInfo := br.boltVars.BoltInfo.Get()
	clientBehaviorParams := boltInfo["client_behavior_params"].(map[string]interface{})
	cleanerOn := clientBehaviorParams["cleaner_on"]
	cleanerOnBool, ok := cleanerOn.(bool)
	if !ok {
		return false, fmt.Errorf("could not cast cleanerOn to bool")
	}
	return cleanerOnBool, nil
}

// select initial request destination based on cluster_health_metrics and client_behavior_params
func (br *BoltRouter) SelectInitialRequestTarget() (target string, reason string, err error) {
	boltInfo := br.boltVars.BoltInfo.Get()

	clusterHealthy := boltInfo["cluster_healthy"]
	clientBehaviorParams := boltInfo["client_behavior_params"]

	if clusterHealthy == nil || clientBehaviorParams == nil {
		// backwards compatibility: if cluster_healthy or client_behavior_params are not set (potentially running against
		// an older version of quicksilver), we default to bolt (which is the current behavior) as an initial endpoint.
		// This is to avoid a potential regression.
		// Fallback to S3 will happen based on the failover logic.
		return "bolt", "backwards compatibility", nil
	}

	clusterHealthyBool, ok := clusterHealthy.(bool)
	if !ok {
		return "", "", fmt.Errorf("could not cast boltHealthy to bool")
	}

	if clusterHealthyBool {
		clientBehaviorParams, ok := clientBehaviorParams.(map[string]interface{})
		if !ok {
			return "", "", fmt.Errorf("could not cast clientBehaviorParams to map[string]interface{}")
		}

		crunchTrafficPercent, ok := clientBehaviorParams["crunch_traffic_percent"].(string)
		if !ok {
			return "", "", fmt.Errorf("could not cast crunchTrafficPercent to string")
		}

		crunchTrafficPercentInt, err := strconv.Atoi(crunchTrafficPercent)
		if err != nil {
			return "", "", fmt.Errorf("could not cast crunchTrafficPercent to int")
		}

		totalWeight := 1000
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		rnd := r.Intn(totalWeight)

		if rnd < (crunchTrafficPercentInt * totalWeight / 100) {
			return "bolt", "traffic splitting", nil
		} else {
			return "s3", "traffic splitting", nil
		}
	} else {
		return "s3", "cluster unhealthy", nil
	}
}

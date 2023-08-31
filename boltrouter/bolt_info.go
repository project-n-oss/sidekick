package boltrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

const (
	BoltInfoRefreshInterval = 10 * time.Second
)

// SelectBoltEndpoint selects a bolt endpoint from BoltVars.BoltEndpoints from the passed in reqMethod.
// This method will err if not endpoints were selected.
func (br *BoltRouter) SelectBoltEndpoint(reqMethod string) (*url.URL, error) {
	// TODO: check that read replicas are not enabled somehow?
	if br.config.CloudPlatform == "gcp" {
		return url.Parse(fmt.Sprintf("https://%s", br.boltVars.BoltHostname.Get()))
	}

	preferredOrder := br.getPreferredEndpointOrder(reqMethod)
	boltEndpoints := br.boltVars.BoltInfo.Get()

	if len(boltEndpoints) == 0 {
		return nil, fmt.Errorf("boltInfo is empty")
	}

	for _, key := range preferredOrder {
		availableEndpoints, ok := boltEndpoints[key]
		if !ok {
			continue
		}

		availableEndpointsSlice, ok := availableEndpoints.([]interface{})
		if !ok {
			return nil, fmt.Errorf("could not cast availableEndpoints to []string")
		}
		liveEndpoints := make([]string, 0, len(availableEndpointsSlice))
		for _, v := range availableEndpointsSlice {
			endpoint := v.(string)

			if !br.IsOffline(endpoint) {
				liveEndpoints = append(liveEndpoints, endpoint)
			}
		}
		if ok && len(liveEndpoints) > 0 {
			randomIndex := rand.Intn(len(liveEndpoints))
			selectedEndpoint := liveEndpoints[randomIndex]
			if br.config.Local {
				return url.Parse(fmt.Sprintf("http://%s", selectedEndpoint))
			}
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
	br.updateEndpointLiveness()
	return nil
}

// getBoltInfo queries quicksilver and returns a BoltInfo.
// It will always return an empty map if running in local mode.
func (br *BoltRouter) getBoltInfo(ctx context.Context) (BoltInfo, error) {
	// If running locally, we can't connect to quicksilver.
	if br.config.Local {
		if br.config.BoltEndpointOverride == "" {
			return BoltInfo{}, nil
		}
		endpoints := make(BoltInfo)
		endpoints["main_write_endpoints"] = []interface{}{br.config.BoltEndpointOverride}
		endpoints["failover_write_endpoints"] = []interface{}{br.config.BoltEndpointOverride}
		endpoints["main_read_endpoints"] = []interface{}{br.config.BoltEndpointOverride}
		endpoints["failover_read_endpoints"] = []interface{}{br.config.BoltEndpointOverride}
		endpoints["cluster_healthy"] = true
		return endpoints, nil
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
	br.logger.Debug("QS resp", zap.Any("info", info))
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
func (br *BoltRouter) SelectInitialRequestTarget(boltReq *BoltRequest) (target string, reason string, err error) {
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

	if !clusterHealthyBool {
		return "s3", "cluster unhealthy", nil
	}

	params, ok := clientBehaviorParams.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("could not cast clientBehaviorParams to map[string]interface{}")
	}

	crunchTrafficPercent, ok := params["crunch_traffic_percent"].(string)
	if !ok {
		return "", "", fmt.Errorf("could not cast crunchTrafficPercent to string")
	}

	crunchTrafficPercentInt, err := strconv.Atoi(crunchTrafficPercent)
	if err != nil {
		return "", "", fmt.Errorf("could not parse crunchTrafficPercent to int. %v", err)
	}

	if br.config.CrunchTrafficSplit == CrunchTrafficSplitByObjectKeyHash {
		// Take a mod of hashValue and check if it is less than crunchTrafficPercentInt
		if int(boltReq.crcHash)%100 < crunchTrafficPercentInt {
			return "bolt", "traffic splitting", nil
		}
	} else {
		// Randomly select bolt with crunchTrafficPercentInt % probability.
		if rand.Intn(100) < crunchTrafficPercentInt {
			return "bolt", "traffic splitting", nil
		}
	}

	return "s3", "traffic splitting", nil
}

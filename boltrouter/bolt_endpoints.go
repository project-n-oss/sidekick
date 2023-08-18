package boltrouter

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"net/url"
)

// MaybeMarkOffline is called to mark bolt endpoints offline on error.
func (br *BoltRouter) MaybeMarkOffline(url *url.URL, err error) {
	if err == nil || errors.Is(err, context.Canceled) {
		return
	}

	boltEndpoint := url.Host

	br.boltVars.livenessLock.Lock()
	defer br.boltVars.livenessLock.Unlock()

	br.logger.Debug("adding offline status", zap.String("endpoint", boltEndpoint))

	br.boltVars.offlineEndpoints[boltEndpoint] = true
}

// IsOffline is called to check if a bolt endpoint is offline
func (br *BoltRouter) IsOffline(endpoint string) bool {
	//br.logger.Debug("checking offline status", zap.String("endpoint", endpoint))

	br.boltVars.livenessLock.RLock()
	defer br.boltVars.livenessLock.RUnlock()

	_, found := br.boltVars.offlineEndpoints[endpoint]

	return found
}

// updateEndpointLiveness is called after new endpoints from QS are returned.
// In such case, we remove all previous offlineEndpoints, and assume endpoints returned from QS are all live.
func (br *BoltRouter) updateEndpointLiveness() {
	// Look at latest bolt endpoints, and remove or change offline endpoint
	br.boltVars.livenessLock.Lock()
	defer br.boltVars.livenessLock.Unlock()

	if len(br.boltVars.offlineEndpoints) > 0 {
		br.boltVars.offlineEndpoints = make(map[string]bool)
	}
}

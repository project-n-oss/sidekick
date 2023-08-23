package boltrouter

type CrunchTrafficSplitType string

const (
	CrunchTrafficSplitByObjectKeyHash CrunchTrafficSplitType = "objectkeyhash"
	CrunchTrafficSplitByRandomRequest CrunchTrafficSplitType = "random"
)

type Config struct {
	// If set, boltrouter will run in local mode.
	// For example, it will not query quicksilver to get endpoints.
	Local bool `yaml:"Local"`

	// Set the BoltEndpointOverride while running from local mode.
	BoltEndpointOverride string `yaml:"BoltEndpointOverride"`

	// Enable pass through in Bolt.
	Passthrough bool `yaml:"Passthrough"`

	// Enable failover to a AWS request if the Bolt request fails
	Failover bool `yaml:"Failover"`

	// Enable NoFallback404 to disable fallback on 404 response code from AWS request to Bolt or vice-versa.
	// Fallback is useful on GetObject, where object maybe present in the other source.
	NoFallback404 bool `yaml:"NoFallback404"`

	// There are two ways to split the traffic between bolt and object store
	// 1. Random Crunch Traffic Split
	// 2. Hash Based Crunch Traffic Split
	// Random approach could cause data inconsistency if the requests are mix of GET and PUT.
	CrunchTrafficSplit CrunchTrafficSplitType `yaml:"CrunchTrafficSplit"`
}

var DefaultConfig = Config{
	Local:                false,
	Passthrough:          false,
	Failover:             false,
	NoFallback404:        false,
	BoltEndpointOverride: "",
	CrunchTrafficSplit:   CrunchTrafficSplitByObjectKeyHash,
}

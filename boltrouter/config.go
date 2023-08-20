package boltrouter

type Config struct {
	// If set, boltrouter will run in local mode.
	// For example, it will not query quicksilver to get endpoints.
	Local bool `yaml:"Local"`

	// Enable pass through in Bolt.
	Passthrough bool `yaml:"Passthrough"`

	// Enable failover to a AWS request if the Bolt request fails or vice-versa.
	Failover bool `yaml:"Failover"`
}

var DefaultConfig = Config{
	Local:       false,
	Passthrough: false,
	Failover:    true,
}

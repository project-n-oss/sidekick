package boltrouter

type Config struct {
	// If set, boltrouter will be running in local mode.
	// For example, boultrouter will not query quicksilver to get endpoints.
	Local bool `yaml:"Local"`

	// Set the BoltEndpointOverride while running from local mode.
	BoltEndpointOverride string `yaml:"BoltEndpointOverride"`

	// Enable pass through in Bolt.
	Passthrough bool `yaml:"Passthrough"`

	// Enable failover to a aws request if the Bolt request fails.
	Failover bool `yaml:"Failover"`
}

func NewConfig() Config {
	return Config{
		Local:                false,
		Passthrough:          false,
		Failover:             true,
		BoltEndpointOverride: "",
	}
}

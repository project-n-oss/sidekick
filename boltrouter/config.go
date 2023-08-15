package boltrouter

type Config struct {
	// If set, boltrouter will be running in local mode.
	// For example, boultrouter will not query quicksilver to get endpoints.
	Local bool `yaml:"Local"`

	// Enable pass through in Bolt.
	Passthrough bool `yaml:"Passthrough"`

	// Enable failover to a AWS request if the Bolt request fails.
	Failover bool `yaml:"Failover"`
}

func NewConfig() Config {
	return Config{
		Local:       false,
		Passthrough: false,
		Failover:    true,
	}
}

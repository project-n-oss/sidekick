package boltrouter

type Config struct {
	// If set, boltrouter will be running in local mode.
	// For example, boultrouter will not query quicksilver to get endpoints.
	Local bool

	// Enable pass through in Bolt.
	Passthrough bool
}

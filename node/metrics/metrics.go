package metrics

// Config contains all configuration parameters related
// to metrics for celestia-node.
type Config struct {
	TracingEnabled bool
}

func DefaultConfig() Config {
	return Config{
		TracingEnabled: false,
	}
}

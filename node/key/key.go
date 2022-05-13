package key

import "github.com/cosmos/cosmos-sdk/crypto/keyring"

// Config contains configuration values for constructing
// the keyring / signer used for state interaction.
type Config struct {
	// AccName is the prefix of the key the node should use
	// as the default account of the node
	AccName string
	// Backend indicates keyring backend type (e.g.: "test" | "os" | "file")
	Backend string
}

func DefaultConfig() Config {
	return Config{
		"celes",
		keyring.BackendOS,
	}
}

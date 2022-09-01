package core

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
)

func Remote(cfg Config) func(lc fx.Lifecycle) (core.Client, error) {
	return func(lc fx.Lifecycle) (core.Client, error) {
		if cfg.IP == "" {
			return nil, fmt.Errorf("no celestia-core endpoint given")
		}
		client, err := core.NewRemote(cfg.IP, cfg.RPCPort)
		if err != nil {
			return nil, err
		}
		return client, err
	}
}

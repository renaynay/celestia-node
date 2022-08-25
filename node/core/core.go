package core

import (
	"github.com/ipfs/go-blockservice"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/header"
	headercore "github.com/celestiaorg/celestia-node/header/core"
)

func HeaderListener(
	lc fx.Lifecycle,
	ex *core.BlockFetcher,
	bcast header.Broadcaster,
	bServ blockservice.BlockService,
	construct header.ConstructFn,
) *headercore.Listener {
	cl := headercore.NewListener(bcast, ex, bServ, construct)
	lc.Append(fx.Hook{
		OnStart: cl.Start,
		OnStop:  cl.Stop,
	})
	return cl
}

// RemoteClient provides a constructor for core.Client over RPC.
func RemoteClient(cfg Config) (core.Client, error) {
	return core.NewRemote(cfg.IP, cfg.RPCPort)
}

package gateway

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/api/gateway"
	headerServ "github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	shareServ "github.com/celestiaorg/celestia-node/nodebuilder/share"
)

var log = logging.Logger("module/gateway")

func ConstructModule(tp node.Type, cfg *Config) fx.Option {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate()
	if !cfg.Enabled {
		return fx.Options()
	}

	baseComponents := fx.Options(
		fx.Supply(cfg),
		fx.Error(cfgErr),
		fx.Provide(fx.Annotate(
			server,
			fx.OnStart(func(ctx context.Context, server *gateway.Server) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *gateway.Server) error {
				return server.Stop(ctx)
			}),
		)),
	)

	switch tp {
	case node.Light, node.Full:
		return fx.Module(
			"gateway",
			baseComponents,
			fx.Invoke(Handler),
		)
	case node.Bridge:
		return fx.Module(
			"gateway",
			baseComponents,
			fx.Invoke(func(
				share shareServ.Module,
				header headerServ.Module,
				serv *gateway.Server,
			) {
				Handler(share, header, nil, serv)
			}),
		)
	default:
		panic("invalid node type")
	}
}

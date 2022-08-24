package node

import (
	"context"
	"github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/celestia-node/node/core"
	"github.com/celestiaorg/celestia-node/node/header"
	"github.com/celestiaorg/celestia-node/node/p2p"
	"github.com/celestiaorg/celestia-node/node/rpc"
	"github.com/celestiaorg/celestia-node/node/services"
	"github.com/celestiaorg/celestia-node/node/share"
	"github.com/celestiaorg/celestia-node/node/state"
	"github.com/celestiaorg/celestia-node/params"
	headerServ "github.com/celestiaorg/celestia-node/service/header"
	rpcServ "github.com/celestiaorg/celestia-node/service/rpc"
	shareServ "github.com/celestiaorg/celestia-node/service/share"
	stateServ "github.com/celestiaorg/celestia-node/service/state"
	logging "github.com/ipfs/go-log/v2"
	"github.com/raulk/go-watchdog"
	"go.uber.org/fx"
	"sync"
	"time"
)

func Module(tp Type, cfg *node.Config, store node.Store, moduleOpts node.ModuleOpts) fx.Option {

	baseComponents := fx.Options(
		fx.Provide(params.DefaultNetwork),
		fx.Provide(params.BootstrappersFor),
		fx.Provide(context.Background),
		fx.Supply(cfg),
		fx.Supply(store.Config),
		fx.Provide(store.Datastore),
		fx.Provide(store.Keystore),
		// p2p components
		fx.Invoke(invokeWatchdog(store.Path())),
		p2p.Components(cfg.P2P),
		// refactored node modules
		state.Module(tp, cfg.State, moduleOpts.State...),
		header.Module(tp, cfg.Header, moduleOpts.Header...),
		share.Module(tp, cfg.Share, moduleOpts.Share...),
		rpc.Module(tp, cfg.RPC, moduleOpts.RPC...),
		core.Module(tp, cfg.Core, moduleOpts.Core...),
	)

	switch tp {
	case Light:
		return fx.Module(
			"node",
			fx.Supply(Light),
			baseComponents,
			fx.Provide(services.DASer),
			fx.Invoke(rpc.Handler),
		)
	case Bridge:
		return fx.Module(
			"node",
			fx.Supply(Bridge),
			baseComponents,
			fx.Invoke(func(
				state *stateServ.Service,
				share *shareServ.Service,
				header *headerServ.Service,
				rpcSrv *rpcServ.Server,
			) {
				rpc.Handler(state, share, header, rpcSrv, nil)
			}),
		)
	case Full:
		return fx.Module(
			"node",
			fx.Supply(Full),
			baseComponents,
			fx.Provide(services.DASer),
			fx.Invoke(rpc.Handler),
		)
	default:
		panic("wrong node type")
	}
}

// invokeWatchdog starts the memory watchdog that helps to prevent some of OOMs by forcing GCing
// It also collects heap profiles in the given directory when heap grows to more than 90% of memory usage
func invokeWatchdog(pprofdir string) func(lc fx.Lifecycle) error {
	return func(lc fx.Lifecycle) (errOut error) {
		onceWatchdog.Do(func() {
			// to get watchdog information logged out
			watchdog.Logger = logWatchdog
			// these set up heap pprof auto capturing on disk when threshold hit 90% usage
			watchdog.HeapProfileDir = pprofdir
			watchdog.HeapProfileMaxCaptures = 10
			watchdog.HeapProfileThreshold = 0.9

			policy := watchdog.NewWatermarkPolicy(0.50, 0.60, 0.70, 0.85, 0.90, 0.925, 0.95)
			err, stop := watchdog.SystemDriven(0, time.Second*5, policy)
			if err != nil {
				errOut = err
				return
			}

			lc.Append(fx.Hook{
				OnStop: func(context.Context) error {
					stop()
					return nil
				},
			})
		})
		return
	}
}

// TODO(@Wondetan): We must start watchdog only once. This is needed for tests where we run multiple instance
//  of the Node. Ideally, the Node should have some testing options instead, so we can check for it and run without
//  such utilities but it does not hurt to run one instance of watchdog per test.
var onceWatchdog = sync.Once{}

var logWatchdog = logging.Logger("watchdog")

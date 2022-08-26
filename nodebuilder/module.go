package nodebuilder

import (
	"context"
	"sync"
	"time"

	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"

	logging "github.com/ipfs/go-log/v2"
	"github.com/raulk/go-watchdog"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/daser"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/rpc"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
	"github.com/celestiaorg/celestia-node/params"
)

func Module(tp node.Type, cfg *Config, store Store, moduleOpts moduleOpts) fx.Option {

	baseComponents := fx.Options(
		fx.Provide(params.DefaultNetwork),
		fx.Provide(params.BootstrappersFor),
		fx.Provide(context.Background),
		fx.Supply(cfg),
		fx.Supply(store.Config),
		fx.Provide(store.Datastore),
		fx.Provide(store.Keystore),
		fx.Invoke(invokeWatchdog(store.Path())),
		// refactored node modules
		p2p.Module(&cfg.P2P, moduleOpts.p2p...),
		state.Module(tp, &cfg.State, moduleOpts.state...),
		header.Module(tp, &cfg.Header, moduleOpts.header...),
		share.Module(tp, &cfg.Share, moduleOpts.share...),
		rpc.Module(tp, &cfg.RPC, moduleOpts.rpc...),
		core.Module(tp, &cfg.Core, moduleOpts.core...),
		daser.Module(tp),
	)

	return fx.Module(
		"node",
		fx.Supply(tp),
		baseComponents,
	)
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

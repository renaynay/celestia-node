package state

import (
	"github.com/celestiaorg/celestia-node/node/key"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-app/app"
	apptypes "github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/celestia-node/libs/keystore"
	"github.com/celestiaorg/celestia-node/params"
	"github.com/celestiaorg/celestia-node/service/state"
)

var (
	log = logging.Logger("state-access-constructor")
)

func CoreAccessor(endpoint string, conf key.Config) func(fx.Lifecycle, keystore.Keystore, params.Network) (state.Accessor, error) {
	return func(lc fx.Lifecycle, ks keystore.Keystore, net params.Network) (state.Accessor, error) {
		// sanity check keyring backend
		// TODO @renaynay: Include option for setting custom `userInput` parameter with
		//  implementation of https://github.com/celestiaorg/celestia-node/issues/415.
		ring, err := keyring.New(app.Name, conf.Backend, ks.Path(), os.Stdin)
		if err != nil {
			return nil, err
		}
		signer := apptypes.NewKeyringSigner(ring, conf.AccName, string(net))

		log.Infow("constructed keyring signer", "backend", conf.Backend, "path", ks.Path(),
			"keyring account name", conf.AccName)

		ca := state.NewCoreAccessor(signer, endpoint)
		lc.Append(fx.Hook{
			OnStart: ca.Start,
			OnStop:  ca.Stop,
		})
		return ca, nil
	}
}

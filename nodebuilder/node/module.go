package node

import (
	"github.com/gbrlsnchs/jwt/v3"
	"go.uber.org/fx"
)

func ConstructModule(tp Type, path string) fx.Option {
	return fx.Module(
		"node",
		fx.Provide(func() (*jwt.HMACSHA, error) {
			return rpcAuthSecret(tp, path)
		}),
		fx.Provide(func() Module {
			return newAdmin(tp)
		}),
	)
}

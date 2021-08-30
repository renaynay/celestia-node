package util

import (
	"context"

	"go.uber.org/fx"
)

// LifecycleCtx creates a context which will be cancelled when lifecycle stops
//
func LifecycleCtx(lc fx.Lifecycle) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
	return ctx
}

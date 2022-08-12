package cmd

import (
	"context"

	"github.com/celestiaorg/celestia-node/node"
)

// NodeType reads the node type from the context.
func NodeType(ctx context.Context) node.NodeType {
	return ctx.Value(nodeTypeKey{}).(node.NodeType)
}

// StorePath reads the store path from the context.
func StorePath(ctx context.Context) string {
	return ctx.Value(storePathKey{}).(string)
}

// WithNodeType sets the node type in the given context.
func WithNodeType(ctx context.Context, tp node.NodeType) context.Context {
	return context.WithValue(ctx, nodeTypeKey{}, tp)
}

// WithStorePath sets Store Path in the given context.
func WithStorePath(ctx context.Context, storePath string) context.Context {
	return context.WithValue(ctx, storePathKey{}, storePath)
}

// NodeOptions returns config options parsed from Environment(Flags, ENV vars, etc)
func NodeOptions(ctx context.Context) []node.Option {
	options, ok := ctx.Value(optionsKey{}).([]node.Option)
	if !ok {
		return []node.Option{}
	}
	return options
}

// WithNodeOptions add new options to Env.
func WithNodeOptions(ctx context.Context, opts ...node.Option) context.Context {
	options := NodeOptions(ctx)
	return context.WithValue(ctx, optionsKey{}, append(options, opts...))
}

type (
	optionsKey   struct{}
	storePathKey struct{}
	nodeTypeKey  struct{}
)

package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/filecoin-project/go-jsonrpc"

	"github.com/celestiaorg/celestia-node/nodebuilder/das"
	"github.com/celestiaorg/celestia-node/nodebuilder/fraud"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/share"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
)

var authKey = "Authorization"

type API interface {
	fraud.Module
	header.Module
	state.Module
	share.Module
	das.Module
	node.Module
}

type Client struct {
	Fraud  fraud.API
	Header header.API
	State  state.API
	Share  share.API
	DAS    das.API
	Node   node.API

	closer multiClientCloser
}

// multiClientCloser is a wrapper struct to close clients across multiple namespaces.
type multiClientCloser struct {
	closers []jsonrpc.ClientCloser
}

// register adds a new closer to the multiClientCloser
func (m *multiClientCloser) register(closer jsonrpc.ClientCloser) {
	m.closers = append(m.closers, closer)
}

// closeAll closes all saved clients.
func (m *multiClientCloser) closeAll() {
	for _, closer := range m.closers {
		closer()
	}
}

func NewClientWithPerms(ctx context.Context, addr string, token string) (*Client, error) {
	var client Client
	var multiCloser multiClientCloser

	// TODO: this duplication of strings many times across the codebase can be avoided with issue #1176
	var modules = map[string]interface{}{
		"share":  &client.Share.Internal,
		"state":  &client.State.Internal,
		"header": &client.Header.Internal,
		"fraud":  &client.Fraud.Internal,
		"das":    &client.DAS.Internal,
	}
	for name, module := range modules {
		authHeader := http.Header{authKey: []string{fmt.Sprintf("Bearer %s", token)}} // TODO get token
		closer, err := jsonrpc.NewClient(ctx, addr, name, module, authHeader)
		if err != nil {
			return nil, err
		}
		multiCloser.register(closer)
	}

	return &client, nil
}

// NewClient creates a new Client with one connection per namespace.
func NewClient(ctx context.Context, addr string) (*Client, error) {
	var client Client
	var multiCloser multiClientCloser

	// TODO: this duplication of strings many times across the codebase can be avoided with issue #1176
	var modules = map[string]interface{}{
		"share":  &client.Share.Internal,
		"state":  &client.State.Internal,
		"header": &client.Header.Internal,
		"fraud":  &client.Fraud.Internal,
		"das":    &client.DAS.Internal,
		"node":   &client.Node.Internal,
	}
	for name, module := range modules {
		authHeader := http.Header{authKey: []string{fmt.Sprintf("Bearer %s", "TODO")}} // TODO get token
		closer, err := jsonrpc.NewClient(ctx, addr, name, module, authHeader)
		if err != nil {
			return nil, err
		}
		multiCloser.register(closer)
	}

	return &client, nil
}

// Close closes the connections to all namespaces registered on the client.
func (c *Client) Close() {
	c.closer.closeAll()
}

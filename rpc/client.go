package rpc

import (
	"context"
	"fmt"

	"github.com/celestiaorg/celestia-core/rpc/client/http"
	ctypes "github.com/celestiaorg/celestia-core/rpc/core/types"
	"github.com/celestiaorg/celestia-core/types"
)

const subscriber = "NewBlockEvents" // TODO @renaynay: do we want to enable other types of ws subscriptions for the
// TODO  client? We only need blocks for now.

type Client struct {
	http       *http.HTTP
	remoteAddr string
}

// NewClient creates a new http.HTTP client that dials the given remote address,
// returning a wrapped http.HTTP client.
func NewClient(protocol, remoteAddr string) (*Client, error) {
	endpoint := fmt.Sprintf("%s://%s", protocol, remoteAddr)
	httpClient, err := http.New(endpoint, "/websocket")
	if err != nil {
		return nil, err
	}

	return &Client{
		http:       httpClient,
		remoteAddr: remoteAddr,
	}, nil
}

// RemoteAddr returns the remote address that was dialed by the Client.
func (c *Client) RemoteAddr() string {
	return c.remoteAddr
}

// GetStatus queries the remote address for its `Status`.
func (c *Client) GetStatus(ctx context.Context) (*ctypes.ResultStatus, error) {
	return c.http.Status(ctx)
}

// GetBlock queries the remote address for a `Block` at the given height.
func (c *Client) GetBlock(ctx context.Context, height *int64) (*ctypes.ResultBlock, error) {
	return c.http.Block(ctx, height)
}

func (c *Client) Start() error {
	if c.http.IsRunning() {
		return nil
	}
	return c.http.Start()
}

// StartBlockSubscription subscribes to new block events from the remote address, returning
// an event channel on success.
func (c *Client) StartBlockSubscription(ctx context.Context) (<-chan ctypes.ResultEvent, error) {
	return c.http.Subscribe(ctx, subscriber, types.QueryForEvent(types.EventNewBlock).String())
}

// StopBlockSubscription stops the subscription to new block events from the remote address and
// closes the event channel. // TODO @renaynay: does it actually close the channel?
func (c *Client) StopBlockSubscription(ctx context.Context) error {
	return c.http.Unsubscribe(ctx, subscriber, types.QueryForEvent(types.EventNewBlockHeader).String())
}

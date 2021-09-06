package rpc

import (
	"context"
	"fmt"
	"github.com/celestiaorg/celestia-core/libs/sync"
	"github.com/celestiaorg/celestia-core/rpc/client/http"
	rpctypes "github.com/celestiaorg/celestia-core/rpc/core/types"
	"github.com/celestiaorg/celestia-core/types"
	core "github.com/celestiaorg/celestia-core/types"
)

const newBlockSubscriber = "NewBlock/Events"

// Client represents an RPC client designed to communicate
// with Celestia Core.
type Client struct {
	http       *http.HTTP
	remoteAddr string

	mu *sync.Mutex

	closeNewBlockListener func() // called on Unsubscribe
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
func (c *Client) GetStatus(ctx context.Context) (*rpctypes.ResultStatus, error) {
	return c.http.Status(ctx)
}

// GetBlock queries the remote address for a `Block` at the given height.
func (c *Client) GetBlock(ctx context.Context, height *int64) (*core.Block, error) {
	raw, err := c.http.Block(ctx, height)
	return raw.Block, err
}

// Start will start the http.HTTP service which is required for starting subscriptions
// on the Client.
func (c *Client) Start() error {
	if c.http.IsRunning() {
		return nil
	}
	return c.http.Start()
}

// SubscribeNewBlockEvent subscribes to new block events from the remote address, returning
// a new block event channel on success.
func (c *Client) SubscribeNewBlockEvent(ctx context.Context) (<-chan *core.Block, error) {
	// start the client if not started yet
	if !c.http.IsRunning() {
		if err := c.http.Start(); err != nil {
			return nil, err
		}
	}
	eventChan, err := c.http.Subscribe(ctx, newBlockSubscriber, types.QueryForEvent(types.EventNewBlock).String())
	if err != nil {
		return nil, err
	}

	// create a wrapper channel for translating ResultEvent to *core.Block
	newBlockChan := make(chan *core.Block)
	// TODO how to do this safely via locking?
	c.closeNewBlockListener = func() { close(newBlockChan) }

	go func(eventChan <-chan rpctypes.ResultEvent, newBlockChan chan *core.Block) {
		for {
			newEvent := <-eventChan
			rawBlock, ok := newEvent.Data.(types.EventDataNewBlock)
			if !ok {
				// TODO log & ignore?
				continue
			}
			newBlockChan <- rawBlock.Block
		}
	}(eventChan, newBlockChan)

	return newBlockChan, nil
}

// UnsubscribeNewBlockEvent stops the subscription to new block events from the remote address.
// TODO @renaynay: does it actually close the channel?
func (c *Client) UnsubscribeNewBlockEvent(ctx context.Context) error {
	if c.closeNewBlockListener != nil {
		c.closeNewBlockListener()
		// TODO use mutex for this
		c.closeNewBlockListener = nil
	}
	return c.http.Unsubscribe(ctx, newBlockSubscriber, types.QueryForEvent(types.EventNewBlockHeader).String())
}

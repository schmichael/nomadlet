package client

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/schmichael/nomadlet/internal/rpc"
	"github.com/schmichael/nomadlet/internal/structs"
	"github.com/schmichael/nomadlet/internal/uuid"
)

type Client struct {
	node  *structs.Node
	rpc   *rpc.Client
	state *structs.State

	log *slog.Logger
}

func NewClient(config *structs.Config) (*Client, error) {
	// Load/initialize state
	state, err := structs.StateLoad(config.StatePath)
	if err != nil {
		return nil, fmt.Errorf("error loading state: %w", err)
	}

	if state.NodeID == "" || state.NodeSecret == "" {
		// Generate new node id and secret
		state.NodeID = uuid.Generate()
		state.NodeSecret = uuid.Generate()
		if err := state.Store(config.StatePath); err != nil {
			return nil, err
		}
	}

	node, err := structs.MakeNode(state, config)
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc.NewClient(state, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		node:  node,
		rpc:   rpcClient,
		state: state,
		log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})),
	}, nil
}

func (c *Client) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		c.log.Debug("interrupt received")
	}()

	for ctx.Err() == nil {
		start := time.Now()
		if err := c.rpc.StatusPing(); err != nil {
			c.log.Error("error pinging server", "error", err)
		} else {
			c.log.Debug("pinged server!", "latency", time.Since(start))
		}
		time.Sleep(10 * time.Second)
	}
	c.log.Debug("client exiting")
}

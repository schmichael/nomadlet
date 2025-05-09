package client

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/schmichael/nomadlet/client/allocrunner"
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
			AddSource: false,
			Level:     slog.LevelDebug,
		})),
	}, nil
}

func (c *Client) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		c.log.Debug("interrupt received")
	}()

	// 1. Register
	var err error
	var regResp *rpc.NodeUpdateResponse
	for ctx.Err() == nil {
		regResp, err = c.rpc.NodeRegister(c.node)
		if err == nil {
			break
		}
		c.log.Error("error registering node... retrying", "error", err)

		time.Sleep(10 * time.Second)
	}
	if ctx.Err() != nil {
		return
	}

	// 2. Heartbeat
	go c.heartbeat(ctx, regResp.HeartbeatTTL)

	c.log.Info("registered node", "resp", regResp)

	// 3. Run allocs
	go c.fetchAllocs(ctx)

	// 9. Ping in a loop because this was the first code I wrote, and I'm too
	//    attached to it to delete it.
	for ctx.Err() == nil {
		start := time.Now()
		if err := c.rpc.StatusPing(); err != nil {
			c.log.Error("error pinging server", "error", err)
		} else {
			c.log.Debug("pinged server!", "latency", time.Since(start))
		}
		time.Sleep(10 * time.Second)
	}
	c.log.Debug("client exited")
}

func (c *Client) heartbeat(ctx context.Context, initial time.Duration) {
	defer c.log.Debug("heartbeat exited")

	timer := time.NewTimer(initial)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		resp, err := c.rpc.NodeUpdateStatus()
		if err != nil {
			c.log.Error("failed to heartbeat; retrying", "error", err)
			//TODO jitter
			timer.Reset(3 * time.Second)
			continue
		}

		c.log.Debug("heartbeat", "initial", initial, "next", resp.HeartbeatTTL)
		timer.Reset(resp.HeartbeatTTL)
	}
}

func (c *Client) fetchAllocs(ctx context.Context) {
	defer c.log.Debug("no longer fetching allocs")

	allocs := map[string]*allocrunner.AllocRunner{}

	for ctx.Err() == nil {
		allocIndexes, err := c.rpc.NodeGetClientAllocs()
		if err != nil {
			c.log.Error("error fetching client allocs", "error", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for allocID, index := range allocIndexes.Allocs {
			ar, ok := allocs[allocID]
			switch {
			case !ok:
				c.log.Debug("starting alloc", "alloc", allocID)
				// New alloc
				ac := allocrunner.Config{
					AllocID:     allocID,
					ModifyIndex: index,
					RPC:         c.rpc,
					Logger:      c.log.With("alloc_id", allocID),
				}
				ar := allocrunner.New(ac)
				allocs[allocID] = ar
				go ar.Run()
			case ar.ModifyIndex() < index:
				// Updated allocs
				ar.Update()
			default:
				// No change
			}
		}

		// Stop missing allocs
		for allocID, ar := range allocs {
			if _, ok := allocIndexes.Allocs[allocID]; !ok {
				c.log.Debug("stopping alloc", "alloc", allocID)
				ar.Stop()
				delete(allocs, allocID)
			}
		}

		//TODO lol
		time.Sleep(3 * time.Second)
	}
}

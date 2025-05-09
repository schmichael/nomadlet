package rpc

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/schmichael/nomadlet/internal/structs"
	"github.com/ugorji/go/codec"
)

const (
	// rpcMagicByte is the synchronous RpcNomad magic byte
	rpcMagicByte byte = 0x01
)

var (
	msgpackHandle = &codec.MsgpackHandle{}
)

// Client is a synchronous RPC client safe to call from multiple goroutines.
type Client struct {
	region     string
	nodeID     string
	nodeSecret string
	addr       string
	conn       net.Conn
	seq        uint64

	mu sync.Mutex
}

func NewClient(state *structs.State, conf *structs.Config) (*Client, error) {
	return &Client{
		region:     conf.Region,
		nodeID:     state.NodeID,
		nodeSecret: state.NodeSecret,
		addr:       conf.Server,
	}, nil
}

func (c *Client) getConn() (net.Conn, error) {
	if c.conn != nil {
		return c.conn, nil
	}
	tcpaddr, err := net.ResolveTCPAddr("tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving server address: %w", c.addr)
	}
	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	if err := conn.SetKeepAlive(true); err != nil {
		return nil, fmt.Errorf("error setting keepalives: %w", err)
	}
	if err := conn.SetNoDelay(true); err != nil {
		return nil, fmt.Errorf("error setting no delay: %w", err)
	}
	if n, err := conn.Write([]byte{rpcMagicByte}); err != nil || n != 1 {
		return nil, fmt.Errorf("error writing magic byte: err=%w n=%d", err, n)
	}
	c.conn = conn
	return conn, nil
}

func (c *Client) closeConn() {
	c.conn.Close()
	c.conn = nil
}

func (c *Client) do(method string, request, response any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	reqHeader := &requestHeader{
		ServiceMethod: method,
		Seq:           c.seq,
	}

	conn, err := c.getConn()
	if err != nil {
		return err
	}

	enc := codec.NewEncoder(conn, msgpackHandle)

	if err := enc.Encode(reqHeader); err != nil {
		c.closeConn()
		return fmt.Errorf("error writing %q request header: %w",
			method, err)
	}

	if err := enc.Encode(request); err != nil {
		c.closeConn()
		return fmt.Errorf("error writing %q request body: %w",
			method, err)
	}

	// Read resposne
	dec := codec.NewDecoder(conn, msgpackHandle)
	respHeader := &responseHeader{}
	if err := dec.Decode(respHeader); err != nil {
		c.closeConn()
		return fmt.Errorf("error writing %q request body: %w",
			method, err)
	}

	if respHeader.Error != "" {
		// Throw away body and return error
		if err := dec.Decode(&struct{}{}); err != nil {
			c.closeConn()
			return fmt.Errorf("error discarding response body: %w - after %q RPC returned an error: %s", err, method, respHeader.Error)
		}
		return errors.New(respHeader.Error)
	}

	// No body for Status.Ping
	if err := dec.Decode(response); err != nil {
		c.closeConn()
		return fmt.Errorf("error reading %q response body: %w", method, err)
	}

	return nil
}

func (c *Client) StatusPing() error {
	req := &queryRequest{
		Region:    c.region,
		AuthToken: c.nodeSecret,
	}

	return c.do("Status.Ping", req, &struct{}{})
}

func (c *Client) NodeRegister(node *structs.Node) (*NodeUpdateResponse, error) {
	req := &nodeRegisterRequest{
		Node: node,
		WriteRequest: WriteRequest{
			Region:    c.region,
			AuthToken: c.nodeSecret,
		},
	}

	resp := &NodeUpdateResponse{}
	if err := c.do("Node.Register", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) NodeUpdateStatus() (*NodeUpdateResponse, error) {
	req := &NodeUpdateStatusRequest{
		NodeID: c.nodeID,
		Status: "ready",

		WriteRequest: WriteRequest{
			Region:    c.region,
			AuthToken: c.nodeSecret,
		},
	}

	resp := &NodeUpdateResponse{}
	if err := c.do("Node.UpdateStatus", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) NodeGetClientAllocs() (*NodeClientAllocsResponse, error) {
	req := &NodeSpecificRequest{
		NodeID:   c.nodeID,
		SecretID: c.nodeSecret,
		WriteRequest: WriteRequest{
			Region:    c.region,
			AuthToken: c.nodeSecret,
		},
	}

	resp := &NodeClientAllocsResponse{}
	if err := c.do("Node.GetClientAllocs", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) GetAlloc(id string) (*structs.Allocation, error) {
	// Use GetAllocs RPC since we don't know the namespace
	req := &AllocsGetRequest{
		AllocIDs: []string{id},
		QueryOptions: QueryOptions{
			Region:    c.region,
			AuthToken: c.nodeSecret,
		},
	}

	resp := &AllocsGetResponse{}
	if err := c.do("Alloc.GetAllocs", req, resp); err != nil {
		return nil, err
	}

	if n := len(resp.Allocs); n != 1 {
		return nil, fmt.Errorf("unexpected number of allocs: %d", n)
	}

	return resp.Allocs[0], nil
}

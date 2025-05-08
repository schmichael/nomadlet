package rpc

import (
	"errors"
	"fmt"
	"net"

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

type Client struct {
	region     string
	nodeSecret string
	addr       string
	conn       net.Conn
}

func NewClient(state *structs.State, conf *structs.Config) (*Client, error) {
	return &Client{
		region:     conf.Region,
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

func (c *Client) StatusPing() error {
	reqHeader := &requestHeader{
		ServiceMethod: "Status.Ping",
		Seq:           1,
	}

	reqBody := &queryRequest{
		Region:    c.region,
		AuthToken: c.nodeSecret,
	}

	conn, err := c.getConn()
	if err != nil {
		return err
	}

	enc := codec.NewEncoder(conn, msgpackHandle)

	if err := enc.Encode(reqHeader); err != nil {
		c.closeConn()
		return fmt.Errorf("error writing %q request header: %w",
			reqHeader.ServiceMethod, err)
	}

	if err := enc.Encode(reqBody); err != nil {
		c.closeConn()
		return fmt.Errorf("error writing %q request body: %w",
			reqHeader.ServiceMethod, err)
	}

	// Read resposne
	dec := codec.NewDecoder(conn, msgpackHandle)
	respHeader := &responseHeader{}
	if err := dec.Decode(respHeader); err != nil {
		c.closeConn()
		return fmt.Errorf("error writing %q request body: %w",
			reqHeader.ServiceMethod, err)
	}

	if respHeader.Error != "" {
		// Throw away body and return error
		if err := dec.Decode(&struct{}{}); err != nil {
			c.closeConn()
			return fmt.Errorf("error discarding response body: %w - after %q RPC returned an error: %s", err, reqHeader.ServiceMethod, respHeader.Error)
		}
		return errors.New(respHeader.Error)
	}

	// No body for Status.Ping
	if err := dec.Decode(&struct{}{}); err != nil {
		c.closeConn()
		return fmt.Errorf("error reading response body: %w", err)
	}

	return nil
}

type requestHeader struct {
	ServiceMethod string
	Seq           uint64
}

type queryRequest struct {
	Region    string
	AuthToken string
}

type responseHeader struct {
	Method string
	Seq    uint64
	Error  string
}

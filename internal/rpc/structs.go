package rpc

import (
	"time"

	"github.com/schmichael/nomadlet/internal/structs"
)

type requestHeader struct {
	ServiceMethod string
	Seq           uint64
}

type queryRequest struct {
	Region    string
	AuthToken string
}

type QueryOptions struct {
	Region    string
	Namespace string
	AuthToken string
}

type WriteRequest struct {
	Region    string
	AuthToken string
}

type QueryMeta struct {
	Index uint64
}

type responseHeader struct {
	Method string
	Seq    uint64
	Error  string
}

type nodeRegisterRequest struct {
	Node *structs.Node

	WriteRequest
}

type NodeServerInfo struct {
	RPCAdvertiseAddr string
	Datacenter       string
}

type NodeUpdateResponse struct {
	HeartbeatTTL          time.Duration
	Servers               []*NodeServerInfo
	SchedulingEligibility string

	QueryMeta
}

type NodeUpdateStatusRequest struct {
	NodeID string
	Status string

	WriteRequest
}

type NodeSpecificRequest struct {
	NodeID   string
	SecretID string

	WriteRequest
}

type NodeClientAllocsResponse struct {
	Allocs map[string]uint64

	QueryMeta
}

type AllocsGetRequest struct {
	AllocIDs []string
	QueryOptions
}

type AllocsGetResponse struct {
	Allocs []*structs.Allocation
	QueryMeta
}

package structs

import (
	"fmt"
	"os"
)

type Topology struct {
	OverrideTotalCompute uint64
}

type NodeProcessorResources struct {
	Topology Topology
}

type NodeResources struct {
	MinDynamicPort int
	MaxDynamicPort int

	Processors NodeProcessorResources
}

type Node struct {
	ID         string
	SecretID   string
	Datacenter string
	Name       string
	Status     string

	NodeResources *NodeResources
}

func MakeNode(state *State, config *Config) (*Node, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error determining node name: %w", err)
	}

	nr := &NodeResources{
		MinDynamicPort: 20000,
		MaxDynamicPort: 32000,
		Processors: NodeProcessorResources{
			Topology: Topology{
				OverrideTotalCompute: uint64(config.Mhz),
			},
		},
	}

	return &Node{
		ID:            state.NodeID,
		SecretID:      state.NodeSecret,
		Datacenter:    config.Datacenter,
		Name:          name,
		Status:        "initializing",
		NodeResources: nr,
	}, nil
}

package structs

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/schmichael/nomadlet/version"
)

type Topology struct {
	Cores                []Core
	OverrideTotalCompute uint64
}

type NodeProcessorResources struct {
	Topology Topology
}

type Core struct {
	SocketID   uint8
	NodeID     uint8
	ID         uint16
	GuessSpeed uint64
}

type NodeDiskResources struct {
	DiskMB int64
}

type NodeResources struct {
	MinDynamicPort int
	MaxDynamicPort int

	Disk       NodeDiskResources
	Memory     NodeMemoryResources
	Processors NodeProcessorResources
}

type NodeMemoryResources struct {
	MemoryMB int64
}

type Node struct {
	ID         string
	SecretID   string
	Datacenter string
	Name       string
	Status     string

	Attributes map[string]string
	Drivers    map[string]*DriverInfo

	NodeResources *NodeResources
}

func MakeNode(state *State, config *Config) (*Node, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error determining node name: %w", err)
	}

	nr := &NodeResources{
		Disk: NodeDiskResources{
			DiskMB: 2000, //TODO ¯\_(ツ)_/¯
		},
		MinDynamicPort: 20000,
		MaxDynamicPort: 32000,
		Memory: NodeMemoryResources{
			MemoryMB: int64(config.Mem),
		},
		Processors: NodeProcessorResources{
			Topology: Topology{
				Cores: []Core{
					{
						GuessSpeed: uint64(config.Mhz),
					},
				},
				OverrideTotalCompute: uint64(config.Mhz),
			},
		},
	}

	return &Node{
		ID:         state.NodeID,
		SecretID:   state.NodeSecret,
		Datacenter: config.Datacenter,
		Name:       config.Name,
		Status:     "initializing",
		Attributes: map[string]string{
			"cpu.arch":                runtime.GOARCH,
			"cpu.totalcompute":        strconv.Itoa(config.Mhz),
			"cpu.usablecompute":       strconv.Itoa(config.Mhz),
			"kernel.name":             runtime.GOOS,
			"memory.totalbytes":       strconv.Itoa(config.Mem * 1024 * 1024),
			"nomad.service_discovery": "false",
			"os.signals":              "SIGSEGV,SIGSTOP,SIGSYS,SIGWINCH,SIGNULL,SIGALRM,SIGBUS,SIGHUP,SIGILL,SIGIO,SIGTRAP,SIGTSTP,SIGFPE,SIGKILL,SIGPROF,SIGTERM,SIGTTIN,SIGUSR1,SIGUSR2,SIGXCPU,SIGABRT,SIGINT,SIGTTOU,SIGXFSZ,SIGCONT,SIGIOT,SIGPIPE,SIGQUIT",
			"unique.hostname":         hostname,
			"nomadlet.version":        version.Version,
		},
		Drivers: map[string]*DriverInfo{
			"raw_exec": &DriverInfo{
				Detected:          true,
				Healthy:           true,
				HealthDescription: "never felt better",
			},
		},

		NodeResources: nr,
	}, nil
}

type DriverInfo struct {
	Attributes        map[string]string
	Detected          bool
	Healthy           bool
	HealthDescription string
}

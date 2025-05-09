package structs

import "time"

type Allocation struct {
	// msgpack omit empty fields during serialization
	_struct bool `codec:",omitempty"` // nolint: structcheck

	ID        string
	Namespace string
	Job       *Job
	TaskGroup string

	AllocatedResources *AllocatedResources

	DesiredStatus string

	ClientStatus      string
	ClientDescription string

	TaskStates map[string]*TaskState

	SignedIdentities map[string]string `json:"-"`
	SigningKeyID     string

	ModifyIndex uint64

	// AllocModifyIndex is not updated when the client updates allocations. This
	// lets the client pull only the allocs updated by the server.
	AllocModifyIndex uint64
}

func (a *Allocation) Group() *TaskGroup {
	if a.Job == nil {
		return nil
	}
	for _, tg := range a.Job.TaskGroups {
		if tg.Name == a.TaskGroup {
			return tg
		}
	}
	return nil
}

type AllocatedResources struct {
	Tasks          map[string]*AllocatedTaskResources
	TaskLifecycles map[string]*TaskLifecycleConfig
	Shared         AllocatedSharedResources
}

type AllocatedTaskResources struct {
	Cpu      AllocatedCpuResources
	Memory   AllocatedMemoryResources
	Networks Networks
}

type AllocatedMemoryResources struct {
	MemoryMB    int64
	MemoryMaxMB int64
}

type Networks []*NetworkResource

type NetworkResource struct {
	// msgpack omit empty fields during serialization
	_struct bool `codec:",omitempty"` // nolint: structcheck

	ReservedPorts []Port // Host Reserved ports
	DynamicPorts  []Port // Host Dynamically assigned ports
}

type Port struct {
	_struct bool `codec:",omitempty"` // nolint: structcheck

	Label           string
	Value           int
	To              int
	HostNetwork     string
	IgnoreCollision bool
}

type AllocatedCpuResources struct {
	CpuShares     int64
	ReservedCores []uint16
}

type TaskLifecycleConfig struct {
	Hook    string
	Sidecar bool
}

type AllocatedSharedResources struct {
	Networks Networks
	DiskMB   int64
	Ports    AllocatedPorts
}

type AllocatedPorts []AllocatedPortMapping

type AllocatedPortMapping struct {
	_struct bool `codec:",omitempty"` // nolint: structcheck

	Label           string
	Value           int
	To              int
	HostIP          string
	IgnoreCollision bool
}

type TaskState struct {
	State       string
	Failed      bool
	Restarts    uint64
	LastRestart time.Time
	StartedAt   time.Time
	FinishedAt  time.Time
}

type Job struct {
	ID                       string
	ParentID                 string
	Name                     string
	Type                     string
	TaskGroups               []*TaskGroup
	ParameterizedJob         *ParameterizedJobConfig
	Dispatched               bool
	DispatchIdempotencyToken string
	Payload                  []byte
	Meta                     map[string]string
	Status                   string
}

type ParameterizedJobConfig struct {
	Payload      string
	MetaRequired []string
	MetaOptional []string
}

type TaskGroup struct {
	Name          string
	RestartPolicy *RestartPolicy
	Tasks         []*Task
	Meta          map[string]string
	Networks      Networks
	ShutdownDelay *time.Duration
}

type RestartPolicy struct {
	Attempts        int
	Interval        time.Duration
	Delay           time.Duration
	Mode            string
	RenderTemplates bool
}

type Task struct {
	Name            string
	Driver          string
	User            string
	Config          map[string]any
	Env             map[string]string
	Templates       []*Template
	Resources       *Resources
	RestartPolicy   *RestartPolicy
	DispatchPayload *DispatchPayloadConfig
	Lifecycle       *TaskLifecycleConfig
	Meta            map[string]string
	Leader          bool
	KillSignal      string
	KillTimeout     time.Duration
	ShutdownDelay   time.Duration
}

type Resources struct {
	CPU         int
	Cores       int
	MemoryMB    int
	MemoryMaxMB int
	DiskMB      int
	IOPS        int // COMPAT(0.10): Only being used to issue warnings
	Networks    Networks
	SecretsMB   int
}

type DispatchPayloadConfig struct {
	File string
}

type Template struct {
	SourcePath   string
	DestPath     string
	EmbeddedTmpl string
	ChangeMode   string
	ChangeSignal string
	Splay        time.Duration
	LeftDelim    string
	RightDelim   string
	Envvars      bool
}

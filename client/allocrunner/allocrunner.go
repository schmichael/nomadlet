package allocrunner

import (
	"context"
	"log/slog"
	"time"

	"github.com/schmichael/nomadlet/client/allocrunner/taskrunner"
	"github.com/schmichael/nomadlet/internal/rpc"
	"github.com/schmichael/nomadlet/internal/structs"
)

type AllocRunner struct {
	allocID     string
	modifyIndex uint64

	rpc *rpc.Client

	ctx    context.Context
	cancel context.CancelFunc

	log *slog.Logger
}

func New(conf Config) *AllocRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &AllocRunner{
		allocID:     conf.AllocID,
		modifyIndex: conf.ModifyIndex,
		rpc:         conf.RPC,
		ctx:         ctx,
		cancel:      cancel,
		log:         conf.Logger,
	}
}

func (ar *AllocRunner) Run() {
	defer ar.log.Debug("alloc runner exited")

	var alloc *structs.Allocation
	var err error
	// Fetch alloc
	for ar.ctx.Err() == nil && alloc == nil {
		alloc, err = ar.rpc.GetAlloc(ar.allocID)
		if err != nil {
			ar.log.Error("error fetch alloc", "error", err)
			time.Sleep(3 * time.Second)
			continue
		}
	}
	if ar.ctx.Err() != nil {
		return
	}

	tg := alloc.Group()
	if tg == nil {
		ar.log.Error("group not found", "group", alloc.TaskGroup)
		return
	}

	for _, task := range tg.Tasks {
		tc := taskrunner.Config{
			AllocID: ar.allocID,
			Task:    task,
			Logger:  ar.log.With("task", task.Name),
		}
		tr := taskrunner.New(tc)
		go tr.Run(ar.ctx)
	}
}

func (ar *AllocRunner) ModifyIndex() uint64 {
	return ar.modifyIndex
}

func (ar *AllocRunner) Update() {
	ar.log.Warn("update not implemented")
}

func (ar *AllocRunner) Stop() {
	ar.log.Info("stopping")
	ar.cancel()
}

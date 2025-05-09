package allocrunner

import (
	"log/slog"

	"github.com/schmichael/nomadlet/internal/rpc"
)

type Config struct {
	AllocID     string
	ModifyIndex uint64
	RPC         *rpc.Client
	Logger      *slog.Logger
}

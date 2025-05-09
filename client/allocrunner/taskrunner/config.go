package taskrunner

import (
	"log/slog"

	"github.com/schmichael/nomadlet/internal/structs"
)

type Config struct {
	AllocID string
	Task    *structs.Task
	Logger  *slog.Logger
}

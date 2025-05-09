package taskrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/schmichael/nomadlet/internal/structs"
)

type TaskRunner struct {
	allocID string
	task    *structs.Task

	log *slog.Logger
}

func New(conf Config) *TaskRunner {
	return &TaskRunner{
		allocID: conf.AllocID,
		task:    conf.Task,
		log:     conf.Logger,
	}
}

func (tr *TaskRunner) Run(ctx context.Context) {
	defer tr.log.Info("task runner exited")

	buf, _ := json.MarshalIndent(tr.task, "", "  ")
	fmt.Println("TASK:\n" + string(buf))

	// msgpack was a mistake
	commandBytes, _ := tr.task.Config["command"].([]byte)
	command := string(commandBytes)
	if command == "" {
		tr.log.Error("missing command")
		return
	}
	path, err := exec.LookPath(command)
	if err != nil {
		tr.log.Error("error finding command", "error", err)
		return
	}

	var args []string
	argsI, ok := tr.task.Config["args"]
	if ok {
		argsIslice, ok := argsI.([]any)
		if !ok {
			tr.log.Error("invalid type for args: %T", argsI)
			return
		}
		for i, ai := range argsIslice {
			a, ok := ai.([]byte)
			if !ok {
				tr.log.Error("invalid type for arg element %d: %T", i, ai)
				return
			}
			args = append(args, string(a))
		}
	}

	var env []string
	for k, v := range tr.task.Env {
		env = append(env, k+"="+v)
	}

	stdout, err := os.Create(fmt.Sprintf("%s-%s.stdout.log", tr.allocID, tr.task.Name))
	if err != nil {
		tr.log.Error("unable to create stdout log", "error", err)
		return
	}
	defer stdout.Close()

	stderr, err := os.Create(fmt.Sprintf("%s-%s.stderr.log", tr.allocID, tr.task.Name))
	if err != nil {
		tr.log.Error("unable to create stderr log", "error", err)
		return
	}
	defer stderr.Close()

	cmd := &exec.Cmd{
		Path:   path,
		Args:   args,
		Env:    env,
		Stdout: stdout,
		Stderr: stderr,
	}

	if err := cmd.Run(); err != nil {
		tr.log.Error("error running command", "error", err)
	}
}

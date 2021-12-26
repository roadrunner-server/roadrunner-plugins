package server

import (
	"context"
	"os/exec"

	"github.com/spiral/roadrunner/v2/pool"
	"github.com/spiral/roadrunner/v2/worker"
)

// Server creates workers for the application.
type Server interface {
	// CmdFactory return a new command based on the .rr.yaml server.command section
	CmdFactory(env map[string]string) func() *exec.Cmd
	// NewWorker return a new worker with provided and attached by the user listeners and environment variables
	NewWorker(ctx context.Context, env map[string]string) (*worker.Process, error)
	// NewWorkerPool return new pool of workers (PHP) with attached events listeners, env variables and based on the provided configuration
	NewWorkerPool(ctx context.Context, opt *pool.Config, env map[string]string) (pool.Pool, error)
}

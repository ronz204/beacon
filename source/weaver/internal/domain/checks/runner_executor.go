package checks

import "context"

type ExecutionResult struct {
	ExitCode int
	TimedOut bool
	Output   string
}

type Executor interface {
	Run(
		ctx context.Context,
		argv []string,
		dir string,
	) (ExecutionResult, error)
}

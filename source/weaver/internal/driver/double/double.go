package double

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type Changes struct {
	Items []deltas.StateChange
	Err   error
}

func (c Changes) ListChanges(
	context.Context,
	string,
) ([]deltas.StateChange, error) {
	return c.Items, c.Err
}

type Run struct {
	Argv []string
	Dir  string
}

type Executor struct {
	Results map[string]checks.ExecutionResult
	Err     error
	Runs    *[]Run
}

func (e Executor) Run(
	_ context.Context,
	argv []string,
	dir string,
) (checks.ExecutionResult, error) {
	if e.Runs != nil {
		*e.Runs = append(*e.Runs, Run{Argv: argv, Dir: dir})
	}
	if e.Err != nil {
		return checks.ExecutionResult{}, e.Err
	}

	return e.Results[argv[len(argv)-1]], nil
}

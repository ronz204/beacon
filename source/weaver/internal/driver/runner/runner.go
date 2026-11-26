package runner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"time"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

const (
	DefaultTimeout = 10 * time.Minute
	defaultDelay   = 5 * time.Second
	keptOutput     = 64 * 1024
)

type Exec struct {
	Timeout time.Duration
	Delay   time.Duration
	Tee     io.Writer
}

func New(tee io.Writer) Exec {
	return Exec{Timeout: DefaultTimeout, Delay: defaultDelay, Tee: tee}
}

func (e Exec) Run(
	ctx context.Context,
	argv []string,
	dir string,
) (checks.ExecutionResult, error) {
	limited, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	tail := &tailBuffer{limit: keptOutput}
	command := exec.CommandContext(limited, argv[0], argv[1:]...)
	command.Dir = dir
	command.Stdout = e.output(tail)
	command.Stderr = command.Stdout
	command.Cancel = func() error { return killTree(command) }
	command.WaitDelay = e.Delay
	prepareTree(command)

	err := command.Run()
	result := checks.ExecutionResult{Output: tail.String()}

	if limited.Err() != nil && ctx.Err() == nil {
		result.TimedOut = true
		result.ExitCode = -1

		return result, nil
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}

	return finish(result, command, err)
}

func (e Exec) output(tail *tailBuffer) io.Writer {
	if e.Tee == nil {
		return tail
	}

	return io.MultiWriter(tail, e.Tee)
}

func finish(
	result checks.ExecutionResult,
	command *exec.Cmd,
	err error,
) (checks.ExecutionResult, error) {
	var exited *exec.ExitError

	switch {
	case err == nil:
		return result, nil
	case errors.As(err, &exited):
		result.ExitCode = exited.ExitCode()
		return result, nil
	case command.ProcessState == nil:
		return result, checks.FailRunnerUnavailable(err)
	}

	return result, err
}

type tailBuffer struct {
	limit int
	data  bytes.Buffer
}

func (t *tailBuffer) Write(chunk []byte) (int, error) {
	t.data.Write(chunk)

	if extra := t.data.Len() - t.limit; extra > 0 {
		t.data.Next(extra)
	}

	return len(chunk), nil
}

func (t *tailBuffer) String() string {
	return t.data.String()
}

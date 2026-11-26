package runner_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
	"github.com/ronz204/beacon/source/weaver/internal/driver/runner"
)

const helperVariable = "BEACON_RUNNER_HELPER"

func TestMain(m *testing.M) {
	if mode := os.Getenv(helperVariable); mode != "" {
		runHelper(mode)
		return
	}

	os.Exit(m.Run())
}

func runHelper(mode string) {
	switch mode {
	case "exit-three":
		fmt.Println("about to fail")
		os.Exit(3)
	case "noisy":
		for line := 1; line <= 5000; line++ {
			fmt.Println("line", line)
		}
	case "hang":
		time.Sleep(time.Minute)
	case "spawn":
		child := exec.Command(os.Args[0])
		child.Env = append(os.Environ(), helperVariable+"=hang")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		_ = child.Start()
		time.Sleep(time.Minute)
	}
}

func helperArgv() []string {
	return []string{os.Args[0]}
}

func withMode(t *testing.T, mode string) {
	t.Helper()
	t.Setenv(helperVariable, mode)
}

func TestRunReportsTheExitCodeAndTheOutput(t *testing.T) {
	withMode(t, "exit-three")

	result, err := runner.New(nil).Run(
		context.Background(),
		helperArgv(),
		t.TempDir(),
	)

	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 3 {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(result.Output, "about to fail") {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunKeepsOnlyTheTailOfAVeryLongOutput(t *testing.T) {
	withMode(t, "noisy")
	var streamed bytes.Buffer

	result, err := runner.New(&streamed).Run(
		context.Background(),
		helperArgv(),
		t.TempDir(),
	)

	if err != nil {
		t.Fatal(err)
	}
	if len(result.Output) > 64*1024 {
		t.Fatalf("kept %d bytes", len(result.Output))
	}
	if !strings.Contains(result.Output, "line 5000") {
		t.Fatal("the tail must hold the last line")
	}
	if !strings.Contains(streamed.String(), "line 1\n") {
		t.Fatal("the stream must hold everything")
	}
}

func TestRunReportsAMissingProgramAsUnavailable(t *testing.T) {
	_, err := runner.New(nil).Run(
		context.Background(),
		[]string{"definitely-not-installed-program"},
		t.TempDir(),
	)

	var failure checks.RunnerError
	if !errors.As(err, &failure) || failure.Code != checks.RunnerUnavailable {
		t.Fatalf("err = %v", err)
	}
}

func TestRunStopsAtTheTimeoutAndReportsIt(t *testing.T) {
	withMode(t, "hang")
	exec := runner.New(nil)
	exec.Timeout = 300 * time.Millisecond

	started := time.Now()
	result, err := exec.Run(context.Background(), helperArgv(), t.TempDir())

	if err != nil {
		t.Fatal(err)
	}
	if !result.TimedOut {
		t.Fatalf("result = %+v", result)
	}
	if time.Since(started) > 10*time.Second {
		t.Fatal("the timeout did not stop the process")
	}
}

func TestRunKillsTheWholeTreeNotJustTheParent(t *testing.T) {
	withMode(t, "spawn")
	exec := runner.New(nil)
	exec.Timeout = 500 * time.Millisecond
	exec.Delay = 30 * time.Second

	started := time.Now()
	result, err := exec.Run(context.Background(), helperArgv(), t.TempDir())

	if err != nil || !result.TimedOut {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	if elapsed := time.Since(started); elapsed > 10*time.Second {
		t.Fatalf("a surviving child held the output for %s", elapsed)
	}
}

func TestRunStopsWhenTheCallerCancels(t *testing.T) {
	withMode(t, "hang")
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(300*time.Millisecond, cancel)

	_, err := runner.New(nil).Run(ctx, helperArgv(), t.TempDir())

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

package benchmark_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/testing/benchmark"
	"github.com/ronz204/beacon/testing/utilities"
)

const relativeRecord = ".beacon/records/capabilities/bc-a1b1-discounts.yaml"

func beaconIn(t *testing.T) benchmark.Beacon {
	t.Helper()

	return func(directory string, args ...string) (string, error) {
		result := utilities.Beacon(t, directory, args...)
		if result.Code != 0 {
			return result.Stdout, fmt.Errorf(
				"beacon %v exited %d: %s%s",
				args, result.Code, result.Stdout, result.Stderr,
			)
		}

		return result.Stdout, nil
	}
}

func task(t *testing.T, name string) benchmark.Task {
	t.Helper()

	loaded, err := benchmark.LoadTask(
		filepath.Join("testdata", "tasks", name, "task.json"),
	)
	if err != nil {
		t.Fatal(err)
	}

	return loaded
}

func replay(t *testing.T, name string) benchmark.Session {
	t.Helper()

	agent := benchmark.Replay{
		Transcript: filepath.Join("testdata", "transcripts", name+".jsonl"),
	}

	stream, err := agent.Run(context.Background(), "", benchmark.Task{})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	session, err := benchmark.ParseTranscript(stream)
	if err != nil {
		t.Fatal(err)
	}

	return session
}

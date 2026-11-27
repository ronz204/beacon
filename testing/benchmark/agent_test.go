package benchmark_test

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/testing/benchmark"
)

func TestAReplayNeverCallsAnAgent(t *testing.T) {
	agent := benchmark.Replay{
		Transcript: filepath.Join("testdata", "transcripts", "missing.jsonl"),
	}

	stream, err := agent.Run(context.Background(), "", benchmark.Task{})
	if err == nil {
		_, _ = io.Copy(io.Discard, stream)
		t.Fatal("a missing transcript should fail")
	}
}

package action_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/internal/spider/internal/action"
)

const benchmarkPromises = 10

func benchmarkCapability(index int) string {
	var text strings.Builder

	fmt.Fprintf(&text, "id: bc-%04x\ncreated: 2026-11-22\n", index)
	text.WriteString("author: \"agent:sample-agent\"\nname: Sample\n")
	text.WriteString("intent: Sample capability for the load benchmark.\n")
	fmt.Fprintf(&text, "paths: [src/module%d/**]\npromises:\n", index)

	for promise := range benchmarkPromises {
		id := index*benchmarkPromises + promise + 0x10000
		fmt.Fprintf(&text, "  - id: bc-%05x\n", id)
		text.WriteString("    text: Sample promise with a realistic length.\n")
		text.WriteString("    check: {runner: test-go, target: sample}\n")
		text.WriteString("    status: active\n")
	}

	return text.String()
}

func benchmarkRoot(b *testing.B, files int) string {
	b.Helper()

	root := b.TempDir()
	folder := filepath.Join(root, "records", "capabilities")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		b.Fatal(err)
	}

	for index := range files {
		name := fmt.Sprintf("bc-%04x-sample.yaml", index)
		text := []byte(benchmarkCapability(index))
		path := filepath.Join(folder, name)

		if err := os.WriteFile(path, text, 0o644); err != nil {
			b.Fatal(err)
		}
	}

	return root
}

func BenchmarkLoadAndValidate500(b *testing.B) {
	root := benchmarkRoot(b, 500)
	ctx := context.Background()

	for b.Loop() {
		snapshot := action.Load(ctx, repositories(root))
		_, failures := action.Validate(snapshot)

		if len(snapshot.Failures) != 0 || len(failures) != 0 {
			b.Fatalf("load = %v, graph = %v", snapshot.Failures, failures)
		}
	}
}

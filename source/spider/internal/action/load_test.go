package action_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func repositories(root string) action.Repositories {
	store := record.NewStore(root)

	return action.Repositories{
		Goals:        store,
		Capabilities: store,
		Decisions:    store,
		Questions:    store,
		Guides:       store,
		Works:        store,
		Tasks:        store,
	}
}

func TestLoadReadsEveryRecordType(t *testing.T) {
	root := filepath.Join("..", "driver", "record", "testdata", "valid")

	snapshot := action.Load(context.Background(), repositories(root))

	if len(snapshot.Failures) != 0 {
		t.Fatalf("failures = %v", snapshot.Failures)
	}

	counts := map[string][2]int{
		"goals":        {len(snapshot.Goals), 1},
		"capabilities": {len(snapshot.Capabilities), 1},
		"decisions":    {len(snapshot.Decisions), 1},
		"questions":    {len(snapshot.Questions), 1},
		"guides":       {len(snapshot.Guides), 1},
		"works":        {len(snapshot.Works), 1},
		"tasks":        {len(snapshot.Tasks), 2},
	}
	for name, count := range counts {
		if count[0] != count[1] {
			t.Fatalf("%s = %d, want %d", name, count[0], count[1])
		}
	}
}

func TestLoadGathersFailuresFromEveryType(t *testing.T) {
	root := t.TempDir()
	broken := map[string]string{
		"goals":  "bc-aaaa-broken.yaml",
		"tasks":  "bc-bbbb.1-broken.yaml",
		"guides": "bc-cccc-broken.md",
	}

	for folder, name := range broken {
		directory := filepath.Join(root, "records", folder)
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte("a: [x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	snapshot := action.Load(context.Background(), repositories(root))

	if len(snapshot.Failures) != len(broken) {
		t.Fatalf("failures = %v", snapshot.Failures)
	}
}

package action_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/action"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
)

func writeRecords(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, text := range files {
		path := filepath.Join(root, "records", filepath.FromSlash(name))

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func taskText(id, parent, status, extra string) string {
	return "id: " + id + "\ncreated: 2026-11-22\n" +
		"author: \"human:jane\"\ntitle: Sample\nparent: " + parent + "\n" +
		extra + "status: " + status + "\n"
}

func workText(id string) string {
	return "id: " + id + "\ncreated: 2026-11-22\n" +
		"author: \"human:jane\"\nrequest: Sample\nstatus: active\n"
}

func validateRoot(t *testing.T, root string) (thread.Graph, []error) {
	t.Helper()

	snapshot := action.Load(context.Background(), repositories(root))
	if len(snapshot.Failures) != 0 {
		t.Fatalf("load failures = %v", snapshot.Failures)
	}

	return action.Validate(snapshot)
}

func requireGraphCode(t *testing.T, failures []error, code thread.GraphCode) {
	t.Helper()

	for _, failure := range failures {
		var located thread.GraphError
		if errors.As(failure, &located) && located.Code == code {
			return
		}
	}

	t.Fatalf("code %q not found in %v", code, failures)
}

func TestValidateAcceptsTheValidProject(t *testing.T) {
	root := filepath.Join("..", "driver", "record", "testdata", "valid")

	graph, failures := validateRoot(t, root)

	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	promise, err := kernel.ParseRecordID("bc-c3d4")
	if err != nil {
		t.Fatal(err)
	}
	if neighbors := graph.Neighbors(promise); len(neighbors) != 2 {
		t.Fatalf("neighbors = %v", neighbors)
	}
}

func TestValidateReportsMissingReference(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"tasks/bc-91cd.1-sample.yaml": taskText(
			"bc-91cd.1",
			"bc-91cd",
			"open",
			"",
		),
	})

	_, failures := validateRoot(t, root)

	requireGraphCode(t, failures, thread.GraphTargetMissing)
}

func TestValidateReportsTaskCycle(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"works/bc-91cd-sample.yaml": workText("bc-91cd"),
		"tasks/bc-91cd.1-sample.yaml": taskText(
			"bc-91cd.1",
			"bc-91cd",
			"open",
			"blocks: [bc-91cd.2]\n",
		),
		"tasks/bc-91cd.2-sample.yaml": taskText(
			"bc-91cd.2",
			"bc-91cd",
			"open",
			"blocks: [bc-91cd.1]\n",
		),
	})

	_, failures := validateRoot(t, root)

	requireGraphCode(t, failures, thread.GraphCycle)
}

func TestValidateReportsHierarchyRules(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"works/bc-91cd-sample.yaml": workText("bc-91cd"),
		"tasks/bc-91cd.1-sample.yaml": taskText(
			"bc-91cd.1",
			"bc-91cd",
			"done",
			"",
		),
		"tasks/bc-91cd.1.1-sample.yaml": taskText(
			"bc-91cd.1.1",
			"bc-91cd",
			"open",
			"",
		),
		"tasks/bc-91cd.2.1-sample.yaml": taskText(
			"bc-91cd.2.1",
			"bc-91cd",
			"open",
			"",
		),
	})

	_, failures := validateRoot(t, root)

	requireGraphCode(t, failures, thread.GraphTaskClosedEarly)
	requireGraphCode(t, failures, thread.GraphTaskOrphan)
}

func TestValidateCoversEveryRelationField(t *testing.T) {
	const missing = "bc-ffff"
	root := writeRecords(t, map[string]string{
		"goals/bc-0001-sample.yaml": "id: bc-0001\ncreated: 2026-11-22\n" +
			"author: \"human:jane\"\nstatement: Sample\nsuccess: Sample\n" +
			"status: active\n",
		"capabilities/bc-7f3a-sample.yaml": "id: bc-7f3a\n" +
			"created: 2026-11-22\nauthor: \"human:jane\"\nname: Sample\n" +
			"intent: Sample\nserves: " + missing + "\npromises:\n" +
			"  - id: bc-c3d4\n    text: Sample\n" +
			"    check: {manual: Open it}\n    status: proposed\n" +
			"    origin: " + missing + "\n    refers: [" + missing + "]\n",
		"decisions/bc-a1b2-sample.yaml": "id: bc-a1b2\ncreated: 2026-11-22\n" +
			"author: \"human:jane\"\nwhat: Sample\nwhy: Sample\n" +
			"supersedes: " + missing + "\nstatus: active\n",
		"questions/bc-e5f6-sample.yaml": "id: bc-e5f6\ncreated: 2026-11-22\n" +
			"author: \"human:jane\"\ntext: Sample\n" +
			"blocks: [" + missing + "]\n" +
			"status: open\norigin: " + missing + "\n",
		"guides/bc-0a9b-sample.md": "---\ntitle: Sample\nkind: domain\n" +
			"refers: [" + missing + "]\n---\nBody\n",
		"works/bc-91cd-sample.yaml": "id: bc-91cd\ncreated: 2026-11-22\n" +
			"author: \"human:jane\"\nrequest: Sample\n" +
			"implements: [" + missing + "]\nserves: " + missing + "\n" +
			"status: active\n",
		"tasks/bc-91cd.1-sample.yaml": taskText(
			"bc-91cd.1",
			"bc-91cd",
			"open",
			"blocks: [bc-ffff.1]\nimplements: "+missing+"\n",
		),
	})

	_, failures := validateRoot(t, root)

	if len(failures) != 11 {
		t.Fatalf("failures = %d: %v", len(failures), failures)
	}
	for _, failure := range failures {
		var located thread.GraphError
		if !errors.As(failure, &located) {
			t.Fatalf("unexpected failure: %v", failure)
		}
		if located.Code != thread.GraphTargetMissing {
			t.Fatalf("code = %q", located.Code)
		}
	}
}

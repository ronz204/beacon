package record_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
	"github.com/ronz204/beacon/internal/spider/internal/driver/record"
)

const (
	caps = "capabilities"
	asks = "questions"
)

var sampleNames = map[string]string{
	"goals":     "bc-0001-sample.yaml",
	caps:        "bc-7f3a-sample.yaml",
	"decisions": "bc-a1b2-sample.yaml",
	"questions": "bc-e5f6-sample.yaml",
	"works":     "bc-91cd-sample.yaml",
	"tasks":     "bc-91cd.1-sample.yaml",
	"guides":    "bc-0a9b-sample.md",
}

func is[C ~string](code C) func(error) bool {
	return func(err error) bool {
		var got kernel.Error[C]
		return errors.As(err, &got) && got.Code == code
	}
}

func writeRecord(t *testing.T, root, folder, name string, data []byte) {
	t.Helper()

	directory := filepath.Join(root, "records", folder)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func listFailures(store record.Store, folder string) []error {
	ctx := context.Background()

	switch folder {
	case "goals":
		_, failures := store.ListGoals(ctx)
		return failures
	case caps:
		_, failures := store.ListCapabilities(ctx)
		return failures
	case "decisions":
		_, failures := store.ListDecisions(ctx)
		return failures
	case "questions":
		_, failures := store.ListQuestions(ctx)
		return failures
	case "works":
		_, failures := store.ListWorks(ctx)
		return failures
	case "tasks":
		_, failures := store.ListTasks(ctx)
		return failures
	}

	_, failures := store.ListGuides(ctx)
	return failures
}

func TestStoreLoadsEveryValidType(t *testing.T) {
	ctx := context.Background()
	store := record.NewStore(filepath.Join("testdata", "valid"))

	goals, failures := store.ListGoals(ctx)
	requireLoaded(t, len(goals), 1, failures)

	capabilities, failures := store.ListCapabilities(ctx)
	requireLoaded(t, len(capabilities), 1, failures)

	decisions, failures := store.ListDecisions(ctx)
	requireLoaded(t, len(decisions), 1, failures)

	questions, failures := store.ListQuestions(ctx)
	requireLoaded(t, len(questions), 1, failures)

	works, failures := store.ListWorks(ctx)
	requireLoaded(t, len(works), 1, failures)

	tasks, failures := store.ListTasks(ctx)
	requireLoaded(t, len(tasks), 2, failures)

	guides, failures := store.ListGuides(ctx)
	requireLoaded(t, len(guides), 1, failures)

	promises := capabilities[0].Promises()
	if len(promises) != 2 {
		t.Fatalf("promises = %d", len(promises))
	}
	if _, found := promises[1].Attested(); !found {
		t.Fatal("second promise should be attested")
	}
	if !promises[1].Check().IsManual() {
		t.Fatal("second promise should be manual")
	}
	if promises[0].Check().Params()["target"] != "auth/lockout" {
		t.Fatalf("params = %v", promises[0].Check().Params())
	}
	if guides[0].ID().String() != "bc-0a9b" {
		t.Fatalf("guide id = %q", guides[0].ID())
	}
	if !guides[0].Always() {
		t.Fatal("guide should be always")
	}
	if tasks[0].ID().String() != "bc-91cd.1" {
		t.Fatalf("first task = %q", tasks[0].ID())
	}

	wantRequest := "Migrar el almacenamiento a Postgres\n" +
		"sin perder los datos actuales.\n"
	if works[0].Request().String() != wantRequest {
		t.Fatalf("request = %q", works[0].Request())
	}
}

func requireLoaded(t *testing.T, got, want int, failures []error) {
	t.Helper()

	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}
	if got != want {
		t.Fatalf("loaded = %d, want %d", got, want)
	}
}

func TestStoreReportsInvalidFixtures(t *testing.T) {
	cases := []struct {
		folder  string
		name    string
		matches func(error) bool
	}{
		{"goals", "status_unknown", is(ledger.GoalStatusInvalid)},
		{"goals", "missing_statement", is(ledger.GoalFieldMissing)},
		{"goals", "id_child", is(ledger.GoalIDInvalid)},
		{"goals", "date_impossible", is(kernel.KernelDateInvalid)},
		{"goals", "author_malformed", is(kernel.KernelActorInvalid)},
		{caps, "paths_not_list", is(record.RecordFieldType)},
		{caps, "promises_missing", is(record.RecordFieldMissing)},
		{caps, "promises_not_list", is(record.RecordFieldType)},
		{caps, "check_not_map", is(record.RecordFieldType)},
		{caps, "glob_negated", is(kernel.KernelGlobInvalid)},
		{caps, "promise_repeated", is(ledger.PromiseRepeated)},
		{caps, "runner_and_manual", is(ledger.PromiseCheckInvalid)},
		{caps, "param_option", is(ledger.PromiseCheckInvalid)},
		{caps, "attested_automatic", is(ledger.PromiseAttestManual)},
		{caps, "attested_by_agent", is(ledger.PromiseAttestInvalid)},
		{"decisions", "status_unknown", is(memory.DecisionStatusInvalid)},
		{"decisions", "review_invalid", is(memory.ReviewSpanInvalid)},
		{asks, "answered_without_answer", is(memory.QuestionAnswerMissing)},
		{"works", "request_blank", is(agenda.WorkRequestInvalid)},
		{"tasks", "parent_mismatch", is(agenda.TaskParentMismatch)},
		{"tasks", "claimed_without_owner", is(agenda.TaskOwnerMissing)},
		{"tasks", "branch_option", is(agenda.TaskBranchInvalid)},
		{"guides", "kind_unknown", is(memory.GuideKindInvalid)},
		{"guides", "missing_title", is(memory.GuideFieldMissing)},
	}

	for _, c := range cases {
		t.Run(c.folder+"/"+c.name, func(t *testing.T) {
			target := sampleNames[c.folder]
			source := c.name + filepath.Ext(target)
			path := filepath.Join("testdata", "invalid", c.folder, source)
			root := t.TempDir()
			writeRecord(t, root, c.folder, target, readFixture(t, path))

			failures := listFailures(record.NewStore(root), c.folder)

			for _, failure := range failures {
				if c.matches(failure) {
					return
				}
			}
			t.Fatalf("expected failure not found in %v", failures)
		})
	}
}

func TestStoreRejectsIDDifferentFromFileName(t *testing.T) {
	const name = "bc-ffff-id-distinto-al-archivo.yaml"
	path := filepath.Join("testdata", "invalid", "goals", name)
	root := t.TempDir()
	writeRecord(t, root, "goals", name, readFixture(t, path))

	failures := listFailures(record.NewStore(root), "goals")

	if len(failures) != 1 || !is(record.RecordFileMismatch)(failures[0]) {
		t.Fatalf("failures = %v", failures)
	}
}

func TestStoreRejectsGuideNameWithoutID(t *testing.T) {
	data := []byte("---\ntitle: Sample\nkind: domain\n---\nBody\n")
	root := t.TempDir()
	writeRecord(t, root, "guides", "sample.md", data)

	failures := listFailures(record.NewStore(root), "guides")

	if len(failures) != 1 || !is(record.RecordFileMismatch)(failures[0]) {
		t.Fatalf("failures = %v", failures)
	}
}

func TestStoreKeepsLoadingAfterBrokenFiles(t *testing.T) {
	valid := readFixture(t, filepath.Join(
		"testdata", "valid", "records", "goals",
		"bc-0001-pagos-sin-friccion.yaml",
	))
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-0001-sample.yaml", valid)
	writeRecord(t, root, "goals", "bc-aaaa-broken.yaml", []byte("a: [x\n"))
	writeRecord(t, root, "goals", "bc-bbbb-broken.yaml", []byte("- item\n"))

	goals, failures := record.NewStore(root).ListGoals(context.Background())

	if len(goals) != 1 {
		t.Fatalf("goals = %d", len(goals))
	}
	if len(failures) != 2 {
		t.Fatalf("failures = %v", failures)
	}
}

func TestStoreReportsEveryFieldOfAFile(t *testing.T) {
	data := []byte("id: bc-0001\ncreated: nope\nauthor: jane\nstatus: x\n")
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-0001-sample.yaml", data)

	failures := listFailures(record.NewStore(root), "goals")

	if len(failures) != 3 {
		t.Fatalf("failures = %v", failures)
	}

	var located kernel.RecordFailure
	if !errors.As(failures[0], &located) {
		t.Fatalf("expected RecordFailure, got %v", failures[0])
	}
	if located.Details["line"] != 2 {
		t.Fatalf("line = %v", located.Details["line"])
	}
}

func TestStoreReturnsFilesInNameOrder(t *testing.T) {
	data := func(id string) []byte {
		return []byte("id: " + id + "\ncreated: 2026-11-22\n" +
			"author: human:jane\nstatement: Sample\n" +
			"success: Sample\nstatus: active\n")
	}
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-cccc-third.yaml", data("bc-cccc"))
	writeRecord(t, root, "goals", "bc-aaaa-first.yaml", data("bc-aaaa"))
	writeRecord(t, root, "goals", "bc-bbbb-second.yaml", data("bc-bbbb"))

	goals, failures := record.NewStore(root).ListGoals(context.Background())

	requireLoaded(t, len(goals), 3, failures)
	want := []string{"bc-aaaa", "bc-bbbb", "bc-cccc"}
	for index, goal := range goals {
		if goal.ID().String() != want[index] {
			t.Fatalf("goal %d = %q", index, goal.ID())
		}
	}
}

func TestStoreTreatsMissingFolderAsEmpty(t *testing.T) {
	goals, failures := record.NewStore(t.TempDir()).ListGoals(
		context.Background(),
	)

	requireLoaded(t, len(goals), 0, failures)
}

func TestStoreHandlesForeignFiles(t *testing.T) {
	root := t.TempDir()
	writeRecord(t, root, "goals", ".gitkeep", nil)

	failures := listFailures(record.NewStore(root), "goals")
	if failures != nil {
		t.Fatalf("hidden file should be ignored: %v", failures)
	}

	writeRecord(t, root, "goals", "bc-0001-sample.yml", []byte("a: 1\n"))

	failures = listFailures(record.NewStore(root), "goals")

	if len(failures) != 1 || !is(record.RecordFileExtension)(failures[0]) {
		t.Fatalf("failures = %v", failures)
	}
}

func TestStoreHonorsCancelledContext(t *testing.T) {
	path := filepath.Join(
		"testdata", "valid", "records", "goals",
		"bc-0001-pagos-sin-friccion.yaml",
	)
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-0001-sample.yaml", readFixture(t, path))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	goals, failures := record.NewStore(root).ListGoals(ctx)

	if len(goals) != 0 || len(failures) != 1 {
		t.Fatalf("goals = %d, failures = %v", len(goals), failures)
	}
	if !errors.Is(failures[0], context.Canceled) {
		t.Fatalf("failure = %v", failures[0])
	}
}

func TestStoreAcceptsUnknownFields(t *testing.T) {
	cases := []struct {
		folder string
		source string
	}{
		{"goals", "goal.yaml"},
		{caps, "capability.yaml"},
		{"guides", "guide.md"},
	}

	for _, c := range cases {
		t.Run(c.folder, func(t *testing.T) {
			path := filepath.Join("testdata", "unknown", c.source)
			root := t.TempDir()
			target := sampleNames[c.folder]
			writeRecord(t, root, c.folder, target, readFixture(t, path))

			failures := listFailures(record.NewStore(root), c.folder)

			if len(failures) != 0 {
				t.Fatalf("failures = %v", failures)
			}
		})
	}
}

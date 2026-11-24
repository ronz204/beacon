package action_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/internal/spider/internal/action"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
)

const (
	workFile    = "works/bc-91cd-sample.yaml"
	claimFields = "owner: \"agent:sample-agent\"\nbranch: "
)

func questionText(status string) string {
	text := "id: bc-e5f6\ncreated: 2026-11-22\nauthor: \"human:jane\"\n" +
		"text: Sample\nblocks: [bc-91cd]\nstatus: " + status + "\n"

	if status == "answered" {
		text += "answer: Five attempts\n"
	}

	return text
}

func openTask(id string) string {
	return taskText(id, "bc-91cd", "open", "")
}

func taskFile(id string) string {
	return "tasks/" + id + "-sample.yaml"
}

func readyOf(
	t *testing.T,
	files map[string]string,
	branch string,
) action.ReadyResult {
	t.Helper()

	var parsed agenda.TaskBranch
	if branch != "" {
		var err error
		if parsed, err = agenda.ParseTaskBranch(branch); err != nil {
			t.Fatal(err)
		}
	}

	root := writeRecords(t, files)
	result, failures := action.Ready(
		context.Background(),
		repositories(root),
		parsed,
	)
	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	return result
}

func readyTaskIDs(result action.ReadyResult) []string {
	ids := make([]string, 0, len(result.Tasks))

	for _, task := range result.Tasks {
		ids = append(ids, task.ID().String())
	}

	return ids
}

func requireReady(t *testing.T, result action.ReadyResult, want ...string) {
	t.Helper()

	got := readyTaskIDs(result)
	if !slices.Equal(got, want) {
		t.Fatalf("ready = %v, want %v", got, want)
	}
}

func TestReadyHidesWorkBlockedByOpenQuestion(t *testing.T) {
	files := map[string]string{
		workFile:                        workText("bc-91cd"),
		taskFile("bc-91cd.1"):           openTask("bc-91cd.1"),
		"questions/bc-e5f6-sample.yaml": questionText("open"),
	}

	blocked := readyOf(t, files, "")
	if len(blocked.Works) != 0 {
		t.Fatalf("works = %d", len(blocked.Works))
	}
	requireReady(t, blocked)

	files["questions/bc-e5f6-sample.yaml"] = questionText("answered")

	answered := readyOf(t, files, "")
	if len(answered.Works) != 1 {
		t.Fatalf("works = %d", len(answered.Works))
	}
	requireReady(t, answered, "bc-91cd.1")
}

func TestReadyWaitsForBlockingTask(t *testing.T) {
	blocks := "blocks: [bc-91cd.2]\n"
	files := map[string]string{
		workFile:              workText("bc-91cd"),
		taskFile("bc-91cd.1"): taskText("bc-91cd.1", "bc-91cd", "open", blocks),
		taskFile("bc-91cd.2"): openTask("bc-91cd.2"),
	}

	requireReady(t, readyOf(t, files, ""), "bc-91cd.1")

	files[taskFile("bc-91cd.1")] = taskText(
		"bc-91cd.1",
		"bc-91cd",
		"done",
		blocks,
	)

	requireReady(t, readyOf(t, files, ""), "bc-91cd.2")
}

func TestReadyInheritsBlockFromParentTask(t *testing.T) {
	question := strings.Replace(
		questionText("open"),
		"blocks: [bc-91cd]",
		"blocks: [bc-91cd.1]",
		1,
	)
	files := map[string]string{
		workFile:                        workText("bc-91cd"),
		taskFile("bc-91cd.1"):           openTask("bc-91cd.1"),
		taskFile("bc-91cd.1.1"):         openTask("bc-91cd.1.1"),
		taskFile("bc-91cd.2"):           openTask("bc-91cd.2"),
		"questions/bc-e5f6-sample.yaml": question,
	}

	requireReady(t, readyOf(t, files, ""), "bc-91cd.2")
}

func TestReadyOffersLeavesOnly(t *testing.T) {
	files := map[string]string{
		workFile:                workText("bc-91cd"),
		taskFile("bc-91cd.1"):   openTask("bc-91cd.1"),
		taskFile("bc-91cd.1.1"): openTask("bc-91cd.1.1"),
	}

	requireReady(t, readyOf(t, files, ""), "bc-91cd.1.1")

	files[taskFile("bc-91cd.1.1")] = taskText(
		"bc-91cd.1.1",
		"bc-91cd",
		"done",
		"",
	)

	requireReady(t, readyOf(t, files, ""), "bc-91cd.1")
}

func TestReadyRespectsClaims(t *testing.T) {
	files := map[string]string{
		workFile: workText("bc-91cd"),
		taskFile("bc-91cd.1"): taskText(
			"bc-91cd.1",
			"bc-91cd",
			"claimed",
			claimFields+"feat/sample-a\n",
		),
		taskFile("bc-91cd.2"): taskText(
			"bc-91cd.2",
			"bc-91cd",
			"claimed",
			claimFields+"feat/sample-b\n",
		),
		taskFile("bc-91cd.3"): taskText("bc-91cd.3", "bc-91cd", "done", ""),
	}

	requireReady(t, readyOf(t, files, "feat/sample-a"), "bc-91cd.1")
	requireReady(t, readyOf(t, files, ""))
}

func TestReadySkipsTasksOfInactiveWork(t *testing.T) {
	work := strings.Replace(
		workText("bc-91cd"),
		"status: active",
		"status: done",
		1,
	)
	files := map[string]string{
		workFile:              work,
		taskFile("bc-91cd.1"): openTask("bc-91cd.1"),
	}

	result := readyOf(t, files, "")

	if len(result.Works) != 0 {
		t.Fatalf("works = %d", len(result.Works))
	}
	requireReady(t, result)
}

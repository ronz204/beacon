package spider_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
)

const reason = "Replaced by the new sign-in flow."

func closeRecord(
	t *testing.T,
	project spider.Project,
	id string,
	stamps map[string]spider.Approval,
) spider.ClosePrepared {
	t.Helper()

	prepared, err := project.PrepareClose(context.Background(), id, reason)
	if err != nil {
		t.Fatalf("prepare close: %+v", kernel.DescribeFailure(err))
	}

	err = project.Close(context.Background(), prepared, stamps)
	if err != nil {
		t.Fatalf("close: %+v", kernel.DescribeFailure(err))
	}

	return prepared
}

func requireNoGraphFindings(t *testing.T, project spider.Project) {
	t.Helper()

	for _, finding := range project.Doctor(context.Background()) {
		if strings.HasPrefix(finding.Code, "graph_") {
			t.Fatalf("finding = %+v", finding)
		}
	}
}

func archiveLines(t *testing.T, root string) []string {
	t.Helper()

	matches, err := filepath.Glob(
		filepath.Join(root, ".beacon", "archive", "*.jsonl"),
	)
	if err != nil || len(matches) != 1 {
		t.Fatalf("segments = %v, %v", matches, err)
	}

	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}

	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func TestCloseKeepsReferencesResolvable(t *testing.T) {
	project, root := newProject(t)

	closeRecord(t, project, "bc-7f3a", nil)

	requireNoGraphFindings(t, project)
	if _, err := os.Stat(filepath.Join(root, capabilityFile)); err == nil {
		t.Fatal("the capability file is still in records")
	}

	stored, err := project.Get(context.Background(), "bc-7f3a")
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Archived || stored.Type != "capability" {
		t.Fatalf("stored = %+v", stored)
	}
	if got := recordField(t, stored.Record, "reason"); got != reason {
		t.Fatalf("reason = %v", got)
	}
}

func TestClosePromiseLeavesItsCapability(t *testing.T) {
	project, root := newProject(t)

	prepared := closeRecord(t, project, "bc-c4a1", map[string]spider.Approval{
		"bc-c4a1": {By: "human:jane", At: "2026-11-25"},
	})

	if prepared.Type != "promise" || prepared.After == nil {
		t.Fatalf("prepared = %+v", prepared)
	}
	if strings.Contains(readFile(t, root, capabilityFile), "bc-c4a1") {
		t.Fatal("the promise is still in its capability")
	}

	stored, err := project.Get(context.Background(), "bc-c4a1")
	if err != nil {
		t.Fatal(err)
	}
	if got := recordField(t, stored.Record, "capability"); got != "bc-7f3a" {
		t.Fatalf("capability = %v", got)
	}
	approved, _ := recordField(t, stored.Record, "approved").(map[string]any)
	if approved["by"] != "human:jane" {
		t.Fatalf("approved = %v", approved)
	}
	requireNoGraphFindings(t, project)
}

func TestConcurrentClosesKeepEveryLine(t *testing.T) {
	project, root := newProject(t)
	ids := []string{"bc-e5f6", "bc-a1b2", "bc-0001"}

	var group sync.WaitGroup
	for _, id := range ids {
		group.Go(func() {
			prepared, err := project.PrepareClose(
				context.Background(),
				id,
				reason,
			)
			if err == nil {
				err = project.Close(context.Background(), prepared, nil)
			}
			if err != nil {
				t.Errorf("%s: %v", id, err)
			}
		})
	}
	group.Wait()

	if lines := archiveLines(t, root); len(lines) != len(ids) {
		t.Fatalf("lines = %v", lines)
	}
}

func TestCloseRefusesGuidesAndTasksWithOpenSubtasks(t *testing.T) {
	project, _ := newProject(t)

	_, err := project.PrepareClose(context.Background(), "bc-0a9b", reason)
	requireCode(t, err, string(spider.SaveGuideNotArchivable))

	_, err = project.PrepareClose(context.Background(), "bc-91cd.1", reason)
	view := requireCode(t, err, string(spider.SaveTaskHasOpenChildren))
	if subtasks, _ := view.Details["subtasks"].([]string); len(subtasks) != 1 {
		t.Fatalf("details = %v", view.Details)
	}
}

func TestFindSearchesTheArchiveOnRequest(t *testing.T) {
	project, _ := newProject(t)
	closeRecord(t, project, "bc-a1b2", nil)

	active, _ := project.Find(context.Background(), "jwt", false)
	all, _ := project.Find(context.Background(), "jwt", true)

	if len(active.Matches) != 0 || len(all.Matches) != 1 {
		t.Fatalf("active = %+v, all = %+v", active, all)
	}
}

func TestDoctorCompletesAnInterruptedClose(t *testing.T) {
	project, root := newProject(t)
	closeRecord(t, project, "bc-7f3a", nil)
	runGit(t, root, "checkout", "--", capabilityFile)

	findings := project.Doctor(context.Background())

	if _, err := os.Stat(filepath.Join(root, capabilityFile)); err == nil {
		t.Fatal("the interrupted close left the capability in records")
	}
	completed := 0
	for _, finding := range findings {
		if finding.Code == "close_completed" {
			completed++
		}
		if strings.HasPrefix(finding.Code, "graph_") {
			t.Fatalf("finding = %+v", finding)
		}
	}
	if completed != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDoctorCompletesAnInterruptedPromiseClose(t *testing.T) {
	project, root := newProject(t)
	closeRecord(t, project, "bc-c4a1", nil)
	runGit(t, root, "checkout", "--", capabilityFile)

	project.Doctor(context.Background())

	if strings.Contains(readFile(t, root, capabilityFile), "bc-c4a1") {
		t.Fatal("the interrupted close left the promise in its capability")
	}
}

func TestStateChangesReportTheApprovalOfAClosure(t *testing.T) {
	project, _ := newProject(t)
	closeRecord(t, project, "bc-c4a1", map[string]spider.Approval{
		"bc-c4a1": {By: "human:jane", At: "2026-11-25"},
	})

	changes, err := project.StateChanges(context.Background(), "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	var closed []spider.Closure
	for _, change := range changes {
		closed = append(closed, change.Closed...)
	}

	if len(closed) != 1 || closed[0].ID != "bc-c4a1" {
		t.Fatalf("closed = %+v", closed)
	}
	if !strings.Contains(closed[0].Approved, "human:jane") {
		t.Fatalf("approved = %q", closed[0].Approved)
	}
}

func TestStateChangesListThePromisesOfAClosedCapability(t *testing.T) {
	project, _ := newProject(t)
	closeRecord(t, project, "bc-7f3a", map[string]spider.Approval{
		"bc-7f3a": {By: "human:jane", At: "2026-11-25"},
	})

	changes, err := project.StateChanges(context.Background(), "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	var promises []string
	for _, change := range changes {
		for _, closed := range change.Closed {
			promises = append(promises, closed.Promises...)
		}
	}

	if len(promises) != 2 || promises[0] != "bc-c3d4" {
		t.Fatalf("promises = %v", promises)
	}
}

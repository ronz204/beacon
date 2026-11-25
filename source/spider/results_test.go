package spider_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
)

const loginFile = "src/auth/login.go"

func writeSource(t *testing.T, root, file, text string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(file))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func recordPassed(t *testing.T, project spider.Project, promises ...string) {
	t.Helper()

	verdicts := make([]spider.Verdict, 0, len(promises))
	for _, promise := range promises {
		verdicts = append(verdicts, spider.Verdict{
			Promise: promise,
			Outcome: "passed",
		})
	}

	err := project.RecordResults(context.Background(), verdicts)
	if err != nil {
		t.Fatalf("record results: %+v", kernel.DescribeFailure(err))
	}
}

func promiseState(t *testing.T, project spider.Project) spider.PromiseView {
	t.Helper()

	analysis, findings := project.Analyze(context.Background())
	requireNoFindings(t, findings)

	for _, promise := range analysis.Promises {
		if promise.ID == "bc-c3d4" {
			return promise
		}
	}

	t.Fatal("promise bc-c3d4 missing")
	return spider.PromiseView{}
}

func TestCheckWithUncommittedChangesHasNoDrift(t *testing.T) {
	project, root := newProject(t)
	writeSource(t, root, loginFile, "package auth\n")

	recordPassed(t, project, "bc-c3d4")

	promise := promiseState(t, project)
	if promise.State != "holds" || promise.Drifted || promise.NotRunHere {
		t.Fatalf("promise = %+v", promise)
	}
}

func TestEditingAfterTheCheckDrifts(t *testing.T) {
	project, root := newProject(t)
	writeSource(t, root, loginFile, "package auth\n")
	recordPassed(t, project, "bc-c3d4")

	writeSource(t, root, loginFile, "package auth\n\nvar locked = true\n")

	if promise := promiseState(t, project); !promise.Drifted {
		t.Fatalf("promise = %+v", promise)
	}
}

func TestCommittingWhatWasCheckedHasNoDrift(t *testing.T) {
	project, root := newProject(t)
	writeSource(t, root, loginFile, "package auth\n")
	recordPassed(t, project, "bc-c3d4")

	runGit(t, root, "add", loginFile)
	runGit(t, root, "commit", "-q", "-m", "login")

	if promise := promiseState(t, project); promise.Drifted {
		t.Fatalf("promise = %+v", promise)
	}
}

func TestConcurrentResultsAreAllKept(t *testing.T) {
	project, root := newProject(t)

	var group sync.WaitGroup
	for _, promise := range []string{"bc-c3d4", "bc-c4a1"} {
		group.Go(func() {
			verdict := spider.Verdict{Promise: promise, Outcome: "failed"}
			err := project.RecordResults(
				context.Background(),
				[]spider.Verdict{verdict},
			)
			if err != nil {
				t.Errorf("%s: %v", promise, err)
			}
		})
	}
	group.Wait()

	data, err := os.ReadFile(
		filepath.Join(root, ".git", "beacon", "results.json"),
	)
	if err != nil {
		t.Fatal(err)
	}

	var document struct {
		Results map[string]any `json:"results"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Results) != 2 {
		t.Fatalf("results = %v", document.Results)
	}
}

func TestRecordResultsRejectsAnUnknownPromise(t *testing.T) {
	project, _ := newProject(t)

	err := project.RecordResults(context.Background(), []spider.Verdict{
		{Promise: "bc-eeee", Outcome: "passed"},
	})

	requireCode(t, err, "record_missing")
}

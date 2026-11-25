package spider_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/spider"
)

var fixtureRoot = filepath.Join(
	"internal", "driver", "record", "testdata", "valid",
)

func newProject(t *testing.T) (spider.Project, string) {
	t.Helper()

	root := t.TempDir()
	state := filepath.Join(root, ".beacon")
	if err := os.CopyFS(state, os.DirFS(fixtureRoot)); err != nil {
		t.Fatal(err)
	}

	runGit(t, root, "init", "-q", "-b", "main")
	runGit(t, root, "config", "user.name", "jane")
	runGit(t, root, "config", "user.email", "jane@example.com")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", "fixtures")
	pointAttestationAtHead(t, root)

	project, err := spider.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	return project, root
}

func pointAttestationAtHead(t *testing.T, root string) {
	t.Helper()

	command := exec.Command("git", "rev-parse", "--short", "HEAD")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(
		root, ".beacon", "records", "capabilities",
		"bc-7f3a-auth-login.yaml",
	)
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	head := strings.TrimSpace(string(output))
	edited := strings.Replace(string(data), "9f3c2ab", head, 1)
	if err := os.WriteFile(file, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	runGit(t, root, "commit", "-q", "-am", "attest at head")
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func requireNoFindings(t *testing.T, findings []spider.Finding) {
	t.Helper()

	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func recordField(t *testing.T, record spider.Record, key string) any {
	t.Helper()

	var fields map[string]any
	if err := json.Unmarshal(record.Data, &fields); err != nil {
		t.Fatalf("record JSON %s: %v", record.Data, err)
	}

	return fields[key]
}

func TestBriefStartsWithTheGoal(t *testing.T) {
	project, _ := newProject(t)

	view, findings := project.Brief(context.Background(), []string{"x"})

	requireNoFindings(t, findings)
	if view.Lines[0].Section != "goal" {
		t.Fatalf("lines = %+v", view.Lines)
	}
	if !slices.Contains(view.Lines, spider.BriefLine{
		Section: "pending",
		Text:    "Needs approval: x",
	}) {
		t.Fatalf("pending line missing: %+v", view.Lines)
	}
}

func TestContextReturnsTheOwningCapability(t *testing.T) {
	project, _ := newProject(t)

	view, findings := project.Context(
		context.Background(),
		[]string{"src/auth/login.go"},
	)

	requireNoFindings(t, findings)
	if len(view.Capabilities) != 1 || view.Capabilities[0].ID != "bc-7f3a" {
		t.Fatalf("capabilities = %+v", view.Capabilities)
	}
	if len(view.Capabilities[0].Promises) != 2 {
		t.Fatalf("promises = %+v", view.Capabilities[0].Promises)
	}
}

func TestHookContextDeliversOncePerSession(t *testing.T) {
	project, _ := newProject(t)
	paths := []string{"src/auth/login.go"}

	first, findings, err := project.HookContext(
		context.Background(),
		"session-a",
		paths,
	)
	if err != nil {
		t.Fatal(err)
	}
	requireNoFindings(t, findings)

	second, _, err := project.HookContext(
		context.Background(),
		"session-a",
		paths,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(first.Lines) == 0 || len(second.Lines) != 0 {
		t.Fatalf("first = %v, second = %v", first.Lines, second.Lines)
	}
}

func TestReadyAndAnalyzeExposeTheGraph(t *testing.T) {
	project, _ := newProject(t)

	ready, findings, err := project.Ready(
		context.Background(),
		"feat/migrate-postgres",
	)
	if err != nil {
		t.Fatal(err)
	}
	requireNoFindings(t, findings)
	if ready.Works == nil || ready.Tasks == nil {
		t.Fatalf("ready = %+v", ready)
	}

	analysis, findings := project.Analyze(context.Background())

	requireNoFindings(t, findings)
	if len(analysis.Promises) != 2 || len(analysis.Signals) == 0 {
		t.Fatalf("analysis = %+v", analysis)
	}
}

func TestFindReturnsMatchesByType(t *testing.T) {
	project, _ := newProject(t)

	view, findings := project.Find(context.Background(), "postgres", false)

	requireNoFindings(t, findings)
	if !slices.Contains(view.Matches, spider.Match{
		Type: "work",
		ID:   "bc-91cd",
		Text: "Migrate storage to Postgres",
	}) {
		t.Fatalf("matches = %+v", view.Matches)
	}
}

func TestGetReturnsThePayloadShape(t *testing.T) {
	project, _ := newProject(t)

	stored, err := project.Get(context.Background(), "bc-7f3a")
	if err != nil {
		t.Fatal(err)
	}

	if stored.Type != "capability" {
		t.Fatalf("type = %s", stored.Type)
	}
	if stored.File != ".beacon/records/capabilities/bc-7f3a-auth-login.yaml" {
		t.Fatalf("file = %s", stored.File)
	}
	if got := recordField(t, stored.Record, "serves"); got != "bc-0001" {
		t.Fatalf("serves = %v", got)
	}
}

func TestGetReturnsTheGuideBody(t *testing.T) {
	project, _ := newProject(t)

	stored, err := project.Get(context.Background(), "bc-0a9b")
	if err != nil {
		t.Fatal(err)
	}

	if got := recordField(t, stored.Record, "always"); got != true {
		t.Fatalf("always = %v", got)
	}
	body, _ := recordField(t, stored.Record, "body").(string)
	if !strings.HasPrefix(body, "Three components") {
		t.Fatalf("body = %q", body)
	}
}

func TestGetReportsAMissingRecord(t *testing.T) {
	project, _ := newProject(t)

	_, err := project.Get(context.Background(), "bc-ffff")

	if err == nil || !strings.Contains(err.Error(), "record_missing") {
		t.Fatalf("err = %v", err)
	}
}

func TestStateChangesDecodeBothSides(t *testing.T) {
	project, root := newProject(t)
	file := filepath.Join(
		root, ".beacon", "records", "questions",
		"bc-e5f6-attempt-limit.yaml",
	)
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	answered := "status: answered\nanswer: Five"
	edited := strings.Replace(string(data), "status: open", answered, 1)
	if err := os.WriteFile(file, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	changes, err := project.StateChanges(context.Background(), "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	if len(changes) != 1 || changes[0].Kind != "modified" {
		t.Fatalf("changes = %+v", changes)
	}
	if got := recordField(t, *changes[0].Before, "status"); got != "open" {
		t.Fatalf("before status = %v", got)
	}
	if got := recordField(t, *changes[0].After, "status"); got != "answered" {
		t.Fatalf("after status = %v", got)
	}
}

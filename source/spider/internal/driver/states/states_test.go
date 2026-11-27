package states_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

func newStore(t *testing.T) (states.Store, string) {
	t.Helper()

	directory := filepath.Join(t.TempDir(), "beacon")
	store, err := states.NewStore(directory)
	if err != nil {
		t.Fatal(err)
	}

	return store, directory
}

func claim(t *testing.T, task, branch string) agenda.TaskClaim {
	t.Helper()

	id, err := agenda.ParseTaskID(task)
	if err != nil {
		t.Fatal(err)
	}

	parsedBranch, err := agenda.ParseTaskBranch(branch)
	if err != nil {
		t.Fatal(err)
	}

	day, err := kernel.ParseCalendarDate("2026-11-24")
	if err != nil {
		t.Fatal(err)
	}

	return agenda.TaskClaim{
		Task:     id,
		Branch:   parsedBranch,
		Worktree: "/work/sample",
		Day:      day,
	}
}

func session(t *testing.T, raw string) kernel.SessionID {
	t.Helper()

	id, err := kernel.ParseSessionID(raw)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func TestClaimsStartEmpty(t *testing.T) {
	store, _ := newStore(t)

	claims, err := store.ListClaims(context.Background())

	if err != nil || len(claims) != 0 {
		t.Fatalf("claims = %v, err = %v", claims, err)
	}
}

func TestSaveClaimReplacesTheSameTask(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()

	first := claim(t, "bc-91cd.1", "feat/sample-a")
	second := claim(t, "bc-91cd.1", "feat/sample-b")
	other := claim(t, "bc-91cd.2", "feat/sample-a")

	for _, item := range []agenda.TaskClaim{first, second, other} {
		if err := store.SaveClaim(ctx, item); err != nil {
			t.Fatal(err)
		}
	}

	claims, err := store.ListClaims(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(claims) != 2 {
		t.Fatalf("claims = %+v", claims)
	}
	if claims[0].Branch.String() != "feat/sample-b" {
		t.Fatalf("first claim = %+v", claims[0])
	}

	err = store.RemoveClaims(ctx, []agenda.TaskID{first.Task})
	if err != nil {
		t.Fatal(err)
	}

	claims, err = store.ListClaims(ctx)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claims = %+v, err = %v", claims, err)
	}
}

func TestConcurrentClaimsAreAllKept(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	tasks := []string{
		"bc-91cd.1",
		"bc-91cd.2",
		"bc-91cd.3",
		"bc-91cd.4",
		"bc-91cd.5",
		"bc-91cd.6",
		"bc-91cd.7",
		"bc-91cd.8",
	}

	var group sync.WaitGroup
	for _, task := range tasks {
		item := claim(t, task, "feat/sample-a")
		group.Go(func() {
			if err := store.SaveClaim(ctx, item); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()

	claims, err := store.ListClaims(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(claims) != len(tasks) {
		t.Fatalf("kept %d of %d claims", len(claims), len(tasks))
	}
}

func TestInjectedKeysAreKeptPerSession(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	first := session(t, "session-a")
	second := session(t, "session-b")

	err := store.RecordInjected(ctx, first, []string{"record:bc-0001"})
	if err != nil {
		t.Fatal(err)
	}

	keys, err := store.InjectedKeys(ctx, first)
	if err != nil || !keys["record:bc-0001"] {
		t.Fatalf("keys = %v, err = %v", keys, err)
	}

	keys, err = store.InjectedKeys(ctx, second)
	if err != nil || len(keys) != 0 {
		t.Fatalf("other session keys = %v, err = %v", keys, err)
	}
}

func TestOldSessionsArePruned(t *testing.T) {
	store, directory := newStore(t)
	ctx := context.Background()
	old := `{"version":1,"sessions":{"old":{"day":"2020-01-01",` +
		`"keys":["record:bc-0001"]}}}`

	path := filepath.Join(directory, "injected.json")
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	err := store.RecordInjected(ctx, session(t, "new"), []string{"x"})
	if err != nil {
		t.Fatal(err)
	}

	var document struct {
		Sessions map[string]any `json:"sessions"`
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}

	if _, kept := document.Sessions["old"]; kept {
		t.Fatalf("old session kept: %s", data)
	}
	if _, kept := document.Sessions["new"]; !kept {
		t.Fatalf("new session missing: %s", data)
	}
}

func TestCorruptStateIsReported(t *testing.T) {
	store, directory := newStore(t)
	path := filepath.Join(directory, "claims.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ListClaims(context.Background()); err == nil {
		t.Fatal("expected an error for a corrupt file")
	}
}

func TestOpenUsesTheCommonGitFolder(t *testing.T) {
	root := t.TempDir()
	command := exec.Command("git", "init", "-q")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, output)
	}

	store, err := states.Open(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}

	err = store.SaveClaim(context.Background(), claim(t, "bc-91cd.1", "main"))
	if err != nil {
		t.Fatal(err)
	}

	claims := filepath.Join(root, ".git", "beacon", "claims.json")
	if _, err := os.Stat(claims); err != nil {
		t.Fatalf("claims file not in .git/beacon: %v", err)
	}
}

type metricsFixture struct {
	Version  int `json:"version"`
	Sessions map[string]struct {
		Day    string `json:"day"`
		Events map[string]struct {
			Calls    int `json:"calls"`
			Injected int `json:"injected"`
			Lines    int `json:"lines"`
		} `json:"events"`
	} `json:"sessions"`
}

func readMetrics(t *testing.T, directory string) metricsFixture {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(directory, "metrics.json"))
	if err != nil {
		t.Fatal(err)
	}

	var metrics metricsFixture
	if err := json.Unmarshal(data, &metrics); err != nil {
		t.Fatal(err)
	}

	return metrics
}

func TestInjectionsAreCountedPerSessionAndEvent(t *testing.T) {
	store, directory := newStore(t)
	ctx := context.Background()
	first := session(t, "session-a")

	for _, lines := range []int{4, 0, 6} {
		if err := store.RecordInjection(ctx, first, "edit", lines); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.RecordInjection(ctx, first, "start", 9); err != nil {
		t.Fatal(err)
	}
	other := session(t, "b")
	if err := store.RecordInjection(ctx, other, "edit", 1); err != nil {
		t.Fatal(err)
	}

	metrics := readMetrics(t, directory)
	edit := metrics.Sessions["session-a"].Events["edit"]

	if edit.Calls != 3 || edit.Injected != 2 || edit.Lines != 10 {
		t.Errorf("edit = %+v", edit)
	}
	if metrics.Sessions["session-a"].Events["start"].Lines != 9 {
		t.Errorf("start = %+v", metrics.Sessions["session-a"].Events["start"])
	}
	if metrics.Sessions["b"].Events["edit"].Calls != 1 || metrics.Version != 1 {
		t.Errorf("metrics = %+v", metrics)
	}
}

func TestConcurrentInjectionsLoseNoCount(t *testing.T) {
	store, directory := newStore(t)
	ctx := context.Background()
	id := session(t, "session-a")

	var group sync.WaitGroup
	for range 20 {
		group.Add(1)
		go func() {
			defer group.Done()

			if err := store.RecordInjection(ctx, id, "edit", 2); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()

	edit := readMetrics(t, directory).Sessions["session-a"].Events["edit"]
	if edit.Calls != 20 || edit.Lines != 40 {
		t.Errorf("edit = %+v", edit)
	}
}

func TestOldMetricsSessionsArePruned(t *testing.T) {
	store, directory := newStore(t)
	old := `{"version":1,"sessions":{"old":{"day":"2020-01-01",` +
		`"events":{"edit":{"calls":1,"injected":1,"lines":3}}}}}`

	path := filepath.Join(directory, "metrics.json")
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	err := store.RecordInjection(
		context.Background(), session(t, "new"), "start", 1,
	)
	if err != nil {
		t.Fatal(err)
	}

	metrics := readMetrics(t, directory)
	if _, kept := metrics.Sessions["old"]; kept || len(metrics.Sessions) != 1 {
		t.Errorf("sessions = %+v", metrics.Sessions)
	}
}

func moment(t *testing.T, text string) time.Time {
	t.Helper()

	when, err := time.Parse(time.RFC3339, text)
	if err != nil {
		t.Fatal(err)
	}

	return when
}

func TestTheActiveAgentIsThePersonsLastHookInThatWorktree(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	now := moment(t, "2026-11-26T10:00:00Z")

	err := store.NoteAgent(ctx, "/work/a", "claude-code", "s1", now)
	if err != nil {
		t.Fatal(err)
	}

	agent, err := store.ActiveAgent(ctx, "/work/a", now.Add(time.Hour))
	if err != nil || agent != "claude-code" {
		t.Errorf("agent %q, err %v", agent, err)
	}

	agent, err = store.ActiveAgent(ctx, "/work/b", now)
	if err != nil || agent != "" {
		t.Errorf("other worktree: agent %q, err %v", agent, err)
	}
}

func TestTheActiveAgentExpiresAfterADay(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()
	now := moment(t, "2026-11-26T10:00:00Z")

	err := store.NoteAgent(ctx, "/work/a", "claude-code", "s1", now)
	if err != nil {
		t.Fatal(err)
	}

	agent, err := store.ActiveAgent(ctx, "/work/a", now.Add(25*time.Hour))
	if err != nil || agent != "" {
		t.Errorf("agent %q, err %v", agent, err)
	}
}

func TestNotingTheSameSessionTwiceDoesNotRewriteTheFile(t *testing.T) {
	store, directory := newStore(t)
	ctx := context.Background()
	now := moment(t, "2026-11-26T10:00:00Z")
	path := filepath.Join(directory, "activity.json")

	err := store.NoteAgent(ctx, "/w", "claude-code", "s1", now)
	if err != nil {
		t.Fatal(err)
	}

	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	later := now.Add(time.Minute)
	err = store.NoteAgent(ctx, "/w", "claude-code", "s1", later)
	if err != nil {
		t.Fatal(err)
	}

	second, err := os.ReadFile(path)
	if err != nil || string(first) != string(second) {
		t.Errorf("file was rewritten: %s vs %s", first, second)
	}

	err = store.NoteAgent(ctx, "/w", "claude-code", "s2", later)
	if err != nil {
		t.Fatal(err)
	}

	third, err := os.ReadFile(path)
	if err != nil || string(first) == string(third) {
		t.Errorf("a new session should be written: %s", third)
	}
}

func TestOldWorktreeActivityIsPruned(t *testing.T) {
	store, directory := newStore(t)
	ctx := context.Background()
	old := `{"version":1,"worktrees":{"/gone":{"agent":"claude-code",` +
		`"session":"s0","at":"2020-01-01T00:00:00Z"}}}`

	path := filepath.Join(directory, "activity.json")
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	now := moment(t, "2026-11-26T10:00:00Z")
	if err := store.NoteAgent(ctx, "/w", "claude-code", "s1", now); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), "/gone") {
		t.Errorf("old entry survived: %s", data)
	}
}

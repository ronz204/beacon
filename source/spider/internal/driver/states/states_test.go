package states_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

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

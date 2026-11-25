package commit_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
)

const goalPath = ".beacon/records/goals/bc-0001-sample.yaml"

type sandbox struct {
	t    *testing.T
	root string
}

func newSandbox(t *testing.T) sandbox {
	t.Helper()

	box := sandbox{t: t, root: t.TempDir()}
	box.git("", "init", "-q")
	box.git("", "config", "user.name", "jane")
	box.git("", "config", "user.email", "jane@example.com")
	box.git("", "config", "commit.gpgsign", "false")

	return box
}

func (s sandbox) git(day string, args ...string) string {
	s.t.Helper()

	command := exec.Command("git", args...)
	command.Dir = s.root
	if day != "" {
		stamp := day + "T12:00:00"
		command.Env = append(
			os.Environ(),
			"GIT_AUTHOR_DATE="+stamp,
			"GIT_COMMITTER_DATE="+stamp,
		)
	}

	output, err := command.CombinedOutput()
	if err != nil {
		s.t.Fatalf("git %v: %v\n%s", args, err, output)
	}

	return strings.TrimSpace(string(output))
}

func (s sandbox) write(relative, text string) {
	s.t.Helper()

	path := filepath.Join(s.root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		s.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		s.t.Fatal(err)
	}
}

func (s sandbox) commit(day string) string {
	s.t.Helper()

	s.git(day, "add", "-A")
	s.git(day, "commit", "-q", "-m", "sample change")

	return s.git("", "rev-parse", "HEAD")
}

func requireHistoryFailure(t *testing.T, err error) {
	t.Helper()

	var got kernel.KernelError
	if !errors.As(err, &got) || got.Code != kernel.KernelHistoryFailed {
		t.Fatalf("expected history failure, got %v", err)
	}
}

func TestFilesChangedSinceIncludesCommittedAndLocalWork(t *testing.T) {
	box := newSandbox(t)
	box.write("src/a.go", "first")
	box.write(goalPath, "first")
	base := box.commit("2026-01-10")

	box.write("src/a.go", "second")
	box.commit("2026-02-10")
	box.write(goalPath, "edited")
	box.write("src/new.go", "untracked")
	box.write("src/café.go", "untracked")

	repository := commit.NewRepository(box.root)
	files, err := repository.FilesChangedSince(context.Background(), base)
	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(files)
	want := []string{goalPath, "src/a.go", "src/café.go", "src/new.go"}
	if !slices.Equal(files, want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
}

func TestFilesChangedSinceRejectsAnythingButAHash(t *testing.T) {
	repository := commit.NewRepository(newSandbox(t).root)

	for _, value := range []string{"--force", "main", ""} {
		_, err := repository.FilesChangedSince(context.Background(), value)

		requireHistoryFailure(t, err)
	}
}

func TestFilesChangedSinceReportsUnknownCommit(t *testing.T) {
	box := newSandbox(t)
	box.write("src/a.go", "first")
	box.commit("2026-01-10")

	repository := commit.NewRepository(box.root)
	_, err := repository.FilesChangedSince(context.Background(), "abcdef12")

	requireHistoryFailure(t, err)
}

func TestCommitsSinceListsDaysAndFiles(t *testing.T) {
	box := newSandbox(t)
	box.write("src/a.go", "first")
	box.commit("2026-01-10")
	box.write("src/a.go", "second")
	box.write("src/b.go", "second")
	box.commit("2026-02-10")

	since, err := kernel.ParseCalendarDate("2026-02-01")
	if err != nil {
		t.Fatal(err)
	}

	repository := commit.NewRepository(box.root)
	commits, err := repository.CommitsSince(context.Background(), since)
	if err != nil {
		t.Fatal(err)
	}

	if len(commits) != 1 {
		t.Fatalf("commits = %+v", commits)
	}
	if commits[0].Day.String() != "2026-02-10" {
		t.Fatalf("day = %q", commits[0].Day)
	}
	if !slices.Equal(commits[0].Files, []string{"src/a.go", "src/b.go"}) {
		t.Fatalf("files = %v", commits[0].Files)
	}
}

func TestGuideChangeDaysKeepsTheLatestCommit(t *testing.T) {
	box := newSandbox(t)
	guide := ".beacon/records/guides/bc-0a9b-sample.md"
	box.write(guide, "first")
	box.commit("2026-03-05")
	box.write(guide, "second")
	box.commit("2026-04-01")
	box.write("src/a.go", "unrelated")
	box.commit("2026-05-01")

	repository := commit.NewRepository(box.root)
	days, err := repository.GuideChangeDays(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(days) != 1 {
		t.Fatalf("days = %v", days)
	}
	for id, day := range days {
		if id.String() != "bc-0a9b" || day.String() != "2026-04-01" {
			t.Fatalf("days = %v", days)
		}
	}
}

func TestStateChangesReportsRecordsAndBaselines(t *testing.T) {
	box := newSandbox(t)
	removed := ".beacon/records/goals/bc-0002-sample.yaml"
	box.write(goalPath, "before")
	box.write(removed, "gone")
	box.write("src/a.go", "code")
	base := box.commit("2026-01-10")

	box.write(goalPath, "after")
	box.write("src/a.go", "changed code")
	box.write("tests/login-snapshots/a.png", "image")
	if err := os.Remove(filepath.Join(box.root, removed)); err != nil {
		t.Fatal(err)
	}

	baseline, err := kernel.ParsePathGlob("tests/**/*-snapshots/**")
	if err != nil {
		t.Fatal(err)
	}

	repository := commit.NewRepository(box.root)
	changes, err := repository.StateChanges(
		context.Background(),
		base,
		[]kernel.PathGlob{baseline},
		match.NewMatcher(),
	)
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]commit.StateChange)
	for _, change := range changes {
		found[change.Path] = change
	}

	if len(found) != 3 {
		t.Fatalf("changes = %+v", changes)
	}

	modified := found[goalPath]
	if modified.Kind != commit.ChangeModified {
		t.Fatalf("goal = %+v", modified)
	}
	before, after := string(modified.Before), string(modified.After)
	if before != "before" || after != "after" {
		t.Fatalf("goal contents = %q -> %q", before, after)
	}

	deleted := found[removed]
	isDeleted := deleted.Kind == commit.ChangeDeleted
	if !isDeleted || string(deleted.Before) != "gone" {
		t.Fatalf("removed = %+v", deleted)
	}

	image := found["tests/login-snapshots/a.png"]
	if image.Kind != commit.ChangeAdded || !image.Baseline {
		t.Fatalf("baseline = %+v", image)
	}
}

func TestStateChangesRejectsUnsafeRevision(t *testing.T) {
	repository := commit.NewRepository(newSandbox(t).root)

	for _, value := range []string{"--output=x", "HEAD:secret", "a b", ""} {
		_, err := repository.StateChanges(
			context.Background(),
			value,
			nil,
			match.NewMatcher(),
		)

		requireHistoryFailure(t, err)
	}
}

func TestListBranchesAndWorktrees(t *testing.T) {
	box := newSandbox(t)
	box.write("src/a.go", "first")
	box.commit("2026-01-10")
	box.git("", "branch", "feat/sample-a")

	repository := commit.NewRepository(box.root)
	branches, err := repository.ListBranches(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	names := make([]string, 0, len(branches))
	for _, branch := range branches {
		names = append(names, branch.String())
	}
	if !slices.Contains(names, "feat/sample-a") {
		t.Fatalf("branches = %v", names)
	}

	ctx := context.Background()
	exists, err := repository.WorktreeExists(ctx, box.root)
	if err != nil || !exists {
		t.Fatalf("existing worktree = %v, err = %v", exists, err)
	}

	gone := filepath.Join(box.root, "gone")
	exists, err = repository.WorktreeExists(ctx, gone)
	if err != nil || exists {
		t.Fatalf("missing worktree = %v, err = %v", exists, err)
	}
}

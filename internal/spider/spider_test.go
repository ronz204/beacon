package spider_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/internal/spider"
)

func makeFolders(t *testing.T, paths ...string) {
	t.Helper()

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func requireMissing(t *testing.T, err error) {
	t.Helper()

	var got spider.ProjectError
	if !errors.As(err, &got) || got.Code != spider.ProjectMissing {
		t.Fatalf("expected project_missing, got %v", err)
	}
}

func TestLocateFindsProjectFromNestedFolder(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "src", "auth")
	makeFolders(t, filepath.Join(root, ".beacon"), nested)

	if _, err := spider.Locate(nested); err != nil {
		t.Fatal(err)
	}
}

func TestLocateStopsAtTheRepositoryRoot(t *testing.T) {
	outer := t.TempDir()
	repository := filepath.Join(outer, "repository")
	nested := filepath.Join(repository, "src")
	makeFolders(
		t,
		filepath.Join(outer, ".beacon"),
		filepath.Join(repository, ".git"),
		nested,
	)

	_, err := spider.Locate(nested)

	requireMissing(t, err)
}

func TestLocateAcceptsWorktreeGitFile(t *testing.T) {
	outer := t.TempDir()
	repository := filepath.Join(outer, "worktree")
	makeFolders(t, filepath.Join(outer, ".beacon"), repository)

	gitFile := filepath.Join(repository, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := spider.Locate(repository)

	requireMissing(t, err)
}

func TestLocateFindsProjectAtTheRepositoryRoot(t *testing.T) {
	repository := t.TempDir()
	makeFolders(
		t,
		filepath.Join(repository, ".beacon"),
		filepath.Join(repository, ".git"),
	)

	if _, err := spider.Locate(repository); err != nil {
		t.Fatal(err)
	}
}

package files_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/driver/files"
	"github.com/ronz204/beacon/source/system"
)

func newRepo(t *testing.T) (files.Repo, string) {
	t.Helper()

	directory := t.TempDir()
	_, err := system.Git(context.Background(), directory, "init", "-q")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}

	repo, err := files.Open(context.Background(), directory)
	if err != nil {
		t.Fatal(err)
	}

	return repo, directory
}

func TestWriteThenReadRoundTrips(t *testing.T) {
	ctx := context.Background()
	repo, _ := newRepo(t)

	if err := repo.Write(ctx, ".claude/settings.json", "{}\r\n"); err != nil {
		t.Fatal(err)
	}

	text, exists, err := repo.Read(ctx, ".claude/settings.json")
	if err != nil || !exists || text != "{}\r\n" {
		t.Errorf("text %q, exists %v, err %v", text, exists, err)
	}
}

func TestReadOfAMissingFileIsNotAnError(t *testing.T) {
	repo, _ := newRepo(t)

	text, exists, err := repo.Read(context.Background(), "AGENTS.md")
	if err != nil || exists || text != "" {
		t.Errorf("text %q, exists %v, err %v", text, exists, err)
	}
}

func TestBackupLandsUnderTheGitFolder(t *testing.T) {
	ctx := context.Background()
	repo, directory := newRepo(t)

	if err := repo.Backup(ctx, "stamp", "CLAUDE.md", "old"); err != nil {
		t.Fatal(err)
	}

	saved := filepath.Join(
		directory, ".git", "beacon", "backups", "stamp", "CLAUDE.md",
	)
	data, err := os.ReadFile(saved)
	if err != nil || string(data) != "old" {
		t.Errorf("backup %q, err %v", data, err)
	}
}

func TestPathsOutsideTheRepoAreRejected(t *testing.T) {
	ctx := context.Background()
	repo, _ := newRepo(t)

	for _, path := range []string{"../escape.md", "/abs.md", ".."} {
		err := repo.Write(ctx, path, "x")

		var failure action.InstallError
		if !errors.As(err, &failure) || failure.Code != action.FilePathInvalid {
			t.Errorf("path %q: err = %v", path, err)
		}
	}
}

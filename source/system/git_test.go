package system_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/system"
)

func TestGitRunsInTheGivenFolder(t *testing.T) {
	directory := t.TempDir()

	_, err := system.Git(context.Background(), directory, "init", "-q")
	if err != nil {
		t.Fatal(err)
	}

	output, err := system.Git(
		context.Background(),
		directory,
		"rev-parse",
		"--is-inside-work-tree",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(output) != "true" {
		t.Fatalf("output = %q", output)
	}
}

func TestGitReportsTheFailingCommand(t *testing.T) {
	_, err := system.Git(context.Background(), t.TempDir(), "log")

	if err == nil || !strings.Contains(err.Error(), "git log") {
		t.Fatalf("err = %v", err)
	}
}

func TestGitCommonDirIsAbsoluteInsideARepository(t *testing.T) {
	root := t.TempDir()
	if _, err := system.Git(context.Background(), root, "init", "-q"); err != nil {
		t.Fatal(err)
	}

	common, err := system.GitCommonDir(context.Background(), root)

	if err != nil {
		t.Fatal(err)
	}

	if !filepath.IsAbs(common) || filepath.Base(common) != ".git" {
		t.Fatalf("common = %q", common)
	}
}

func TestGitCommonDirFailsOutsideARepository(t *testing.T) {
	if _, err := system.GitCommonDir(context.Background(), t.TempDir()); err == nil {
		t.Fatal("expected an error")
	}
}

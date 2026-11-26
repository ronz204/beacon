package anchor_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/source/anchor"
	"github.com/ronz204/beacon/source/system"
)

func newProject(t *testing.T) string {
	t.Helper()

	directory := t.TempDir()
	_, err := system.Git(context.Background(), directory, "init", "-q")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}

	return directory
}

func TestDryRunReportsTheGuideAndWritesNothing(t *testing.T) {
	ctx := context.Background()
	directory := newProject(t)

	tool, err := anchor.New(ctx, directory)
	if err != nil {
		t.Fatal(err)
	}

	view, err := tool.Install(ctx, anchor.InstallOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}

	if len(view.Changes) != 1 || view.Changes[0].Action != "create" {
		t.Fatalf("changes = %+v", view.Changes)
	}
	if _, err := os.Stat(filepath.Join(directory, "AGENTS.md")); err == nil {
		t.Error("dry run created AGENTS.md")
	}
}

func TestInstallTwiceChangesNothingTheSecondTime(t *testing.T) {
	ctx := context.Background()
	directory := newProject(t)

	tool, err := anchor.New(ctx, directory)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tool.Install(ctx, anchor.InstallOptions{}); err != nil {
		t.Fatal(err)
	}

	view, err := tool.Install(ctx, anchor.InstallOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(view.Changes) != 0 {
		t.Errorf("second install changed %+v", view.Changes)
	}
}

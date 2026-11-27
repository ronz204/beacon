package anchor_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func newAnchor(t *testing.T, directory string) anchor.Anchor {
	t.Helper()

	tool, err := anchor.New(context.Background(), directory)
	if err != nil {
		t.Fatal(err)
	}

	return tool
}

func exists(directory, name string) bool {
	_, err := os.Stat(filepath.Join(directory, name))

	return !errors.Is(err, os.ErrNotExist)
}

func TestPreviewReportsTheFilesAndWritesNothing(t *testing.T) {
	directory := newProject(t)
	options := anchor.InstallOptions{Preview: true, Hooks: true, Reinject: true}

	view, err := newAnchor(t, directory).Install(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}

	if len(view.Changes) != 4 || view.Changes[0].Action != "create" {
		t.Fatalf("changes = %+v", view.Changes)
	}
	if exists(directory, "AGENTS.md") || exists(directory, ".gitattributes") {
		t.Error("preview created files")
	}
}

func TestInstallTwiceChangesNothingTheSecondTime(t *testing.T) {
	ctx := context.Background()
	tool := newAnchor(t, newProject(t))
	options := anchor.InstallOptions{Hooks: true, Reinject: true}

	if _, err := tool.Install(ctx, options); err != nil {
		t.Fatal(err)
	}

	view, err := tool.Install(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	if len(view.Changes) != 0 {
		t.Errorf("second install changed %+v", view.Changes)
	}
}

func TestUninstallRestoresExistingFilesAndDeletesCreatedOnes(t *testing.T) {
	ctx := context.Background()
	directory := newProject(t)
	notes := filepath.Join(directory, "AGENTS.md")
	options := anchor.InstallOptions{Hooks: true, Reinject: true}

	if err := os.WriteFile(notes, []byte("# Mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := newAnchor(t, directory)
	if _, err := tool.Install(ctx, options); err != nil {
		t.Fatal(err)
	}

	view, err := tool.Uninstall(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	restored, err := os.ReadFile(notes)
	if err != nil || string(restored) != "# Mine\n" {
		t.Errorf("AGENTS.md = %q, err %v", restored, err)
	}
	for _, created := range []string{
		".gitattributes", "CLAUDE.md", ".claude/settings.json",
	} {
		if exists(directory, created) {
			t.Errorf("%s should have been deleted", created)
		}
	}
	if !view.Uninstall || len(view.Changes) != 4 {
		t.Errorf("view = %+v", view)
	}
}

func TestUninstallKeepsABlockThePersonEdited(t *testing.T) {
	ctx := context.Background()
	directory := newProject(t)
	notes := filepath.Join(directory, "AGENTS.md")
	options := anchor.InstallOptions{Hooks: true, Reinject: true}
	tool := newAnchor(t, directory)

	if _, err := tool.Install(ctx, options); err != nil {
		t.Fatal(err)
	}

	installed, err := os.ReadFile(notes)
	if err != nil {
		t.Fatal(err)
	}

	edited := []byte(strings.Replace(
		string(installed), "Follow it.", "Follow it, mostly.", 1,
	))
	if err := os.WriteFile(notes, edited, 0o644); err != nil {
		t.Fatal(err)
	}

	view, err := tool.Uninstall(ctx, options)
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(notes)
	if err != nil || string(after) != string(edited) {
		t.Errorf("edited file changed: %q, err %v", after, err)
	}
	if len(view.Kept) != 1 || view.Kept[0].Path != "AGENTS.md" {
		t.Errorf("kept = %+v", view.Kept)
	}
}

func TestInstallBacksUpAnExistingFile(t *testing.T) {
	ctx := context.Background()
	directory := newProject(t)
	options := anchor.InstallOptions{Hooks: true, Reinject: true}

	attributes := filepath.Join(directory, ".gitattributes")
	err := os.WriteFile(attributes, []byte("*.go text\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := newAnchor(t, directory).Install(ctx, options); err != nil {
		t.Fatal(err)
	}

	matches, err := filepath.Glob(filepath.Join(
		directory, ".git", "beacon", "backups", "*", ".gitattributes",
	))
	if err != nil || len(matches) != 1 {
		t.Errorf("backups = %v, err %v", matches, err)
	}
}

func TestInstallReportsWhatTheOptionsSkip(t *testing.T) {
	ctx := context.Background()
	tool := newAnchor(t, newProject(t))

	cases := []struct {
		options anchor.InstallOptions
		want    int
	}{
		{anchor.InstallOptions{Preview: true, Precommit: true}, 2},
		{anchor.InstallOptions{Preview: true, Hooks: true, Precommit: true}, 1},
		{
			anchor.InstallOptions{
				Preview:   true,
				Hooks:     true,
				Reinject:  true,
				Precommit: true,
			},
			0,
		},
		{anchor.InstallOptions{Preview: true, Hooks: true, Reinject: true}, 1},
	}

	for _, c := range cases {
		view, err := tool.Install(ctx, c.options)
		if err != nil {
			t.Fatal(err)
		}

		if len(view.Skipped) != c.want {
			t.Errorf("%+v: skipped = %v", c.options, view.Skipped)
		}
	}
}

func TestInstallSaysWhetherTheProjectIsInitialized(t *testing.T) {
	ctx := context.Background()
	directory := newProject(t)
	options := anchor.InstallOptions{Preview: true, Hooks: true, Reinject: true}
	tool := newAnchor(t, directory)

	view, err := tool.Install(ctx, options)
	if err != nil || view.Initialized {
		t.Fatalf("initialized = %v, err %v", view.Initialized, err)
	}

	config := filepath.Join(directory, ".beacon", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(config), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte("schema: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	view, err = tool.Install(ctx, options)
	if err != nil || !view.Initialized {
		t.Errorf("initialized = %v, err %v", view.Initialized, err)
	}
}

package e2e_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

type installData struct {
	DryRun      bool `json:"dryRun"`
	Uninstall   bool `json:"uninstall"`
	Initialized bool `json:"initialized"`
	Changes     []struct {
		Path   string   `json:"path"`
		Action string   `json:"action"`
		Parts  []string `json:"parts"`
	} `json:"changes"`
	Skipped []string `json:"skipped"`
	Kept    []struct {
		Path string `json:"path"`
	} `json:"kept"`
}

func newBareRepository(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	utils.Git(t, root, "init", "-q", "-b", "main")
	utils.Git(t, root, "config", "user.name", "jane")

	return root
}

func present(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))

	return !errors.Is(err, os.ErrNotExist)
}

func runInstall(t *testing.T, root string, args ...string) installData {
	t.Helper()

	arguments := append([]string{"install", "--json"}, args...)
	result := utils.Beacon(t, root, arguments...)
	if result.Code != 0 {
		t.Fatalf("exit %d: %s%s", result.Code, result.Stdout, result.Stderr)
	}

	var data installData
	decodeData(t, result.Stdout, &data)

	return data
}

func TestInstallDryRunListsTheFilesAndWritesNothing(t *testing.T) {
	root := newBareRepository(t)

	data := runInstall(t, root, "--dry-run")

	if !data.DryRun || len(data.Changes) != 4 {
		t.Fatalf("data = %+v", data)
	}
	if present(root, "AGENTS.md") || present(root, ".gitattributes") {
		t.Error("dry run wrote files")
	}
}

func TestInstallCreatesTheGuideAndTheLineEndingRule(t *testing.T) {
	root := newBareRepository(t)

	runInstall(t, root)

	guide := utils.ReadFile(t, filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(guide, "beacon brief") {
		t.Errorf("AGENTS.md = %q", guide)
	}

	attributes := utils.ReadFile(t, filepath.Join(root, ".gitattributes"))
	if !strings.Contains(attributes, ".beacon/** text eol=lf") {
		t.Errorf(".gitattributes = %q", attributes)
	}
}

func TestInstallTwiceChangesNothing(t *testing.T) {
	root := newBareRepository(t)

	runInstall(t, root)
	data := runInstall(t, root)

	if len(data.Changes) != 0 {
		t.Errorf("second install changed %+v", data.Changes)
	}
}

func TestUninstallLeavesTheRepositoryAsItWas(t *testing.T) {
	root := newBareRepository(t)
	utils.WriteFile(t, root, "AGENTS.md", "# Mine\n")

	runInstall(t, root)
	data := runInstall(t, root, "--uninstall")

	if !data.Uninstall || len(data.Changes) != 4 {
		t.Fatalf("data = %+v", data)
	}
	if utils.ReadFile(t, filepath.Join(root, "AGENTS.md")) != "# Mine\n" {
		t.Error("AGENTS.md was not restored")
	}
	if present(root, ".gitattributes") {
		t.Error(".gitattributes should have been deleted")
	}
}

func TestInstallKeepsACopyOfWhatItChanges(t *testing.T) {
	root := newBareRepository(t)
	utils.WriteFile(t, root, "AGENTS.md", "# Mine\n")

	runInstall(t, root)

	matches, err := filepath.Glob(filepath.Join(
		root, ".git", "beacon", "backups", "*", "AGENTS.md",
	))
	if err != nil || len(matches) != 1 {
		t.Fatalf("backups = %v, err %v", matches, err)
	}
	if utils.ReadFile(t, matches[0]) != "# Mine\n" {
		t.Error("the backup is not the original file")
	}
}

func TestInstallSkipHooksReportsWhatItLeftOut(t *testing.T) {
	root := newBareRepository(t)

	data := runInstall(t, root, "--dry-run", "--skip", "hooks")

	if len(data.Skipped) != 2 {
		t.Errorf("skipped = %v", data.Skipped)
	}
}

func TestInstallRejectsAnUnknownSkipPart(t *testing.T) {
	root := newBareRepository(t)

	result := utils.Beacon(t, root, "install", "--json", "--skip", "x")

	if result.Code != 1 || !strings.Contains(result.Stdout, "skip_invalid") {
		t.Errorf("exit %d: %s", result.Code, result.Stdout)
	}
}

func TestInstallPointsToTheNextStepWithoutBeacon(t *testing.T) {
	root := newBareRepository(t)

	result := utils.Beacon(t, root, "install", "--json", "--dry-run")
	envelope := parseRead(t, result.Stdout)

	if !hintContains(envelope.Hint, ".beacon/config.yaml") {
		t.Errorf("hint = %v", envelope.Hint)
	}

	initialized := utils.NewRepository(t)
	result = utils.Beacon(t, initialized, "install", "--json", "--dry-run")

	if parseRead(t, result.Stdout).Hint != nil {
		t.Error("an initialized project should not get the hint")
	}
}

func TestInstallOutsideAGitRepositoryFails(t *testing.T) {
	result := utils.Beacon(t, t.TempDir(), "install", "--json")

	if result.Code != 1 || !strings.Contains(result.Stdout, "history_failed") {
		t.Errorf("exit %d: %s", result.Code, result.Stdout)
	}
}

func TestInstallDryRunTextNamesTheFilesAndTheirLines(t *testing.T) {
	root := newBareRepository(t)

	result := utils.Beacon(t, root, "install", "--dry-run")

	for _, want := range []string{
		"dry run: nothing was written",
		"would create AGENTS.md (guide)",
		"would create .gitattributes (attributes)",
		"  + .beacon/** text eol=lf",
		"hint: create .beacon/config.yaml with schema: 1",
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Errorf("missing %q in:\n%s", want, result.Stdout)
		}
	}
}

func TestInstallDoesNotDuplicateTheLineEndingRuleWrittenByHand(t *testing.T) {
	root := newBareRepository(t)
	utils.WriteFile(t, root, ".gitattributes", ".beacon/** text eol=lf\n")

	data := runInstall(t, root)

	for _, change := range data.Changes {
		if change.Path == ".gitattributes" {
			t.Errorf(".gitattributes should be left alone: %+v", change)
		}
	}
}

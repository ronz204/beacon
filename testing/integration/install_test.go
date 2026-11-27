package integration_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

type installData struct {
	Preview     bool `json:"preview"`
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

func runInstall(t *testing.T, root string, args ...string) installData {
	t.Helper()

	arguments := append([]string{"install", "--json"}, args...)
	result := utilities.Beacon(t, root, arguments...)
	if result.Code != 0 {
		t.Fatalf("exit %d: %s%s", result.Code, result.Stdout, result.Stderr)
	}

	var data installData
	decodeData(t, result.Stdout, &data)

	return data
}

func TestInstallPreviewListsTheFilesAndWritesNothing(t *testing.T) {
	root := newBareRepository(t)

	data := runInstall(t, root, "--preview")

	if !data.Preview || len(data.Changes) != 6 {
		t.Fatalf("data = %+v", data)
	}
	if present(root, "AGENTS.md") || present(root, ".gitattributes") {
		t.Error("preview wrote files")
	}
}

func TestInstallCreatesTheGuideAndTheLineEndingRule(t *testing.T) {
	root := newBareRepository(t)

	runInstall(t, root)

	guide := utilities.ReadFile(t, filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(guide, "beacon brief") {
		t.Errorf("AGENTS.md = %q", guide)
	}

	attributes := utilities.ReadFile(t, filepath.Join(root, ".gitattributes"))
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
	utilities.WriteFile(t, root, "AGENTS.md", "# Mine\n")

	runInstall(t, root)
	data := runInstall(t, root, "--uninstall")

	if !data.Uninstall || len(data.Changes) != 6 {
		t.Fatalf("data = %+v", data)
	}
	if utilities.ReadFile(t, filepath.Join(root, "AGENTS.md")) != "# Mine\n" {
		t.Error("AGENTS.md was not restored")
	}
	if present(root, ".gitattributes") {
		t.Error(".gitattributes should have been deleted")
	}
}

func TestInstallKeepsACopyOfWhatItChanges(t *testing.T) {
	root := newBareRepository(t)
	utilities.WriteFile(t, root, "AGENTS.md", "# Mine\n")

	runInstall(t, root)

	matches, err := filepath.Glob(filepath.Join(
		root, ".git", "beacon", "backups", "*", "AGENTS.md",
	))
	if err != nil || len(matches) != 1 {
		t.Fatalf("backups = %v, err %v", matches, err)
	}
	if utilities.ReadFile(t, matches[0]) != "# Mine\n" {
		t.Error("the backup is not the original file")
	}
}

func TestInstallSkipHooksReportsWhatItLeftOut(t *testing.T) {
	root := newBareRepository(t)

	data := runInstall(t, root, "--preview", "--skip", "hooks")

	if len(data.Skipped) != 2 {
		t.Errorf("skipped = %v", data.Skipped)
	}
}

func TestInstallRejectsAnUnknownSkipPiece(t *testing.T) {
	root := newBareRepository(t)

	result := utilities.Beacon(t, root, "install", "--json", "--skip", "x")

	if result.Code != 1 || !strings.Contains(result.Stdout, "piece_unknown") {
		t.Errorf("exit %d: %s", result.Code, result.Stdout)
	}
}

func TestInstallPointsToTheNextStepWithoutBeacon(t *testing.T) {
	root := newBareRepository(t)

	result := utilities.Beacon(t, root, "install", "--json", "--preview")
	envelope := parseEnvelope(t, result.Stdout)

	if !hintContains(envelope.Hint, ".beacon/config.yaml") {
		t.Errorf("hint = %v", envelope.Hint)
	}

	initialized := utilities.NewRepository(t)
	result = utilities.Beacon(t, initialized, "install", "--json", "--preview")

	if parseEnvelope(t, result.Stdout).Hint != nil {
		t.Error("an initialized project should not get the hint")
	}
}

func TestInstallOutsideAGitRepositoryFails(t *testing.T) {
	result := utilities.Beacon(t, t.TempDir(), "install", "--json")

	if result.Code != 1 || !strings.Contains(result.Stdout, "history_failed") {
		t.Errorf("exit %d: %s", result.Code, result.Stdout)
	}
}

func TestInstallPreviewTextNamesTheFilesAndTheirLines(t *testing.T) {
	root := newBareRepository(t)

	result := utilities.Beacon(t, root, "install", "--preview")

	for _, want := range []string{
		"preview: nothing was written",
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
	utilities.WriteFile(
		t,
		root,
		".gitattributes",
		".beacon/** text eol=lf linguist-generated\n"+
			".claude/rules/beacon-*.md text eol=lf\n"+
			".claude/skills/beacon-*/SKILL.md text eol=lf\n",
	)

	data := runInstall(t, root)

	for _, change := range data.Changes {
		if change.Path == ".gitattributes" {
			t.Errorf(".gitattributes should be left alone: %+v", change)
		}
	}
}

func TestInstallOverAnOldGuideOnlyReplacesTheGuideBlock(t *testing.T) {
	root := newBareRepository(t)
	runInstall(t, root)
	path := filepath.Join(root, "AGENTS.md")
	current := utilities.ReadFile(t, path)
	utilities.WriteFile(t, root, "AGENTS.md", strings.Replace(
		current,
		"beacon attest",
		"beacon old",
		1,
	))

	data := runInstall(t, root)

	if len(data.Changes) != 1 || data.Changes[0].Path != "AGENTS.md" {
		t.Fatalf("changes = %+v", data.Changes)
	}
	if utilities.ReadFile(t, path) != current {
		t.Error("the guide block was not restored")
	}
}

const recordRule = ".claude/rules/beacon-record.md"

func TestInstallWritesTheCoreRulesByDefaultAndUninstallRemovesThem(
	t *testing.T,
) {
	root := newBareRepository(t)

	runInstall(t, root)
	rule := utilities.ReadFile(t, filepath.Join(root, recordRule))

	if !strings.HasPrefix(rule, "---\nbeacon: ") {
		t.Errorf("the rule has no stamp: %q", rule)
	}
	if !strings.Contains(rule, "paths:") {
		t.Errorf("the rule lost its paths: %q", rule)
	}

	runInstall(t, root, "--uninstall")

	if present(root, recordRule) {
		t.Error("uninstall left the rule")
	}
}

func TestInstallSkipsARuleByName(t *testing.T) {
	root := newBareRepository(t)

	data := runInstall(t, root, "--skip", "record")

	if present(root, recordRule) {
		t.Error("--skip record still wrote the rule")
	}
	if len(data.Skipped) != 1 || data.Skipped[0] != "record" {
		t.Errorf("skipped = %v", data.Skipped)
	}
}

func TestInstallRejectsAnUnknownPack(t *testing.T) {
	root := newBareRepository(t)

	result := utilities.Beacon(t, root, "install", "--json", "--pack", "nope")

	if result.Code != 1 || !strings.Contains(result.Stdout, "pack_unknown") {
		t.Errorf("exit %d: %s", result.Code, result.Stdout)
	}
	if present(root, "AGENTS.md") {
		t.Error("a failed install wrote files")
	}
}

func TestInstallKeepsARuleEditedByHand(t *testing.T) {
	root := newBareRepository(t)
	runInstall(t, root)

	original := utilities.ReadFile(t, filepath.Join(root, recordRule))
	edited := original + "- mine\n"
	utilities.WriteFile(t, root, recordRule, edited)

	data := runInstall(t, root)
	again := utilities.ReadFile(t, filepath.Join(root, recordRule))

	if again != edited {
		t.Error("an edited rule was overwritten")
	}
	if len(data.Kept) != 1 || data.Kept[0].Path != recordRule {
		t.Errorf("kept = %+v", data.Kept)
	}
}

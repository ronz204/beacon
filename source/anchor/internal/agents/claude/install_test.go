package claude_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/agents/claude"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

type settings struct {
	Permissions struct {
		Allow []string `json:"allow"`
	} `json:"permissions"`
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Type    string `json:"type"`
			Command string `json:"command"`
			Timeout int    `json:"timeout"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func installed(t *testing.T, options agents.Options) settings {
	t.Helper()

	contribution, err := claude.Agent{}.Contribute(options)
	if err != nil {
		t.Fatal(err)
	}

	var entries []blocks.Entry
	for _, file := range contribution.Entries {
		entries = append(entries, file.Entries...)
	}

	text, err := blocks.MergeEntries("", entries)
	if err != nil {
		t.Fatal(err)
	}

	var parsed settings
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatal(err)
	}

	return parsed
}

func matchers(parsed settings, event string) []string {
	var found []string
	for _, group := range parsed.Hooks[event] {
		found = append(found, group.Matcher)
	}

	return found
}

func TestFullInstallCoversStartCompactAndEdit(t *testing.T) {
	parsed := installed(t, agents.Options{Hooks: true, Reinject: true})

	wantAllow := []string{"Bash(beacon *)", "PowerShell(beacon *)"}
	if !slices.Equal(parsed.Permissions.Allow, wantAllow) {
		t.Errorf("allow = %v", parsed.Permissions.Allow)
	}

	wantStart := []string{"startup|resume", "clear|compact"}
	if got := matchers(parsed, "SessionStart"); !slices.Equal(got, wantStart) {
		t.Errorf("SessionStart matchers = %v", got)
	}
	if got := matchers(parsed, "PreToolUse"); !slices.Equal(got, []string{
		"Edit|Write",
	}) {
		t.Errorf("PreToolUse matchers = %v", got)
	}
}

func TestHooksCallTheBinaryByName(t *testing.T) {
	parsed := installed(t, agents.Options{Hooks: true, Reinject: true})

	want := map[string]string{
		"SessionStart": "beacon hook start --agent claude-code",
		"PreToolUse":   "beacon hook edit --agent claude-code",
	}

	for event, command := range want {
		for _, group := range parsed.Hooks[event] {
			for _, hook := range group.Hooks {
				if hook.Type != "command" || hook.Command != command {
					t.Errorf("%s hook = %+v", event, hook)
				}
				if hook.Timeout <= 0 {
					t.Errorf("%s hook has no timeout", event)
				}
			}
		}
	}
}

func TestWithoutReinjectionOnlyTheStartHookRemains(t *testing.T) {
	parsed := installed(t, agents.Options{Hooks: true})

	got := matchers(parsed, "SessionStart")
	if !slices.Equal(got, []string{"startup|resume"}) {
		t.Errorf("SessionStart matchers = %v", got)
	}
	if len(parsed.Hooks["PreToolUse"]) != 0 {
		t.Errorf("PreToolUse = %+v", parsed.Hooks["PreToolUse"])
	}
}

func TestWithoutHooksOnlyThePermissionRemains(t *testing.T) {
	parsed := installed(t, agents.Options{})

	if len(parsed.Hooks) != 0 {
		t.Errorf("hooks = %+v", parsed.Hooks)
	}
	if len(parsed.Permissions.Allow) != 2 {
		t.Errorf("allow = %v", parsed.Permissions.Allow)
	}
}

func TestTheAgentDeclaresWhichMomentsItCovers(t *testing.T) {
	want := []agents.Moment{
		agents.MomentStart, agents.MomentCompact, agents.MomentEdit,
	}

	if got := (claude.Agent{}).Moments(); !slices.Equal(got, want) {
		t.Errorf("moments = %v", got)
	}
}

func TestTheImportBlockPointsAtTheGuide(t *testing.T) {
	contribution, err := claude.Agent{}.Contribute(agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	file := contribution.Blocks[0]
	if file.Path != "CLAUDE.md" || file.Part != agents.PartImport {
		t.Errorf("file = %+v", file)
	}

	text, err := blocks.Upsert("", file.Block)
	if err != nil || text == "" {
		t.Errorf("text %q, err %v", text, err)
	}
}

func TestPiecesBecomeStampedFilesUnderTheClaudeFolders(t *testing.T) {
	pieces := []agents.Piece{
		{
			Pack:    "core",
			Kind:    agents.PieceRule,
			Name:    "alpha",
			Content: "---\n---\n# Alpha\n",
		},
		{
			Pack:    "core",
			Kind:    agents.PieceSkill,
			Name:    "gamma",
			Content: "---\n---\n# Gamma\n",
		},
	}

	contribution, err := claude.Agent{}.Contribute(
		agents.Options{Pieces: pieces},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []agents.ManagedFile{
		{Path: ".claude/rules/beacon-alpha.md", Part: agents.PartRule},
		{Path: ".claude/skills/beacon-gamma/SKILL.md", Part: agents.PartSkill},
	}
	if len(contribution.Files) != len(want) {
		t.Fatalf("files = %+v", contribution.Files)
	}

	for index, file := range contribution.Files {
		if file.Path != want[index].Path || file.Part != want[index].Part {
			t.Errorf("file %d = %s (%s)", index, file.Path, file.Part)
		}
		if blocks.InspectManaged(file.Content) != blocks.OwnershipManaged {
			t.Errorf("%s is not stamped", file.Path)
		}
	}
}

func TestAPieceWithoutFrontmatterCannotBeInstalled(t *testing.T) {
	piece := agents.Piece{
		Pack:    "core",
		Kind:    agents.PieceRule,
		Name:    "broken",
		Content: "# No frontmatter\n",
	}

	_, err := claude.Agent{}.Contribute(
		agents.Options{Pieces: []agents.Piece{piece}},
	)
	if err == nil {
		t.Error("the piece was accepted")
	}
}

func TestWithoutPiecesNoFilesAreContributed(t *testing.T) {
	contribution, err := claude.Agent{}.Contribute(agents.Options{})

	if err != nil || len(contribution.Files) != 0 {
		t.Errorf("files = %+v, err %v", contribution.Files, err)
	}
}

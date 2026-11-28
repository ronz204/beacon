package catalog_test

import (
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v4"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/driver/catalog"
)

const (
	ruleLimit  = 40
	skillLimit = 60
)

var sixLetters = regexp.MustCompile(`^[a-z]{6}$`)

func shipped(t *testing.T) []agents.Piece {
	t.Helper()

	return shippedWith(t, catalog.DefaultBinding())
}

func shippedWith(t *testing.T, binding catalog.Binding) []agents.Piece {
	t.Helper()

	source, err := fs.Sub(assets.Packs, "packs")
	if err != nil {
		t.Fatal(err)
	}

	pieces, err := catalog.New(source).Everything(binding)
	if err != nil {
		t.Fatal(err)
	}

	return pieces
}

func frontmatter(content string) string {
	rest, found := strings.CutPrefix(content, "---\n")
	if !found {
		return ""
	}

	if strings.HasPrefix(rest, "---\n") {
		return ""
	}

	head, _, closed := strings.Cut(rest, "\n---\n")
	if !closed {
		return ""
	}

	return head
}

func TestEveryShippedPieceNameHasSixLetters(t *testing.T) {
	for _, piece := range shipped(t) {
		if !sixLetters.MatchString(piece.Name) {
			t.Errorf("%s is not six lowercase letters", piece.Name)
		}
	}
}

func TestEveryShippedPieceStaysUnderItsLineLimit(t *testing.T) {
	for _, piece := range shipped(t) {
		limit := ruleLimit
		if piece.Kind == agents.PieceSkill {
			limit = skillLimit
		}

		lines := strings.Count(piece.Content, "\n")
		if lines > limit {
			t.Errorf("%s has %d lines, over %d", piece.Name, lines, limit)
		}
	}
}

func TestEveryShippedFrontmatterIsReadable(t *testing.T) {
	for _, piece := range shipped(t) {
		var head struct {
			Paths []string `yaml:"paths"`
		}

		text := frontmatter(piece.Content)
		if text == "" && !strings.HasPrefix(piece.Content, "---\n---\n") {
			t.Errorf("%s has no frontmatter", piece.Name)
			continue
		}

		if err := yaml.Unmarshal([]byte(text), &head); err != nil {
			t.Errorf("%s frontmatter does not parse: %v", piece.Name, err)
		}
	}
}

func pathsOf(t *testing.T, piece agents.Piece) []string {
	t.Helper()

	var head struct {
		Paths []string `yaml:"paths"`
	}
	text := frontmatter(piece.Content)
	if err := yaml.Unmarshal([]byte(text), &head); err != nil {
		t.Fatalf("%s: %v", piece.Name, err)
	}

	return head.Paths
}

func TestScopedRulesLoadOnlyOnTheirStackFiles(t *testing.T) {
	scoped := map[string]bool{
		"record": true,
		"pledge": true,
		"frugal": true,
		"locale": true,
		"vendor": true,
	}
	binding := catalog.DefaultBinding().ForStacks([]string{"go"})

	for _, piece := range shippedWith(t, binding) {
		if piece.Kind != agents.PieceRule {
			continue
		}

		got := pathsOf(t, piece)
		if scoped[piece.Name] && len(got) == 0 {
			t.Errorf("%s has no paths with a stack", piece.Name)
		}
		if !scoped[piece.Name] && len(got) != 0 {
			t.Errorf("%s must load always, has paths %v", piece.Name, got)
		}
	}
}

func TestWithoutAStackTheStackRulesLoadAlways(t *testing.T) {
	for _, piece := range shipped(t) {
		if piece.Name == "record" || piece.Kind != agents.PieceRule {
			continue
		}

		if got := pathsOf(t, piece); len(got) != 0 {
			t.Errorf("%s has paths %v without a stack", piece.Name, got)
		}
	}
}

func TestCommitRuleCarriesTheConfiguredLimits(t *testing.T) {
	binding := catalog.DefaultBinding()
	binding.Commit = catalog.Commit{Min: 30, Max: 72}

	for _, piece := range shippedWith(t, binding) {
		if piece.Name != "commit" {
			continue
		}

		if !strings.Contains(piece.Content, "between 30 and 72") {
			t.Errorf("limits missing: %s", piece.Content)
		}
	}
}

func TestLocaleRuleAddsTheInterfaceOnlyWithAUILanguage(t *testing.T) {
	binding := catalog.DefaultBinding()
	binding.Language.UI = "es"

	for _, piece := range shippedWith(t, binding) {
		if piece.Name == "locale" && !strings.Contains(piece.Content, "`es`") {
			t.Errorf("ui language missing: %s", piece.Content)
		}
	}
	for _, piece := range shipped(t) {
		hasUI := strings.Contains(piece.Content, "interface")
		if piece.Name == "locale" && hasUI {
			t.Errorf("ui section without a ui language: %s", piece.Content)
		}
	}
}

var commandNames = []string{
	"attest", "brief", "check", "claim", "close", "context", "diff",
	"doctor", "find", "fmt", "get", "help", "hook", "init", "install",
	"ready", "record", "snares", "status", "verify",
}

func TestNoSkillIsNamedLikeABeaconCommand(t *testing.T) {
	for _, piece := range shipped(t) {
		if piece.Kind != agents.PieceSkill {
			continue
		}

		if slices.Contains(commandNames, piece.Name) {
			t.Errorf("skill %s collides with a command", piece.Name)
		}
	}
}

func TestEverySkillDescriptionFitsTheListingBudget(t *testing.T) {
	for _, piece := range shipped(t) {
		if piece.Kind != agents.PieceSkill {
			continue
		}

		var head struct {
			Description string `yaml:"description"`
		}
		text := frontmatter(piece.Content)
		if err := yaml.Unmarshal([]byte(text), &head); err != nil {
			t.Fatalf("%s: %v", piece.Name, err)
		}

		if len(head.Description) == 0 || len(head.Description) > 140 {
			t.Errorf("%s description has %d chars", piece.Name,
				len(head.Description))
		}
	}
}

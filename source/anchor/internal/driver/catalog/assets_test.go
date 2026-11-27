package catalog_test

import (
	"io/fs"
	"regexp"
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

	source, err := fs.Sub(assets.Packs, "packs")
	if err != nil {
		t.Fatal(err)
	}

	pieces, err := catalog.New(source).Everything(catalog.DefaultBinding())
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

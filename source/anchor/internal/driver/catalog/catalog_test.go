package catalog_test

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/driver/catalog"
)

func sample() fstest.MapFS {
	return fstest.MapFS{
		"core/rules/plain.md": {Data: []byte("---\n---\n# Plain\n")},
		"core/rules/limit.md.tmpl": {
			Data: []byte("---\n---\nmax {{.Commit.Max}}\n"),
		},
		"core/skills/alpha/SKILL.md": {Data: []byte("---\n---\n# Alpha\n")},
		"git/rules/fancy.md":         {Data: []byte("---\n---\n# Fancy\n")},
	}
}

func names(pieces []agents.Piece) []string {
	found := make([]string, 0, len(pieces))
	for _, piece := range pieces {
		found = append(found, piece.Pack+"/"+piece.Name)
	}

	return found
}

func codeOf(t *testing.T, err error) action.InstallCode {
	t.Helper()

	var failure action.InstallError
	if !errors.As(err, &failure) {
		t.Fatalf("err = %v", err)
	}

	return failure.Code
}

func TestSelectReturnsOnlyThePacksAsked(t *testing.T) {
	tool := catalog.New(sample())

	pieces, err := tool.Select([]string{"core"}, nil, catalog.DefaultBinding())
	if err != nil {
		t.Fatal(err)
	}

	got := names(pieces)
	want := []string{"core/limit", "core/plain", "core/alpha"}
	if len(got) != len(want) {
		t.Fatalf("pieces = %v", got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("pieces = %v, want %v", got, want)
		}
	}
}

func TestSelectRendersTemplatesWithTheBinding(t *testing.T) {
	tool := catalog.New(sample())

	pieces, err := tool.Select([]string{"core"}, nil, catalog.DefaultBinding())
	if err != nil {
		t.Fatal(err)
	}

	if pieces[0].Content != "---\n---\nmax 50\n" {
		t.Errorf("content = %q", pieces[0].Content)
	}
	rule, skill := pieces[0].Kind, pieces[2].Kind
	if rule != agents.PieceRule || skill != agents.PieceSkill {
		t.Errorf("kinds = %v, %v", rule, skill)
	}
}

func TestSelectLeavesOutTheSkippedPieces(t *testing.T) {
	tool := catalog.New(sample())

	pieces, err := tool.Select(
		[]string{"core", "git"},
		[]string{"plain"},
		catalog.DefaultBinding(),
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range names(pieces) {
		if name == "core/plain" {
			t.Errorf("plain was not skipped: %v", names(pieces))
		}
	}
	if len(pieces) != 3 {
		t.Errorf("pieces = %v", names(pieces))
	}
}

func TestSelectRejectsAnUnknownPackOrPiece(t *testing.T) {
	tool := catalog.New(sample())

	_, err := tool.Select([]string{"nope"}, nil, catalog.DefaultBinding())
	if codeOf(t, err) != action.PackUnknown {
		t.Errorf("pack: err = %v", err)
	}

	_, err = tool.Select(
		[]string{"core"},
		[]string{"nope"},
		catalog.DefaultBinding(),
	)
	if codeOf(t, err) != action.PieceUnknown {
		t.Errorf("piece: err = %v", err)
	}
}

func TestSkippingAPieceOfAnotherPackIsAllowed(t *testing.T) {
	tool := catalog.New(sample())

	_, err := tool.Select(
		[]string{"core"},
		[]string{"fancy"},
		catalog.DefaultBinding(),
	)
	if err != nil {
		t.Errorf("err = %v", err)
	}
}

func TestEverythingReturnsEveryPack(t *testing.T) {
	tool := catalog.New(sample())

	pieces, err := tool.Everything(catalog.DefaultBinding())
	if err != nil || len(pieces) != 4 {
		t.Errorf("pieces = %v, err %v", names(pieces), err)
	}
}

func TestATemplateWithAnUnknownKeyIsAnAssetFailure(t *testing.T) {
	source := fstest.MapFS{
		"core/rules/broken.md.tmpl": {Data: []byte("---\n---\n{{.Nope}}\n")},
	}
	tool := catalog.New(source)

	_, err := tool.Everything(catalog.DefaultBinding())
	if codeOf(t, err) != action.PackAssetInvalid {
		t.Errorf("err = %v", err)
	}
}

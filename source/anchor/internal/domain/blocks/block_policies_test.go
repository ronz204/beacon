package blocks_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

const sample = "<!-- BEGIN BEACON (managed) -->\nrule\n<!-- END BEACON -->"

func mustBlock(t *testing.T, raw string) blocks.Block {
	t.Helper()

	block, err := blocks.ParseBlock(raw)
	if err != nil {
		t.Fatal(err)
	}

	return block
}

func mustUpsert(t *testing.T, text string, block blocks.Block) string {
	t.Helper()

	result, err := blocks.Upsert(text, block)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func TestUpsertIsIdempotent(t *testing.T) {
	block := mustBlock(t, sample)
	texts := []string{"", "# Notes\n", "# Notes", "# Notes\r\n"}

	for _, text := range texts {
		once := mustUpsert(t, text, block)
		twice := mustUpsert(t, once, block)

		if once != twice {
			t.Errorf("text %q: upsert changed %q to %q", text, once, twice)
		}
	}
}

func TestUpsertAppendsAfterABlankLine(t *testing.T) {
	block := mustBlock(t, sample)

	got := mustUpsert(t, "# Notes\n", block)

	want := "# Notes\n\n" + sample + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpsertKeepsTheLineEndingOfTheFile(t *testing.T) {
	block := mustBlock(t, sample)

	got := mustUpsert(t, "# Notes\r\n", block)

	lone := strings.ReplaceAll(got, "\r\n", "")
	if strings.Contains(lone, "\n") {
		t.Errorf("mixed line endings in %q", got)
	}
}

func TestUpsertReplacesAStaleBlockAndKeepsTheRest(t *testing.T) {
	block := mustBlock(t, sample)
	stale := "top\n<!-- BEGIN BEACON (old) -->\nold\n" +
		"<!-- END BEACON -->\nbottom\n"

	got := mustUpsert(t, stale, block)

	want := "top\n" + sample + "\nbottom\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpsertFailsWhenTheEndMarkerIsMissing(t *testing.T) {
	block := mustBlock(t, sample)

	_, err := blocks.Upsert("<!-- BEGIN BEACON (old) -->\nold\n", block)

	var failure blocks.BlockError
	if !errors.As(err, &failure) || failure.Code != blocks.BlockUnclosed {
		t.Errorf("err = %v, want %s", err, blocks.BlockUnclosed)
	}
}

func TestRemoveRestoresTheOriginalFile(t *testing.T) {
	block := mustBlock(t, sample)
	originals := []string{"", "# Notes\n", "# Notes\r\n\r\ntext\r\n"}

	for _, original := range originals {
		installed := mustUpsert(t, original, block)

		got, status, err := blocks.Remove(installed, block)
		if err != nil {
			t.Fatal(err)
		}

		if status != blocks.RemovalRemoved || got != original {
			t.Errorf("original %q: got %q (%s)", original, got, status)
		}
	}
}

func TestRemoveKeepsContentAfterTheBlock(t *testing.T) {
	block := mustBlock(t, sample)
	text := "top\n\n" + sample + "\nbottom\n"

	got, status, err := blocks.Remove(text, block)
	if err != nil {
		t.Fatal(err)
	}

	if status != blocks.RemovalRemoved || got != "top\n\nbottom\n" {
		t.Errorf("got %q (%s)", got, status)
	}
}

func TestRemoveKeepsABlockThePersonEdited(t *testing.T) {
	block := mustBlock(t, sample)
	edited := "<!-- BEGIN BEACON (managed) -->\nmy rule\n<!-- END BEACON -->\n"

	got, status, err := blocks.Remove(edited, block)
	if err != nil {
		t.Fatal(err)
	}

	if status != blocks.RemovalEdited || got != edited {
		t.Errorf("got %q (%s)", got, status)
	}
}

func TestRemoveReportsAnAbsentBlock(t *testing.T) {
	block := mustBlock(t, sample)

	got, status, err := blocks.Remove("# Notes\n", block)
	if err != nil {
		t.Fatal(err)
	}

	if status != blocks.RemovalAbsent || got != "# Notes\n" {
		t.Errorf("got %q (%s)", got, status)
	}
}

func TestHashStyleBlocksRoundTrip(t *testing.T) {
	block := mustBlock(t, "# BEGIN BEACON\nrule\n# END BEACON")
	original := "*.go text\n"

	installed := mustUpsert(t, original, block)
	got, status, err := blocks.Remove(installed, block)
	if err != nil {
		t.Fatal(err)
	}

	if status != blocks.RemovalRemoved || got != original {
		t.Errorf("got %q (%s)", got, status)
	}
}

func TestCoversRecognizesTheSameLinesWrittenByHand(t *testing.T) {
	block := mustBlock(t, "# BEGIN BEACON\nrule\n# END BEACON")

	cases := map[string]bool{
		"*.go text\nrule\n":                    true,
		"rule\r\n":                             true,
		"*.go text\n":                          false,
		"":                                     false,
		"# BEGIN BEACON\nrule\n# END BEACON\n": false,
	}

	for text, want := range cases {
		if got := blocks.Covers(text, block); got != want {
			t.Errorf("Covers(%q) = %v, want %v", text, got, want)
		}
	}
}

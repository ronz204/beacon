package blocks_test

import (
	"errors"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

func mustEntry(
	t *testing.T,
	mode blocks.EntryMode,
	value any,
	path ...string,
) blocks.Entry {
	t.Helper()

	entry, err := blocks.ParseEntry(mode, value, path...)
	if err != nil {
		t.Fatal(err)
	}

	return entry
}

func sampleEntries(t *testing.T) []blocks.Entry {
	t.Helper()

	hook := map[string]any{"command": "beacon hook start"}

	return []blocks.Entry{
		mustEntry(t, blocks.EntryAppend, allowRule, "permissions", "allow"),
		mustEntry(t, blocks.EntrySet, "agent:sample", "env", "BEACON_ACTOR"),
		mustEntry(t, blocks.EntryAppend, hook, "hooks", "SessionStart"),
	}
}

const allowRule = "Bash(beacon *)"

const existing = `{
  "$schema": "https://example.test/schema.json",
  "permissions": {
    "allow": [
      "Bash(go test *)"
    ],
    "deny": [
      "Read(**/.env)"
    ]
  }
}
`

func TestMergeKeepsWhatWasThereAndItsOrder(t *testing.T) {
	got, err := blocks.MergeEntries(existing, sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "$schema": "https://example.test/schema.json",
  "permissions": {
    "allow": [
      "Bash(go test *)",
      "Bash(beacon *)"
    ],
    "deny": [
      "Read(**/.env)"
    ]
  },
  "env": {
    "BEACON_ACTOR": "agent:sample"
  },
  "hooks": {
    "SessionStart": [
      {
        "command": "beacon hook start"
      }
    ]
  }
}
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestMergeIsIdempotent(t *testing.T) {
	once, err := blocks.MergeEntries(existing, sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	twice, err := blocks.MergeEntries(once, sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	if once != twice {
		t.Errorf("second merge changed the file:\n%s", twice)
	}
}

func TestMergeLeavesBytesAloneWhenNothingChanges(t *testing.T) {
	compact := `{"permissions":{"allow":["Bash(beacon *)"]}}`
	entry := mustEntry(
		t, blocks.EntryAppend, "Bash(beacon *)", "permissions", "allow",
	)

	got, err := blocks.MergeEntries(compact, []blocks.Entry{entry})
	if err != nil || got != compact {
		t.Errorf("got %q, err %v", got, err)
	}
}

func TestMergeKeepsTabsAndCRLF(t *testing.T) {
	text := "{\r\n\t\"a\": 1\r\n}\r\n"
	entry := mustEntry(t, blocks.EntrySet, "x", "b")

	got, err := blocks.MergeEntries(text, []blocks.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}

	want := "{\r\n\t\"a\": 1,\r\n\t\"b\": \"x\"\r\n}\r\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeBuildsANewFile(t *testing.T) {
	entry := mustEntry(t, blocks.EntrySet, "x", "env", "A")

	got, err := blocks.MergeEntries("", []blocks.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}

	want := "{\n  \"env\": {\n    \"A\": \"x\"\n  }\n}\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMergeRefusesToOverwriteADifferentValue(t *testing.T) {
	text := `{"env": {"ACTOR": "human:jane"}}`
	entry := mustEntry(t, blocks.EntrySet, "agent:sample", "env", "ACTOR")

	_, err := blocks.MergeEntries(text, []blocks.Entry{entry})

	var failure blocks.EntryError
	if !errors.As(err, &failure) || failure.Code != blocks.EntryConflict {
		t.Errorf("err = %v, want %s", err, blocks.EntryConflict)
	}
}

func TestMergeRejectsFilesThatAreNotAnObject(t *testing.T) {
	entry := mustEntry(t, blocks.EntrySet, "x", "a")

	for _, text := range []string{`[1]`, `{"a":`, `{} {}`, `nope`} {
		_, err := blocks.MergeEntries(text, []blocks.Entry{entry})

		var failure blocks.EntryError
		if !errors.As(err, &failure) || failure.Code != blocks.DocumentInvalid {
			t.Errorf("text %q: err = %v", text, err)
		}
	}
}

func TestRemoveLeavesTheFileAsItWas(t *testing.T) {
	installed, err := blocks.MergeEntries(existing, sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	got, kept, err := blocks.RemoveEntries(installed, sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	if got != existing || len(kept) != 0 {
		t.Errorf("got:\n%s\nkept %d", got, len(kept))
	}
}

func TestRemoveKeepsAValueThePersonChanged(t *testing.T) {
	text := `{"env": {"ACTOR": "human:jane"}}`
	entry := mustEntry(t, blocks.EntrySet, "agent:sample", "env", "ACTOR")

	got, kept, err := blocks.RemoveEntries(text, []blocks.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}

	if got != text || len(kept) != 1 {
		t.Errorf("got %q, kept %d", got, len(kept))
	}
}

func TestRemoveEmptiesTheFileItself(t *testing.T) {
	installed, err := blocks.MergeEntries("", sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	got, _, err := blocks.RemoveEntries(installed, sampleEntries(t))
	if err != nil {
		t.Fatal(err)
	}

	empty, err := blocks.IsEmptyDocument(got)
	if err != nil || !empty {
		t.Errorf("empty = %v, err %v, text %q", empty, err, got)
	}
}

func TestParseEntryRejectsBadInput(t *testing.T) {
	cases := []struct {
		mode  blocks.EntryMode
		value any
		path  []string
	}{
		{blocks.EntrySet, "x", nil},
		{blocks.EntrySet, "x", []string{"a", ""}},
		{blocks.EntryMode("other"), "x", []string{"a"}},
		{blocks.EntrySet, make(chan int), []string{"a"}},
	}

	for _, c := range cases {
		if _, err := blocks.ParseEntry(c.mode, c.value, c.path...); err == nil {
			t.Errorf("ParseEntry(%v, %v) should fail", c.mode, c.path)
		}
	}
}

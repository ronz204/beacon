package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestParseRecordIDAccepts(t *testing.T) {
	valid := []string{
		"bc-0001",
		"bc-a1b2",
		"bc-a1b2c",
		"bc-91cd.1",
		"bc-91cd.1.1",
		"bc-91cd.10.2",
	}

	for _, raw := range valid {
		t.Run(raw, func(t *testing.T) {
			got, err := kernel.ParseRecordID(raw)

			if err != nil {
				t.Fatal(err)
			}
			if got.String() != raw {
				t.Fatalf("got %q", got)
			}
		})
	}
}

func TestParseRecordIDRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"bc-",
		"bc-abc",
		"bc-A1B2",
		"BC-a1b2",
		"xx-a1b2",
		"bc-a1g2",
		"bc-a1b2.",
		"bc-a1b2.0",
		"bc-a1b2.01",
		"bc-a1b2.x",
		"bc-a1b2..1",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParseRecordID(raw)

			requireKernelCode(t, err, kernel.KernelRecordIDInvalid)
		})
	}
}

func TestRecordIDRoot(t *testing.T) {
	cases := map[string]string{
		"bc-91cd":     "bc-91cd",
		"bc-91cd.1":   "bc-91cd",
		"bc-91cd.1.2": "bc-91cd",
	}

	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			id, err := kernel.ParseRecordID(raw)
			if err != nil {
				t.Fatal(err)
			}

			if id.Root().String() != want {
				t.Fatalf("root = %q, want %q", id.Root(), want)
			}
		})
	}
}

func TestRecordIDParent(t *testing.T) {
	cases := []struct {
		raw    string
		parent string
		found  bool
	}{
		{"bc-91cd", "", false},
		{"bc-91cd.1", "bc-91cd", true},
		{"bc-91cd.1.2", "bc-91cd.1", true},
	}

	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			id, err := kernel.ParseRecordID(c.raw)
			if err != nil {
				t.Fatal(err)
			}

			parent, found := id.Parent()

			if found != c.found {
				t.Fatalf("found = %v, want %v", found, c.found)
			}
			if found && parent.String() != c.parent {
				t.Fatalf("parent = %q, want %q", parent, c.parent)
			}
		})
	}
}

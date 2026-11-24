package record_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/internal/spider/internal/driver/record"
)

func addFixtureSeeds(f *testing.F, folders ...string) {
	f.Helper()

	for _, folder := range folders {
		paths, err := filepath.Glob(filepath.Join("testdata", folder, "*"))
		if err != nil {
			f.Fatal(err)
		}

		for _, path := range paths {
			if data, err := os.ReadFile(path); err == nil {
				f.Add(data)
			}
		}
	}
}

func FuzzFormatRecord(f *testing.F) {
	addFixtureSeeds(f, "canonical", "quoting", "unknown", "invalid/goals")
	f.Add([]byte("id: bc-0001\nnongoals: ['a: b', \"x\\ny\"]\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		first, err := record.FormatRecord("fuzz.yaml", caps, data)
		if err != nil {
			return
		}

		second, err := record.FormatRecord("fuzz.yaml", caps, first)
		if err != nil {
			t.Fatalf("formatted output does not parse: %v\n%s", err, first)
		}
		if string(second) != string(first) {
			t.Fatalf("not idempotent:\n%s\n---\n%s", first, second)
		}
	})
}

func FuzzLoadCapability(f *testing.F) {
	addFixtureSeeds(f, "invalid/capabilities", "unknown")
	f.Add([]byte("id: bc-7f3a\npromises: [{check: {runner: [x]}}]\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		root := t.TempDir()
		writeRecord(t, root, caps, "bc-7f3a-sample.yaml", data)

		store := record.NewStore(root)
		capabilities, failures := store.ListCapabilities(t.Context())

		if len(capabilities)+len(failures) == 0 {
			t.Fatal("a file must load or fail")
		}
	})
}

func FuzzFormatGuide(f *testing.F) {
	f.Add([]byte("---\ntitle: Sample\nkind: domain\n---\nBody\n"))
	f.Add([]byte("---\r\nkind: domain\r\n---\r\n---\r\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		first, err := record.FormatGuide("fuzz.md", data)
		if err != nil {
			return
		}

		second, err := record.FormatGuide("fuzz.md", first)
		if err != nil {
			t.Fatalf("formatted output does not parse: %v\n%q", err, first)
		}
		if string(second) != string(first) {
			t.Fatalf("not idempotent:\n%q\n---\n%q", first, second)
		}
	})
}

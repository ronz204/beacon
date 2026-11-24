package record_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
	"github.com/ronz204/beacon/internal/spider/internal/driver/record"
)

func storeWithConfig(t *testing.T, text string) record.Store {
	t.Helper()

	root := t.TempDir()
	path := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	return record.NewStore(root)
}

func TestLoadConfigReadsValues(t *testing.T) {
	store := record.NewStore(filepath.Join("testdata", "valid"))

	config, failures := store.LoadConfig()

	if failures != nil {
		t.Fatalf("failures = %v", failures)
	}
	if config.Schema != 1 {
		t.Fatalf("schema = %d", config.Schema)
	}
	if config.Limits.Guide != 150 {
		t.Fatalf("guide = %d", config.Limits.Guide)
	}
	if config.Review["architecture"].String() != "12mo" {
		t.Fatalf("review = %v", config.Review)
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	config, failures := storeWithConfig(t, "schema: 1\n").LoadConfig()

	if failures != nil {
		t.Fatalf("failures = %v", failures)
	}

	limits := record.ConfigLimits{
		Guide: 150,
		Brief: 60,
		Hook:  20,
	}
	if config.Limits != limits {
		t.Fatalf("limits = %+v", config.Limits)
	}

	review := map[string]string{
		"convention":   "6mo",
		"architecture": "12mo",
		"product":      "12mo",
		"domain":       "12mo",
		"runbook":      "6mo",
		"decision":     "6mo",
	}
	for key, want := range review {
		if got := config.Review[key].String(); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestLoadConfigOverridesOnlyWhatItDeclares(t *testing.T) {
	text := "schema: 1\nlimits:\n  brief: 40\nreview:\n  runbook: 30d\n"

	config, failures := storeWithConfig(t, text).LoadConfig()

	if failures != nil {
		t.Fatalf("failures = %v", failures)
	}
	if config.Limits.Brief != 40 || config.Limits.Hook != 20 {
		t.Fatalf("limits = %+v", config.Limits)
	}
	if config.Review["runbook"].String() != "30d" {
		t.Fatalf("runbook = %q", config.Review["runbook"])
	}
	if config.Review["domain"].String() != "12mo" {
		t.Fatalf("domain = %q", config.Review["domain"])
	}
}

func TestLoadConfigToleratesUnknownFields(t *testing.T) {
	path := filepath.Join("testdata", "unknown", "config.yaml")
	store := storeWithConfig(t, string(readFixture(t, path)))

	if _, failures := store.LoadConfig(); failures != nil {
		t.Fatalf("failures = %v", failures)
	}
}

func TestLoadConfigRejectsInvalidFixtures(t *testing.T) {
	cases := []struct {
		name string
		code record.RecordCode
	}{
		{"schema_newer", record.RecordSchemaNewer},
		{"schema_missing", record.RecordSchemaMissing},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name := c.name + ".yaml"
			path := filepath.Join("testdata", "invalid", "config", name)
			store := storeWithConfig(t, string(readFixture(t, path)))

			_, failures := store.LoadConfig()

			if len(failures) != 1 || !is(c.code)(failures[0]) {
				t.Fatalf("failures = %v", failures)
			}
		})
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		matches func(error) bool
	}{
		{"schema text", "schema: one\n", is(record.RecordFieldType)},
		{"schema zero", "schema: 0\n", is(record.RecordFieldType)},
		{"schema map", "schema: {a: 1}\n", is(record.RecordFieldType)},
		{
			"limit zero",
			"schema: 1\nlimits: {brief: 0}\n",
			is(record.RecordFieldType),
		},
		{
			"limits not map",
			"schema: 1\nlimits: 10\n",
			is(record.RecordFieldType),
		},
		{
			"review span",
			"schema: 1\nreview: {runbook: soon}\n",
			is(memory.ReviewSpanInvalid),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, failures := storeWithConfig(t, c.text).LoadConfig()

			if len(failures) != 1 || !c.matches(failures[0]) {
				t.Fatalf("failures = %v", failures)
			}
		})
	}
}

func TestLoadConfigReportsMissingFile(t *testing.T) {
	_, failures := record.NewStore(t.TempDir()).LoadConfig()

	if len(failures) != 1 || !errors.Is(failures[0], fs.ErrNotExist) {
		t.Fatalf("failures = %v", failures)
	}
}

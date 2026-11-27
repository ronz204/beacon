package record_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
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

const packsConfig = `schema: 1
packs:
  enabled: [core, git]
  skip: [refute]
  stacks: [go]
  language:
    code: en
    ui: es
  commit:
    min: 40
    max: 50
`

func TestLoadConfigReadsThePacksSection(t *testing.T) {
	config, failures := storeWithConfig(t, packsConfig).LoadConfig()

	if failures != nil {
		t.Fatalf("failures = %v", failures)
	}

	packs := config.Packs
	if !packs.Present || len(packs.Enabled) != 2 || packs.Skip[0] != "refute" {
		t.Fatalf("packs = %+v", packs)
	}
	if packs.Stacks[0] != "go" || packs.Language.UI != "es" {
		t.Fatalf("packs = %+v", packs)
	}
	if packs.Commit.Min != 40 || packs.Commit.Max != 50 {
		t.Fatalf("commit = %+v", packs.Commit)
	}
}

func TestLoadConfigWithoutPacksIsNotPresent(t *testing.T) {
	config, failures := storeWithConfig(t, "schema: 1\n").LoadConfig()

	if failures != nil || config.Packs.Present {
		t.Fatalf("packs = %+v, failures = %v", config.Packs, failures)
	}
}

func TestLoadConfigRejectsInvalidPacks(t *testing.T) {
	cases := []struct {
		name string
		text string
		code record.RecordCode
	}{
		{"pack name", "schema: 1\npacks: {enabled: [Core]}\n", packsInvalid},
		{"piece name", "schema: 1\npacks: {skip: [refute1]}\n", packsInvalid},
		{"stack name", "schema: 1\npacks: {stacks: [Go]}\n", packsInvalid},
		{
			"language",
			"schema: 1\npacks: {language: {code: EN}}\n",
			packsInvalid,
		},
		{"commit zero", "schema: 1\npacks: {commit: {min: 0}}\n", packsField},
		{
			"commit inverted",
			"schema: 1\npacks: {commit: {min: 60, max: 50}}\n",
			packsInvalid,
		},
		{"not a mapping", "schema: 1\npacks: 3\n", packsField},
		{"list not a list", "schema: 1\npacks: {enabled: core}\n", packsField},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, failures := storeWithConfig(t, c.text).LoadConfig()

			if len(failures) != 1 || !is(c.code)(failures[0]) {
				t.Fatalf("failures = %v", failures)
			}
		})
	}
}

const (
	packsInvalid = record.RecordPacksInvalid
	packsField   = record.RecordFieldType
)

func TestFormatConfigPutsThePacksSectionInOrder(t *testing.T) {
	input := "packs:\n  commit: {max: 50, min: 40}\n  stacks: [go]\n" +
		"  enabled: [core]\nschema: 1\nrunners: {}\n"
	want := "schema: 1\nrunners: {}\npacks:\n  enabled: [core]\n" +
		"  stacks: [go]\n  commit:\n    min: 40\n    max: 50\n"

	got, err := record.FormatConfig("config.yaml", []byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}

	again, err := record.FormatConfig("config.yaml", got)
	if err != nil || string(again) != want {
		t.Fatalf("not idempotent:\n%s", again)
	}
}

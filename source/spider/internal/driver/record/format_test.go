package record_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func formatText(t *testing.T, folder, text string) string {
	t.Helper()

	output, err := record.FormatRecord("sample.yaml", folder, []byte(text))
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}

func TestFormatRecordMatchesCanonicalFixtures(t *testing.T) {
	cases := []struct {
		name   string
		folder string
	}{
		{"key_order", "goals"},
		{"sorted_ids", caps},
		{"quotes", "goals"},
		{"flow_style", caps},
		{"unknown_fields", "goals"},
		{"request_block", "works"},
		{"request_indented", "works"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base := filepath.Join("testdata", "canonical", c.name)
			input := readFixture(t, base+".in.yaml")
			want := string(readFixture(t, base+".out.yaml"))

			got := formatText(t, c.folder, string(input))
			if got != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
			if again := formatText(t, c.folder, got); again != got {
				t.Fatalf("not idempotent:\n%s", again)
			}
		})
	}
}

func TestFormatKeepsValidFixturesUnchanged(t *testing.T) {
	root := filepath.Join("testdata", "valid", "records")

	err := filepath.WalkDir(root, func(
		path string,
		entry fs.DirEntry,
		err error,
	) error {
		if err != nil || entry.IsDir() {
			return err
		}

		data := readFixture(t, path)
		folder := filepath.Base(filepath.Dir(path))

		var output []byte
		if folder == "guides" {
			output, err = record.FormatGuide(path, data)
		} else {
			output, err = record.FormatRecord(path, folder, data)
		}
		if err != nil {
			return err
		}
		if string(output) != string(data) {
			t.Fatalf("%s changed:\n%s", path, output)
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func fixtureListValues(t *testing.T, name string) ([]string, []string) {
	t.Helper()

	path := filepath.Join("testdata", "quoting", name)
	data := readFixture(t, path)

	root, err := record.ParseRecord(path, data)
	if err != nil {
		t.Fatal(err)
	}

	var values []string
	for _, item := range mappingValue(t, root, "values").Content {
		values = append(values, item.Value)
	}

	var lines []string
	for line := range strings.Lines(string(data)) {
		if strings.HasPrefix(line, "  - ") {
			lines = append(lines, line)
		}
	}

	return values, lines
}

func TestFormatQuotesWhatCouldBeReinterpreted(t *testing.T) {
	values, lines := fixtureListValues(t, "must-quote.yaml")

	for index, value := range values {
		t.Run(value, func(t *testing.T) {
			quoted := strings.ReplaceAll(value, "'", "''")
			input := "nongoals:\n  - '" + quoted + "'\n"

			got := formatText(t, "goals", input)

			if got != "nongoals:\n"+lines[index] {
				t.Fatalf("got %q, want %q", got, lines[index])
			}
		})
	}
}

func TestFormatLeavesSafeTextPlain(t *testing.T) {
	values, lines := fixtureListValues(t, "stay-plain.yaml")

	for index, value := range values {
		t.Run(value, func(t *testing.T) {
			input := "nongoals:\n  - \"" + value + "\"\n"

			got := formatText(t, "goals", input)

			if got != "nongoals:\n"+lines[index] {
				t.Fatalf("got %q, want %q", got, lines[index])
			}
		})
	}
}

func TestFormatQuotesFlowCharacters(t *testing.T) {
	path := filepath.Join("testdata", "quoting", "flow.yaml")
	lines := slices.Collect(strings.Lines(string(readFixture(t, path))))
	input := lines[1] + "promises:\n  - id: bc-c3d4\n    " + lines[0]

	got := formatText(t, caps, input)

	if !strings.Contains(got, lines[1]) {
		t.Fatalf("paths changed:\n%s", got)
	}
	if !strings.Contains(got, "    "+lines[0]) {
		t.Fatalf("check changed:\n%s", got)
	}
}

func TestFormatKeepsBooleanFieldsPlain(t *testing.T) {
	text := "---\nalways: true\nkind: domain\ntitle: Sample\n---\nBody\n"

	output, err := record.FormatGuide("guide.md", []byte(text))
	if err != nil {
		t.Fatal(err)
	}

	want := "---\ntitle: Sample\nkind: domain\nalways: true\n---\nBody\n"
	if string(output) != want {
		t.Fatalf("got:\n%s", output)
	}
}

func TestFormatGuideKeepsBodyBytes(t *testing.T) {
	body := "  Indented\n\n---\nEnd without newline"
	data := []byte("---\nkind: domain\ntitle: Sample\n---\n" + body)

	output, err := record.FormatGuide("guide.md", data)
	if err != nil {
		t.Fatal(err)
	}

	want := "---\ntitle: Sample\nkind: domain\n---\n" + body
	if string(output) != want {
		t.Fatalf("got:\n%s", output)
	}
}

func TestFormatNeverWrapsLongValues(t *testing.T) {
	long := strings.Repeat("palabra ", 40) + "fin"
	input := "id: bc-7f3a\nintent: " + long + "\npromises:\n" +
		"  - id: bc-c3d4\n    check: {manual: " + long + "}\n"

	got := formatText(t, caps, input)

	if !strings.Contains(got, "intent: "+long+"\n") {
		t.Fatalf("intent wrapped:\n%s", got)
	}
	if !strings.Contains(got, "check: {manual: "+long+"}\n") {
		t.Fatalf("check wrapped:\n%s", got)
	}
}

func TestFormatKeepsCommentedSectionsAsBlocks(t *testing.T) {
	input := "id: bc-7f3a\npaths:\n  # main code\n  - src/**\npromises:\n" +
		"  - id: bc-c3d4\n    check:\n      # owner team\n      runner: x\n"

	got := formatText(t, caps, input)

	if !strings.Contains(got, "paths:\n  # main code\n  - src/**\n") {
		t.Fatalf("paths changed:\n%s", got)
	}
	if !strings.Contains(got, "    check:\n      # owner team\n") {
		t.Fatalf("check changed:\n%s", got)
	}
	if again := formatText(t, caps, got); again != got {
		t.Fatalf("not idempotent:\n%s", again)
	}
}

func TestFormatKeepsComments(t *testing.T) {
	input := "# head note\n\nstatus: active # inline note\nid: bc-0001\n"

	got := formatText(t, "goals", input)

	if !strings.Contains(got, "# head note") {
		t.Fatalf("head comment lost:\n%s", got)
	}
	if !strings.Contains(got, "status: active # inline note") {
		t.Fatalf("inline comment lost:\n%s", got)
	}
	if again := formatText(t, "goals", got); again != got {
		t.Fatalf("not idempotent:\n%s", again)
	}
}

func TestFormatKeepsUnknownFixtureFields(t *testing.T) {
	path := filepath.Join("testdata", "unknown", "capability.yaml")

	got := formatText(t, caps, string(readFixture(t, path)))

	expected := []string{
		"device: laptop",
		"weight: 3",
		"owners: [equipo-auth, equipo-core]",
		"level: 2",
	}
	for _, text := range expected {
		if !strings.Contains(got, text) {
			t.Fatalf("%q lost:\n%s", text, got)
		}
	}
	if again := formatText(t, caps, got); again != got {
		t.Fatalf("not idempotent:\n%s", again)
	}
}

func TestStoreFormatChecksThenWrites(t *testing.T) {
	base := filepath.Join("testdata", "canonical", "key_order")
	want := readFixture(t, base+".out.yaml")
	messy := readFixture(t, base+".in.yaml")
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-0001-sample.yaml", messy)
	path := filepath.Join(root, "records", "goals", "bc-0001-sample.yaml")
	store := record.NewStore(root)
	ctx := context.Background()

	changed, failures := store.Format(ctx, false)
	if failures != nil || len(changed) != 1 {
		t.Fatalf("changed = %v, failures = %v", changed, failures)
	}
	if got := readFixture(t, path); string(got) != string(messy) {
		t.Fatal("check mode wrote the file")
	}

	changed, failures = store.Format(ctx, true)
	if failures != nil || len(changed) != 1 {
		t.Fatalf("changed = %v, failures = %v", changed, failures)
	}
	if got := readFixture(t, path); string(got) != string(want) {
		t.Fatalf("got:\n%s", got)
	}

	changed, failures = store.Format(ctx, true)
	if failures != nil || len(changed) != 0 {
		t.Fatalf("changed = %v, failures = %v", changed, failures)
	}
}

func TestStoreFormatNeverTouchesInvalidFiles(t *testing.T) {
	broken := []byte("status: paused\nid: bc-0001\n")
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-0001-sample.yaml", broken)
	path := filepath.Join(root, "records", "goals", "bc-0001-sample.yaml")

	changed, failures := record.NewStore(root).Format(
		context.Background(),
		true,
	)

	if len(changed) != 0 || len(failures) == 0 {
		t.Fatalf("changed = %v, failures = %v", changed, failures)
	}
	if got := readFixture(t, path); string(got) != string(broken) {
		t.Fatalf("invalid file changed:\n%s", got)
	}
}

func TestStoreFormatNormalizesWindowsLines(t *testing.T) {
	path := filepath.Join(
		"testdata", "valid", "records", "goals",
		"bc-0001-frictionless-payments.yaml",
	)
	want := readFixture(t, path)
	windows := strings.ReplaceAll(string(want), "\n", "\r\n")
	root := t.TempDir()
	writeRecord(t, root, "goals", "bc-0001-sample.yaml", []byte(windows))

	changed, failures := record.NewStore(root).Format(
		context.Background(),
		true,
	)
	if failures != nil || len(changed) != 1 {
		t.Fatalf("changed = %v, failures = %v", changed, failures)
	}

	target := filepath.Join(root, "records", "goals", "bc-0001-sample.yaml")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q", got)
	}
}

func TestStoreFormatCoversGuides(t *testing.T) {
	data := []byte("---\nkind: domain\ntitle: Sample\n---\nBody\n")
	root := t.TempDir()
	writeRecord(t, root, "guides", "bc-0a9b-sample.md", data)
	path := filepath.Join(root, "records", "guides", "bc-0a9b-sample.md")

	_, failures := record.NewStore(root).Format(context.Background(), true)
	if failures != nil {
		t.Fatalf("failures = %v", failures)
	}

	want := "---\ntitle: Sample\nkind: domain\n---\nBody\n"
	if got := readFixture(t, path); string(got) != want {
		t.Fatalf("got:\n%s", got)
	}
}

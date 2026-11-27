package record_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	yaml "go.yaml.in/yaml/v4"
)

func readFixture(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func requireFailure(
	t *testing.T,
	err error,
	code record.RecordCode,
	line int,
) {
	t.Helper()

	var got record.RecordError
	if !errors.As(err, &got) {
		t.Fatalf("expected RecordError, got %v", err)
	}
	if got.Code != code {
		t.Fatalf("code = %q, want %q", got.Code, code)
	}
	if got.Details["line"] != line {
		t.Fatalf("line = %v, want %d", got.Details["line"], line)
	}
}

func mappingKeys(root *yaml.Node) []string {
	keys := make([]string, 0, len(root.Content)/2)

	for index := 0; index < len(root.Content); index += 2 {
		keys = append(keys, root.Content[index].Value)
	}

	return keys
}

func mappingValue(t *testing.T, root *yaml.Node, key string) *yaml.Node {
	t.Helper()

	for index := 0; index < len(root.Content); index += 2 {
		if root.Content[index].Value == key {
			return root.Content[index+1]
		}
	}

	t.Fatalf("key %q not found", key)
	return nil
}

func TestParseRecordRejectsStructuralFixtures(t *testing.T) {
	cases := []struct {
		name string
		code record.RecordCode
		line int
	}{
		{"goals/anchor", record.RecordAnchorForbidden, 5},
		{"goals/alias", record.RecordAnchorForbidden, 4},
		{"goals/merge_key", record.RecordMergeForbidden, 8},
		{"goals/duplicate_key", record.RecordKeyRepeated, 9},
		{"capabilities/duplicate_nested_key", record.RecordKeyRepeated, 15},
		{"goals/custom_tag", record.RecordTagForbidden, 4},
		{"goals/multiple_documents", record.RecordDocumentExtra, 9},
		{"goals/malformed", record.RecordSyntaxInvalid, 5},
		{"goals/empty", record.RecordDocumentEmpty, 1},
		{"goals/not_a_map", record.RecordRootInvalid, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			file := filepath.Join("testdata", "invalid", c.name+".yaml")

			_, err := record.ParseRecord(file, readFixture(t, file))

			requireFailure(t, err, c.code, c.line)
		})
	}
}

func TestParseAcceptsEveryWellFormedFixture(t *testing.T) {
	folders := []string{
		"valid",
		"unknown",
		"quoting",
		"canonical",
	}

	for _, folder := range folders {
		root := filepath.Join("testdata", folder)

		err := filepath.WalkDir(root, func(
			path string,
			entry fs.DirEntry,
			err error,
		) error {
			if err != nil || entry.IsDir() {
				return err
			}

			data := readFixture(t, path)
			if strings.HasSuffix(path, ".md") {
				_, err = record.ParseGuide(path, data)
				return err
			}

			_, err = record.ParseRecord(path, data)
			return err
		})

		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseRecordKeepsRawScalars(t *testing.T) {
	data := []byte("version: 1.0\nflag: yes\nday: 2026-11-22\n")

	root, err := record.ParseRecord("sample.yaml", data)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"version": "1.0",
		"flag":    "yes",
		"day":     "2026-11-22",
	}
	for key, want := range expected {
		if got := mappingValue(t, root, key).Value; got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestParseRecordKeepsUnknownFields(t *testing.T) {
	file := filepath.Join("testdata", "unknown", "capability.yaml")

	root, err := record.ParseRecord(file, readFixture(t, file))
	if err != nil {
		t.Fatal(err)
	}

	keys := mappingKeys(root)
	if !slices.Contains(keys, "owners") {
		t.Fatalf("keys = %v", keys)
	}
	if !slices.Contains(keys, "metadata") {
		t.Fatalf("keys = %v", keys)
	}
}

func TestParseRecordRejectsInlineDefects(t *testing.T) {
	cases := []struct {
		name string
		text string
		code record.RecordCode
		line int
	}{
		{"complex key", "? {a: 1}\n: x\n", record.RecordKeyInvalid, 1},
		{"core tag", "a: !!binary aGk=\n", record.RecordTagForbidden, 1},
		{"unknown alias", "a: *nope\n", record.RecordSyntaxInvalid, 1},
		{"comment only", "# note\n", record.RecordDocumentEmpty, 1},
		{"null root", "---\n", record.RecordRootInvalid, 2},
		{"scalar root", "text\n", record.RecordRootInvalid, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := record.ParseRecord("sample.yaml", []byte(c.text))

			requireFailure(t, err, c.code, c.line)
		})
	}
}

func TestParseRecordNamesTheFile(t *testing.T) {
	_, err := record.ParseRecord("records/sample.yaml", []byte("a: &x 1\n"))

	var got kernel.Error[record.RecordCode]
	if !errors.As(err, &got) {
		t.Fatalf("expected RecordError, got %v", err)
	}
	if got.Details["file"] != "records/sample.yaml" {
		t.Fatalf("file = %v", got.Details["file"])
	}
}

func TestParseRecordKeepsSyntaxCause(t *testing.T) {
	_, err := record.ParseRecord("sample.yaml", []byte("a: [x\n"))

	if _, ok := errors.AsType[*yaml.LoadError](err); !ok {
		t.Fatalf("cause is not reachable: %v", err)
	}
}

func TestParseRecordAcceptsWindowsText(t *testing.T) {
	text := "\xef\xbb\xbfid: bc-0001\r\n" +
		"request: |\r\n  one\r\n  two\r\n"

	root, err := record.ParseRecord("sample.yaml", []byte(text))
	if err != nil {
		t.Fatal(err)
	}

	if got := mappingValue(t, root, "id").Value; got != "bc-0001" {
		t.Fatalf("id = %q", got)
	}
	if got := mappingValue(t, root, "request").Value; got != "one\ntwo\n" {
		t.Fatalf("request = %q", got)
	}
}

func TestParseGuideSplitsFrontmatterAndBody(t *testing.T) {
	cases := []struct {
		name string
		text string
		body string
	}{
		{"body", "---\ntitle: x\n---\none\n", "one\n"},
		{"no final newline", "---\ntitle: x\n---\none", "one"},
		{"closing at end", "---\ntitle: x\n---", ""},
		{"rule in body", "---\ntitle: x\n---\nu\n---\nd\n", "u\n---\nd\n"},
		{"windows", "---\r\ntitle: x\r\n---\r\none\r\n", "one\n"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			guide, err := record.ParseGuide("guide.md", []byte(c.text))

			if err != nil {
				t.Fatal(err)
			}
			if guide.Body != c.body {
				t.Fatalf("body = %q, want %q", guide.Body, c.body)
			}
			if got := mappingValue(t, guide.Root, "title").Value; got != "x" {
				t.Fatalf("title = %q", got)
			}
		})
	}
}

func TestParseGuideRejectsMissingFrontmatter(t *testing.T) {
	file := filepath.Join("testdata", "invalid", "guides", "no_frontmatter.md")

	_, err := record.ParseGuide(file, readFixture(t, file))
	requireFailure(t, err, record.RecordFrontmatterMissing, 1)

	_, err = record.ParseGuide("guide.md", []byte("---\ntitle: x\n"))
	requireFailure(t, err, record.RecordFrontmatterMissing, 1)
}

func TestParseGuideRejectsEmptyFrontmatter(t *testing.T) {
	_, err := record.ParseGuide("guide.md", []byte("---\n---\none\n"))

	requireFailure(t, err, record.RecordDocumentEmpty, 2)
}

func TestParseGuideReportsFileLines(t *testing.T) {
	data := []byte("---\ntitle: x\ntitle: y\n---\none\n")

	_, err := record.ParseGuide("guide.md", data)
	requireFailure(t, err, record.RecordKeyRepeated, 3)

	data = []byte("---\ntitle: x\nkind: [a\n---\none\n")

	_, err = record.ParseGuide("guide.md", data)

	var got record.RecordError
	if !errors.As(err, &got) || got.Code != record.RecordSyntaxInvalid {
		t.Fatalf("got %v", err)
	}
	if line, _ := got.Details["line"].(int); line < 3 {
		t.Fatalf("line = %v", got.Details["line"])
	}

	guide, err := record.ParseGuide(
		"guide.md",
		[]byte("---\nkind: domain\ntitle: x\n---\n"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := mappingValue(t, guide.Root, "title").Line; got != 3 {
		t.Fatalf("title line = %d", got)
	}
}

func TestParseRecordReportsGitConflictMarkers(t *testing.T) {
	data := []byte("id: bc-7f3a\n<<<<<<< HEAD\nname: A\n=======\nname: B\n" +
		">>>>>>> feat/other\n")

	_, err := record.ParseRecord("record.yaml", data)

	requireFailure(t, err, record.RecordConflict, 2)
}

func TestParseGuideReportsConflictMarkersInTheBody(t *testing.T) {
	data := []byte("---\nid: bc-0a9b\n---\n<<<<<<< HEAD\nOne\n=======\nTwo\n" +
		">>>>>>> feat/other\n")

	_, err := record.ParseGuide("guide.md", data)

	requireFailure(t, err, record.RecordConflict, 4)
}

func TestParseRecordIgnoresAnUnclosedMarker(t *testing.T) {
	data := []byte("id: bc-7f3a\nname: \"<<<<<<< not a conflict\"\n")

	if _, err := record.ParseRecord("record.yaml", data); err != nil {
		t.Fatal(err)
	}
}

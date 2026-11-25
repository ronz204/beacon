package system_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ronz204/beacon/source/system"
)

func TestWriteFileAtomicCreatesAndReplaces(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "sample.yaml")

	if err := system.WriteFileAtomic(path, []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if err := system.WriteFileAtomic(path, []byte("second\n")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second\n" {
		t.Fatalf("data = %q", data)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary file left behind: %v", entries)
	}
}

func TestWriteFileAtomicFailsWithoutDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "sample.yaml")

	if err := system.WriteFileAtomic(path, []byte("data")); err == nil {
		t.Fatal("expected an error")
	}
}

func TestRemoveFileDeletesAndReportsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := system.RemoveFile(path); err != nil {
		t.Fatal(err)
	}
	if err := system.RemoveFile(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("second remove = %v", err)
	}
}

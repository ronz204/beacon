package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func NewWorkspace(t *testing.T) string {
	t.Helper()

	source := filepath.Join(
		"..", "..", "source", "spider", "internal",
		"driver", "record", "testdata", "valid",
	)
	root := t.TempDir()
	target := filepath.Join(root, ".beacon")

	if err := os.CopyFS(target, os.DirFS(source)); err != nil {
		t.Fatal(err)
	}

	return root
}

func WriteFile(t *testing.T, root, relative, text string) string {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(relative))

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func ReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func Git(t *testing.T, directory string, args ...string) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = directory

	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

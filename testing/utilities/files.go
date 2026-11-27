package utilities

import (
	"os"
	"path/filepath"
	"testing"
)

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

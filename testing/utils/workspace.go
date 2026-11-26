package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func NewRepository(t *testing.T) string {
	t.Helper()

	root := NewWorkspace(t)

	Git(t, root, "init", "-q", "-b", "main")
	Git(t, root, "config", "user.name", "jane")
	Git(t, root, "config", "user.email", "jane@example.com")
	Git(t, root, "add", ".")
	Git(t, root, "commit", "-q", "-m", "fixtures")
	attestAtHead(t, root)

	return root
}

func attestAtHead(t *testing.T, root string) {
	t.Helper()

	command := exec.Command("git", "rev-parse", "--short", "HEAD")
	command.Dir = root

	head, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}

	file := ".beacon/records/capabilities/bc-7f3a-auth-login.yaml"
	path := filepath.Join(root, filepath.FromSlash(file))
	text := strings.Replace(
		ReadFile(t, path),
		"9f3c2ab",
		strings.TrimSpace(string(head)),
		1,
	)
	WriteFile(t, root, file, text)

	if formatted := Beacon(t, root, "fmt"); formatted.Code != 0 {
		t.Fatalf("fmt failed: %s", formatted.Stderr)
	}

	Git(t, root, "commit", "-q", "-am", "attest at head")
}

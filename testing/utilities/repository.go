package utilities

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

func NewRepository(t *testing.T) string {
	t.Helper()

	root := NewWorkspace(t)
	WriteFile(t, root, ".gitattributes", ".beacon/** text eol=lf\n")

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

package utilities

import (
	"os/exec"
	"testing"
)

func Git(t *testing.T, directory string, args ...string) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = directory

	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func GitTry(t *testing.T, directory string, args ...string) (string, error) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = directory

	output, err := command.CombinedOutput()

	return string(output), err
}

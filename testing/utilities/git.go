package utilities

import (
	"os"
	"os/exec"
	"path/filepath"
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

func GitWithBeacon(
	t *testing.T,
	directory string,
	args ...string,
) (string, error) {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = directory
	command.Env = append(
		os.Environ(),
		"PATH="+filepath.Dir(binaryPath)+string(os.PathListSeparator)+
			os.Getenv("PATH"),
	)

	output, err := command.CombinedOutput()

	return string(output), err
}

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var binaryPath string

type Result struct {
	Code   int
	Stdout string
	Stderr string
}

func RunSuite(m *testing.M) int {
	directory, err := os.MkdirTemp("", "beacon-e2e-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(directory)

	binaryPath = filepath.Join(directory, binaryName())
	if err := buildBinary(binaryPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return m.Run()
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "beacon.exe"
	}

	return "beacon"
}

func buildBinary(target string) error {
	command := exec.Command("go", "build", "-o", target, "./cmd/beacon")
	command.Dir = filepath.Join("..", "..")

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build beacon: %w\n%s", err, output)
	}

	return nil
}

func Beacon(t *testing.T, directory string, args ...string) Result {
	t.Helper()

	var stdout, stderr bytes.Buffer

	command := exec.Command(binaryPath, args...)
	command.Dir = directory
	command.Stdout = &stdout
	command.Stderr = &stderr

	code := 0
	if err := command.Run(); err != nil {
		var exited *exec.ExitError
		if !errors.As(err, &exited) {
			t.Fatal(err)
		}

		code = exited.ExitCode()
	}

	return Result{
		Code:   code,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}

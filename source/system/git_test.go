package system_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/system"
)

func TestGitRunsInTheGivenFolder(t *testing.T) {
	directory := t.TempDir()

	_, err := system.Git(context.Background(), directory, "init", "-q")
	if err != nil {
		t.Fatal(err)
	}

	output, err := system.Git(
		context.Background(),
		directory,
		"rev-parse",
		"--is-inside-work-tree",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(output) != "true" {
		t.Fatalf("output = %q", output)
	}
}

func TestGitReportsTheFailingCommand(t *testing.T) {
	_, err := system.Git(context.Background(), t.TempDir(), "log")

	if err == nil || !strings.Contains(err.Error(), "git log") {
		t.Fatalf("err = %v", err)
	}
}

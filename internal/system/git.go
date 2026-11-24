package system

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Git(
	ctx context.Context,
	directory string,
	args ...string,
) (string, error) {
	arguments := append([]string{"-c", "core.quotepath=off"}, args...)
	command := exec.CommandContext(ctx, "git", arguments...)
	command.Dir = directory

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		reason := strings.TrimSpace(stderr.String())
		invoked := strings.Join(args, " ")
		return "", fmt.Errorf("git %s: %w: %s", invoked, err, reason)
	}

	return stdout.String(), nil
}

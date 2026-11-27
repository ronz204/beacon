package system

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const gitFolder = ".git"

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

func GitCommonDir(ctx context.Context, root string) (string, error) {
	own := filepath.Join(root, gitFolder)
	if info, err := os.Stat(own); err == nil && info.IsDir() {
		return own, nil
	}

	output, err := Git(ctx, root, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}

	common := strings.TrimSpace(output)
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}

	return common, nil
}

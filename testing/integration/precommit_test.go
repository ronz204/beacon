package integration_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

type snaresData struct {
	Blocks []struct {
		Kind  string   `json:"kind"`
		Hint  string   `json:"hint"`
		Items []string `json:"items"`
	} `json:"blocks"`
	Warnings []string `json:"warnings"`
	Checked  []string `json:"checked"`
}

func runPreCommit(t *testing.T, root string) (utilities.Result, snaresData) {
	t.Helper()

	result := utilities.Beacon(t, root, "snares", "pre-commit", "--json")

	var data snaresData
	decodeData(t, result.Stdout, &data)

	return result, data
}

func blockKinds(data snaresData) []string {
	kinds := make([]string, 0, len(data.Blocks))
	for _, block := range data.Blocks {
		kinds = append(kinds, block.Kind)
	}

	return kinds
}

func TestPreCommitPassesACleanCommit(t *testing.T) {
	root := passingRepository(t)
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")

	result, data := runPreCommit(t, root)

	if result.Code != 0 || len(data.Blocks) != 0 {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if len(data.Checked) != 1 {
		t.Errorf("checked = %v", data.Checked)
	}
}

func TestPreCommitBlocksALevelThreeChangeWithoutAYes(t *testing.T) {
	root := passingRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")

	result, data := runPreCommit(t, root)

	kinds := blockKinds(data)
	if result.Code != 2 || len(kinds) == 0 || kinds[0] != "unapproved_change" {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}

	envelope := parseEnvelope(t, result.Stdout)
	if !hintContains(envelope.Hint, "--approve") {
		t.Errorf("hint = %v", envelope.Hint)
	}
}

func TestPreCommitBlocksARedCheck(t *testing.T) {
	root := repositoryWithRunner(t)
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")

	result, data := runPreCommit(t, root)

	kinds := blockKinds(data)
	if result.Code != 2 || len(kinds) != 1 || kinds[0] != "failed_check" {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
}

func TestPreCommitBlocksAnInvalidRecord(t *testing.T) {
	root := passingRepository(t)
	utilities.WriteFile(t, root, capabilityPath, "id: [broken\n")

	result, data := runPreCommit(t, root)

	kinds := blockKinds(data)
	if result.Code != 2 || len(kinds) != 1 || kinds[0] != "invalid_state" {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
}

func TestPreCommitOnARepositoryWithoutCommitsOnlyWarns(t *testing.T) {
	root := utilities.NewWorkspace(t)
	utilities.Git(t, root, "init", "-q", "-b", "main")

	result, data := runPreCommit(t, root)

	if result.Code != 0 || len(data.Warnings) != 1 {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
}

const hookFile = ".git/hooks/pre-commit"

func TestInstallWritesTheHookAndABlockedCommitStaysOut(t *testing.T) {
	root := passingRepository(t)
	runInstall(t, root)
	editByHand(t, root, "After 5 failed", "After 10 failed")

	blocked, err := utilities.GitWithBeacon(
		t,
		root,
		"commit", "-q", "-am", "weaken",
	)
	if err == nil || !strings.Contains(blocked, "level 3") {
		t.Fatalf("the commit was not blocked: %v\n%s", err, blocked)
	}

	skipped, err := utilities.GitWithBeacon(
		t,
		root,
		"commit", "-q", "--no-verify", "-am", "weaken",
	)
	if err != nil {
		t.Fatalf("--no-verify failed: %v\n%s", err, skipped)
	}
}

func TestInstallLeavesAnExecutableHook(t *testing.T) {
	root := newBareRepository(t)
	runInstall(t, root)

	path := filepath.Join(root, filepath.FromSlash(hookFile))
	text := utilities.ReadFile(t, path)
	if !strings.HasPrefix(text, "#!/bin/sh\n") ||
		!strings.Contains(text, "beacon snares pre-commit || exit $?") {
		t.Errorf("hook = %q", text)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Errorf("mode = %v", info.Mode())
	}
}

func TestInstallTwiceKeepsTheHookAndSkipLeavesItOut(t *testing.T) {
	root := newBareRepository(t)
	runInstall(t, root)
	again := runInstall(t, root)

	if len(again.Changes) != 0 {
		t.Errorf("second install changed %+v", again.Changes)
	}

	other := newBareRepository(t)
	runInstall(t, other, "--skip", "precommit")
	if present(other, hookFile) {
		t.Error("--skip precommit still wrote the hook")
	}
}

func TestUninstallRestoresAnExistingHook(t *testing.T) {
	root := newBareRepository(t)
	original := "#!/bin/sh\necho hello\nexit 0\n"
	utilities.WriteFile(t, root, hookFile, original)

	runInstall(t, root)
	installed := utilities.ReadFile(
		t,
		filepath.Join(root, filepath.FromSlash(hookFile)),
	)
	if strings.Index(installed, "beacon snares") > strings.Index(
		installed,
		"exit 0",
	) {
		t.Errorf("the block must come before exit 0: %q", installed)
	}

	runInstall(t, root, "--uninstall")
	restored := utilities.ReadFile(
		t,
		filepath.Join(root, filepath.FromSlash(hookFile)),
	)
	if restored != original {
		t.Errorf("hook = %q", restored)
	}
}

func TestUninstallRemovesAHookBeaconCreated(t *testing.T) {
	root := newBareRepository(t)
	runInstall(t, root)

	runInstall(t, root, "--uninstall")

	if present(root, hookFile) {
		t.Error("the hook is still there")
	}
}

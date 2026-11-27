package integration_test

import (
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

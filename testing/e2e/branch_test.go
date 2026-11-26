package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

const (
	lockoutFrom = "After 5 failed"
	mobileFrom  = "The sign-in screen looks right on mobile."
	tabletText  = "The sign-in screen looks right on tablets."
)

func commitAll(t *testing.T, root, message string) {
	t.Helper()

	utils.Git(t, root, "add", "-A")
	utils.Git(t, root, "commit", "-q", "-m", message)
}

func pendingIDs(t *testing.T, root string) []string {
	t.Helper()

	result := utils.Beacon(t, root, "diff", "--json")

	var ids []string
	for _, entry := range decodeDiff(t, result.Stdout).Pending {
		ids = append(ids, entry.ID)
	}

	return ids
}

func entryIDs(t *testing.T, root string) []string {
	t.Helper()

	result := utils.Beacon(t, root, "diff", "--json")

	var ids []string
	for _, level := range decodeDiff(t, result.Stdout).Levels {
		for _, entry := range level.Changes {
			ids = append(ids, entry.ID)
		}
	}

	return ids
}

func TestBranchDiffShowsOnlyThePromiseTheBranchChanged(t *testing.T) {
	root := utils.NewRepository(t)
	utils.Git(t, root, "switch", "-q", "-c", "feat/lockout")
	editByHand(t, root, lockoutFrom, "After 10 failed")
	commitAll(t, root, "weaken the lockout on the branch")
	utils.Git(t, root, "switch", "-q", "main")
	editByHand(t, root, mobileFrom, tabletText)
	commitAll(t, root, "main changes another promise")
	utils.Git(t, root, "switch", "-q", "feat/lockout")

	ids := entryIDs(t, root)

	if len(ids) != 1 || ids[0] != "bc-c3d4" {
		t.Fatalf("the branch must only see its own change: %v", ids)
	}
}

func TestRebasedBranchDoesNotInheritMainsApprovedChange(t *testing.T) {
	root := utils.NewRepository(t)
	utils.Git(t, root, "switch", "-q", "-c", "feat/mobile")
	editByHand(t, root, mobileFrom, tabletText)
	commitAll(t, root, "change the mobile promise on the branch")
	utils.Git(t, root, "switch", "-q", "main")
	weakened := weakenedLockout(t, root)
	approved := record(t, root, sampleAgent, weakened, "--approve")
	commitAll(t, root, "main weakens the lockout with a yes")
	utils.Git(t, root, "switch", "-q", "feat/mobile")

	rebased, err := utils.GitTry(t, root, "rebase", "main")

	if approved.Code != 0 || err != nil {
		t.Fatalf("approved = %+v, rebase = %s", approved, rebased)
	}

	pending := pendingIDs(t, root)
	if len(pending) != 1 || pending[0] != "bc-c4a1" {
		t.Fatalf("only the branch's own change is pending: %v", pending)
	}
	for _, id := range entryIDs(t, root) {
		if id == "bc-c3d4" {
			t.Fatal("main's change must not show up in the branch")
		}
	}
}

func conflictedRepository(t *testing.T) (string, string) {
	t.Helper()

	root := utils.NewRepository(t)
	utils.Git(t, root, "switch", "-q", "-c", "feat/lockout")
	editByHand(t, root, lockoutFrom, "After 10 failed")
	commitAll(t, root, "one side")
	utils.Git(t, root, "switch", "-q", "main")
	editByHand(t, root, lockoutFrom, "After 20 failed")
	commitAll(t, root, "the other side")

	merged, err := utils.GitTry(t, root, "merge", "feat/lockout")
	if err == nil {
		t.Fatalf("both sides changed the same promise: %s", merged)
	}

	return root, merged
}

func TestDoctorReportsAGitConflictOnTheSamePromise(t *testing.T) {
	root, merged := conflictedRepository(t)

	result := utils.Beacon(t, root, "doctor", "--json")

	envelope := parseRead(t, result.Stdout)
	if !strings.Contains(merged, "CONFLICT") {
		t.Fatalf("git must flag the file: %s", merged)
	}
	if result.Code != 1 || envelope.Error == nil {
		t.Fatalf("doctor = %+v", result)
	}

	finding := envelope.Error.Details.Findings[0]
	if finding.Code != "record_conflict" || finding.File != capabilityPath {
		t.Fatalf("finding = %+v", finding)
	}
	if !hintContains(envelope.Hint, "resolve the git conflict") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestDoctorPassesOnceTheConflictIsResolved(t *testing.T) {
	root, _ := conflictedRepository(t)
	utils.Git(t, root, "checkout", "--ours", capabilityPath)
	commitAll(t, root, "resolve the conflict")

	result := utils.Beacon(t, root, "doctor", "--json")

	if result.Code != 0 {
		t.Fatalf("doctor = %+v", result)
	}
}

func closeTaskOnMain(t *testing.T, root string) {
	t.Helper()

	closed := closeWithReason(t, root, claimedTask)
	if closed.Code != 0 {
		t.Fatalf("close = %+v", closed)
	}

	commitAll(t, root, "main closes the task")
}

func TestOldBranchCannotClaimATaskMainAlreadyClosed(t *testing.T) {
	root := utils.NewRepository(t)
	utils.Git(t, root, "switch", "-q", "-c", "feat/old")
	utils.Git(t, root, "switch", "-q", "main")
	closeTaskOnMain(t, root)
	utils.Git(t, root, "switch", "-q", "feat/old")
	utils.Git(t, root, "merge", "-q", "main")

	result := claim(t, root)

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "claim_task_archived" {
		t.Fatalf("claim = %+v", result)
	}
	if !hintContains(envelope.Hint, "beacon get "+claimedTask) {
		t.Fatalf("hint = %v", envelope.Hint)
	}

	shown := utils.Beacon(t, root, "get", claimedTask, "--json")
	if !strings.Contains(shown.Stdout, `"archived":true`) {
		t.Fatalf("the closed task must be readable: %s", shown.Stdout)
	}
}

func TestOldBranchThatClaimedBeforeMergingGetsAVisibleConflict(t *testing.T) {
	root := utils.NewRepository(t)
	utils.Git(t, root, "switch", "-q", "-c", "feat/old")
	if claimed := claim(t, root); claimed.Code != 0 {
		t.Fatalf("an old branch still sees the task open: %+v", claimed)
	}
	commitAll(t, root, "claim on the old branch")
	utils.Git(t, root, "switch", "-q", "main")
	closeTaskOnMain(t, root)
	utils.Git(t, root, "switch", "-q", "feat/old")

	merged, err := utils.GitTry(t, root, "merge", "main")

	if err == nil || !strings.Contains(merged, "CONFLICT") {
		t.Fatalf("closing and claiming the same task must conflict: %s", merged)
	}
	marker := filepath.Join(root, ".git", "MERGE_HEAD")
	if _, statErr := os.Stat(marker); statErr != nil {
		t.Fatal("the merge must be left open for a person to resolve")
	}
}

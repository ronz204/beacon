package e2e_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

const claimedTask = "bc-91cd.1.1"

func repositoryWithReadyTask(t *testing.T) string {
	t.Helper()

	root := utils.NewRepository(t)
	closed := closeWithReason(t, root, "bc-e5f6")
	if closed.Code != 0 {
		t.Fatalf("close = %+v", closed)
	}

	utils.Git(t, root, "add", "-A")
	utils.Git(t, root, "commit", "-q", "-m", "answer the question")

	return root
}

func addWorktree(t *testing.T, root, branch string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "worktree")
	utils.Git(t, root, "worktree", "add", "-q", "-b", branch, path)

	return path
}

func claim(t *testing.T, directory string, flags ...string) utils.Result {
	t.Helper()

	args := append([]string{"claim", claimedTask, "--json"}, flags...)

	return utils.BeaconWith(
		t,
		directory,
		utils.Options{Actor: sampleAgent},
		args...,
	)
}

func readyTasks(t *testing.T, directory string) []string {
	t.Helper()

	result := utils.Beacon(t, directory, "ready", "--json")

	var data struct {
		Tasks []struct {
			ID string `json:"id"`
		} `json:"tasks"`
	}
	decodeData(t, result.Stdout, &data)

	ids := make([]string, 0, len(data.Tasks))
	for _, task := range data.Tasks {
		ids = append(ids, task.ID)
	}

	return ids
}

func TestClaimWritesTheLocalClaimAndTheRecord(t *testing.T) {
	root := repositoryWithReadyTask(t)

	result := claim(t, root)

	stored := utils.Beacon(t, root, "get", claimedTask, "--json")
	var data struct {
		Record map[string]any `json:"record"`
	}
	decodeData(t, stored.Stdout, &data)

	if result.Code != 0 {
		t.Fatalf("claim = %+v", result)
	}
	if data.Record["status"] != "claimed" || data.Record["branch"] != "main" {
		t.Fatalf("record = %v", data.Record)
	}
	if data.Record["owner"] != sampleAgent {
		t.Fatalf("record = %v", data.Record)
	}
	if !slicesContain(readyTasks(t, root), claimedTask) {
		t.Fatal("the owner branch must still see its task as ready")
	}
}

func slicesContain(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}

	return false
}

func TestSecondWorktreeCannotClaimAndDoesNotSeeTheTask(t *testing.T) {
	root := repositoryWithReadyTask(t)
	other := addWorktree(t, root, "feat/other-work")
	if first := claim(t, root); first.Code != 0 {
		t.Fatalf("first = %+v", first)
	}

	second := claim(t, other)

	envelope := parseRead(t, second.Stdout)
	if second.Code != 2 || envelope.Error.Code != "claim_taken" {
		t.Fatalf("second = %+v", second)
	}
	if !hintContains(envelope.Hint, "main") {
		t.Fatalf("the hint must name the holder: %v", envelope.Hint)
	}
	if slicesContain(readyTasks(t, other), claimedTask) {
		t.Fatal("ready offers a task another worktree holds")
	}
}

func TestForceTakesAClaimedTask(t *testing.T) {
	root := repositoryWithReadyTask(t)
	other := addWorktree(t, root, "feat/other-work")
	claim(t, root)

	forced := claim(t, other, "--force")

	var data struct {
		Replaced struct {
			Branch string `json:"branch"`
		} `json:"replaced"`
	}
	decodeData(t, forced.Stdout, &data)

	if forced.Code != 0 || data.Replaced.Branch != "main" {
		t.Fatalf("forced = %+v", forced)
	}
	if slicesContain(readyTasks(t, root), claimedTask) {
		t.Fatal("the first branch must lose the task after a forced claim")
	}
	if !slicesContain(readyTasks(t, other), claimedTask) {
		t.Fatal("the second branch must now see the task")
	}
}

func TestClaimIsRepeatableOnTheSameBranch(t *testing.T) {
	root := repositoryWithReadyTask(t)

	first := claim(t, root)
	second := claim(t, root)

	if first.Code != 0 || second.Code != 0 {
		t.Fatalf("first = %+v, second = %+v", first, second)
	}
}

func TestClaimRejectsWhatIsNotAClaimableTask(t *testing.T) {
	root := repositoryWithReadyTask(t)

	notTask := utils.Beacon(t, root, "claim", "bc-7f3a", "--json")
	unknown := utils.Beacon(t, root, "claim", "bc-ffff", "--json")

	var first, second struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(notTask.Stdout), &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(unknown.Stdout), &second); err != nil {
		t.Fatal(err)
	}

	if notTask.Code != 1 || first.Error.Code != "claim_not_task" {
		t.Fatalf("notTask = %+v", notTask)
	}
	if unknown.Code != 1 || second.Error.Code != "record_missing" {
		t.Fatalf("unknown = %+v", unknown)
	}
}

func TestClaimNeedsABranch(t *testing.T) {
	root := repositoryWithReadyTask(t)
	utils.Git(t, root, "checkout", "-q", "--detach")

	result := claim(t, root)

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "claim_branch_detached" {
		t.Fatalf("result = %+v", result)
	}
}

func TestDoctorFreesTheClaimOfARemovedWorktree(t *testing.T) {
	root := repositoryWithReadyTask(t)
	other := addWorktree(t, root, "feat/other-work")
	if taken := claim(t, other); taken.Code != 0 {
		t.Fatalf("claim = %+v", taken)
	}
	if blocked := claim(t, root); blocked.Code != 2 {
		t.Fatalf("blocked = %+v", blocked)
	}

	utils.Git(t, root, "worktree", "remove", "--force", other)
	utils.Git(t, root, "branch", "-D", "feat/other-work")
	doctor := utils.Beacon(t, root, "doctor", "--json")
	retry := claim(t, root)

	if doctor.Code != 0 || !strings.Contains(doctor.Stdout, "claim_removed") {
		t.Fatalf("doctor = %+v", doctor)
	}
	if retry.Code != 0 {
		t.Fatalf("retry = %+v", retry)
	}
}

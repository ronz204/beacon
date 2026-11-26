package e2e_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

type diffData struct {
	Revision string `json:"revision"`
	Levels   []struct {
		Level   int         `json:"level"`
		Changes []diffEntry `json:"changes"`
	} `json:"levels"`
	Pending []diffEntry `json:"pending"`
}

type diffEntry struct {
	Target   string `json:"target"`
	ID       string `json:"id"`
	Level    int    `json:"level"`
	Reason   string `json:"reason"`
	NeedsYes bool   `json:"needsYes"`
	Approved bool   `json:"approved"`
}

func decodeDiff(t *testing.T, text string) diffData {
	t.Helper()

	var data diffData
	decodeData(t, text, &data)

	return data
}

func editByHand(t *testing.T, root, from, to string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	utils.WriteFile(t, root, capabilityPath, strings.Replace(
		utils.ReadFile(t, path),
		from,
		to,
		1,
	))
}

func entryFor(data diffData, id string) (diffEntry, bool) {
	for _, level := range data.Levels {
		for _, entry := range level.Changes {
			if entry.ID == id {
				return entry, true
			}
		}
	}

	return diffEntry{}, false
}

func TestDiffOfACleanRepositoryHasNothing(t *testing.T) {
	root := utils.NewRepository(t)

	result := utils.Beacon(t, root, "diff", "--json")

	data := decodeDiff(t, result.Stdout)
	if result.Code != 0 || len(data.Levels) != 0 || len(data.Pending) != 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestDiffShowsAHandEditedPromiseAsPending(t *testing.T) {
	root := utils.NewRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")

	result := utils.Beacon(t, root, "diff", "--json")

	data := decodeDiff(t, result.Stdout)
	envelope := parseRead(t, result.Stdout)
	entry, found := entryFor(data, "bc-c3d4")
	if result.Code != 0 || !found || entry.Level != 3 || !entry.NeedsYes {
		t.Fatalf("result = %+v", result)
	}
	if entry.Approved || len(data.Pending) != 1 {
		t.Fatalf("data = %+v", data)
	}
	if len(envelope.Signals) != 1 || envelope.Signals[0].Kind != "pending_yes" {
		t.Fatalf("signals = %+v", envelope.Signals)
	}
	if !hintContains(envelope.Hint, "--approve") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestDiffTreatsATypoAsMaintenance(t *testing.T) {
	root := utils.NewRepository(t)
	editByHand(t, root, "locks for 15 minutes.", "locks for 15 minutes!!")

	data := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)

	entry, found := entryFor(data, "bc-c3d4")
	if found && entry.NeedsYes {
		t.Fatalf("entry = %+v", entry)
	}
	if len(data.Pending) != 0 {
		t.Fatalf("pending = %+v", data.Pending)
	}
}

func TestRecordApproveResolvesAHandEditedPromise(t *testing.T) {
	root := utils.NewRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")
	payload := jsonOf(t, storedPayload(t, root, "bc-7f3a"))

	silent := record(t, root, sampleAgent, payload)
	again := utils.Beacon(t, root, "diff", "--json")
	stillPending := decodeDiff(t, again.Stdout)
	approved := record(t, root, sampleAgent, payload, "--approve")
	resolved := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)

	if silent.Code != 0 || len(stillPending.Pending) != 1 {
		t.Fatalf("silent = %+v, pending = %+v", silent, stillPending)
	}
	if approved.Code != 0 || len(resolved.Pending) != 0 {
		t.Fatalf("approved = %+v, resolved = %+v", approved, resolved)
	}

	entry, _ := entryFor(resolved, "bc-c3d4")
	if !entry.Approved {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestDiffSeesTheStampOfAClosedPromise(t *testing.T) {
	root := utils.NewRepository(t)
	closed := closeWithReason(t, root, "bc-c3d4", "--approve")

	data := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)

	entry, found := entryFor(data, "bc-c3d4")
	if closed.Code != 0 || !found || !entry.Approved {
		t.Fatalf("entry = %+v, found = %v", entry, found)
	}
	if len(data.Pending) != 0 {
		t.Fatalf("pending = %+v", data.Pending)
	}
}

func TestDiffSinceListsCommittedChanges(t *testing.T) {
	root := utils.NewRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")
	utils.Git(t, root, "commit", "-q", "-am", "weaken the lockout")

	current := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)
	since := decodeDiff(
		t,
		utils.Beacon(t, root, "diff", "--since", "HEAD~1", "--json").Stdout,
	)

	if len(current.Pending) != 0 {
		t.Fatalf("current = %+v", current)
	}
	if len(since.Pending) != 1 || since.Pending[0].ID != "bc-c3d4" {
		t.Fatalf("since = %+v", since)
	}
}

func TestBaselineCaptureNeedsAYesUntilARunnerPromiseIsApproved(t *testing.T) {
	root := utils.NewRepository(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	utils.WriteFile(t, root, capabilityPath, utils.ReadFile(t, path)+
		"  - id: bc-d7a8\n"+
		"    text: The sign-in screen has no visual regressions.\n"+
		"    check: {runner: visual, target: home}\n"+
		"    status: proposed\n")
	utils.Git(t, root, "commit", "-q", "-am", "add a visual promise")
	utils.WriteFile(t, root, "tests/auth/login-snapshots/home.png", "png")

	pending := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)
	payload := jsonOf(t, storedPayload(t, root, "bc-7f3a"))
	approved := record(t, root, sampleAgent, payload, "--approve")
	resolved := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)

	if len(pending.Pending) != 1 || pending.Pending[0].Target != "baseline" {
		t.Fatalf("pending = %+v", pending.Pending)
	}
	if approved.Code != 0 || len(resolved.Pending) != 0 {
		t.Fatalf("approved = %+v, resolved = %+v", approved, resolved)
	}
}

func TestSecondApprovalOfTheSameDayStillResolvesThePending(t *testing.T) {
	root := utils.NewRepository(t)
	first := record(t, root, sampleAgent, weakenedLockout(t, root), "--approve")
	utils.Git(t, root, "commit", "-q", "-am", "approve the first weakening")

	payload := storedPayload(t, root, "bc-7f3a")
	promise := promiseOf(payload, "bc-c3d4")
	promise["text"] = "After 20 failed attempts the account locks."
	second := record(t, root, sampleAgent, jsonOf(t, payload), "--approve")

	data := decodeDiff(t, utils.Beacon(t, root, "diff", "--json").Stdout)
	text := utils.ReadFile(t, filepath.Join(root, capabilityPath))

	if first.Code != 0 || second.Code != 0 {
		t.Fatalf("first = %+v, second = %+v", first, second)
	}
	if !strings.Contains(text, ", n: 2}") {
		t.Fatalf("the repeated stamp must carry a count:\n%s", text)
	}
	if len(data.Pending) != 0 {
		t.Fatalf("pending = %+v", data.Pending)
	}

	entry, found := entryFor(data, "bc-c3d4")
	if !found || !entry.Approved {
		t.Fatalf("entry = %+v, found = %v", entry, found)
	}
}

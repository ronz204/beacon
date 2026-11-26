package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

const (
	questionPath = ".beacon/records/questions/bc-e5f6-attempt-limit.yaml"
	guidePath    = ".beacon/records/guides/bc-0a9b-architecture-overview.md"
	closeReason  = "Answered in the sign-in design."
)

func closeRecord(
	t *testing.T,
	root, actor, id string,
	flags ...string,
) utils.Result {
	t.Helper()

	args := append([]string{"close", id, "--json"}, flags...)

	return utils.BeaconWith(t, root, utils.Options{Actor: actor}, args...)
}

func closeWithReason(
	t *testing.T,
	root, id string,
	flags ...string,
) utils.Result {
	t.Helper()

	flags = append([]string{"--reason", closeReason}, flags...)

	return closeRecord(t, root, sampleAgent, id, flags...)
}

func TestCloseArchivesAQuestionWithoutAskingForAYes(t *testing.T) {
	root := utils.NewRepository(t)

	result := closeWithReason(t, root, "bc-e5f6")

	var stored struct {
		Archived bool `json:"archived"`
	}
	got := utils.Beacon(t, root, "get", "bc-e5f6", "--json")
	decodeData(t, got.Stdout, &stored)

	if result.Code != 0 || !stored.Archived {
		t.Fatalf("close = %+v, get = %s", result, got.Stdout)
	}
	if _, err := os.Stat(filepath.Join(root, questionPath)); err == nil {
		t.Fatal("the question file is still in records")
	}
}

func TestCloseNeedsAReason(t *testing.T) {
	root := utils.NewRepository(t)

	result := closeRecord(t, root, sampleAgent, "bc-e5f6")

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "reason_required" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "--reason") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestCloseOfAnActivePromiseAsksThenStampsTheArchiveLine(t *testing.T) {
	root := utils.NewRepository(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	before := utils.ReadFile(t, path)

	blocked := closeWithReason(t, root, "bc-c3d4")

	var asked struct {
		Details struct {
			Changes []struct {
				ID     string `json:"id"`
				Reason string `json:"reason"`
				Before string `json:"before"`
			} `json:"changes"`
		} `json:"details"`
	}
	decodeError(t, blocked.Stdout, &asked)

	if blocked.Code != 2 || utils.ReadFile(t, path) != before {
		t.Fatalf("blocked = %+v", blocked)
	}
	change := asked.Details.Changes[0]
	if change.ID != "bc-c3d4" || change.Reason != "promise_closed" {
		t.Fatalf("change = %+v", change)
	}
	if change.Before != lockoutText {
		t.Fatalf("change = %+v", change)
	}

	approved := closeWithReason(t, root, "bc-c3d4", "--approve")

	stored := utils.Beacon(t, root, "get", "bc-c3d4", "--json")
	if approved.Code != 0 || !strings.Contains(stored.Stdout, "human:jane") {
		t.Fatalf("approved = %+v, stored = %s", approved, stored.Stdout)
	}
	if strings.Contains(utils.ReadFile(t, path), "bc-c3d4") {
		t.Fatal("the promise is still in its capability")
	}
}

func TestCloseRefusesAGuideAndHintsAtDeletingIt(t *testing.T) {
	root := utils.NewRepository(t)

	result := closeWithReason(t, root, "bc-0a9b")

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "guide_not_archivable" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "delete the guide") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
	if _, err := os.Stat(filepath.Join(root, guidePath)); err != nil {
		t.Fatal("the guide must stay")
	}
}

func TestCloseRefusesATaskWithOpenSubtasks(t *testing.T) {
	root := utils.NewRepository(t)

	result := closeWithReason(t, root, "bc-91cd.1")

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "task_has_open_children" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !strings.Contains(result.Stdout, "bc-91cd.1.1") {
		t.Fatalf("the subtasks are not listed: %s", result.Stdout)
	}
}

func TestCloseOfAnUnknownRecordHintsAtFind(t *testing.T) {
	root := utils.NewRepository(t)

	result := closeWithReason(t, root, "bc-ffff")

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "record_missing" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestCloseOfACapabilityStampsOneLineThatCoversItsPromises(t *testing.T) {
	root := utils.NewRepository(t)

	blocked := closeWithReason(t, root, "bc-7f3a")
	approved := closeWithReason(t, root, "bc-7f3a", "--approve")
	doctor := utils.Beacon(t, root, "doctor", "--json")

	var asked struct {
		Details struct {
			Changes []struct {
				ID string `json:"id"`
			} `json:"changes"`
		} `json:"details"`
	}
	decodeError(t, blocked.Stdout, &asked)

	if blocked.Code != 2 || len(asked.Details.Changes) != 2 {
		t.Fatalf("blocked = %+v", blocked)
	}
	if approved.Code != 0 {
		t.Fatalf("approved = %+v", approved)
	}
	if doctor.Code != 0 {
		t.Fatalf("the graph broke after the close: %s", doctor.Stdout)
	}
}

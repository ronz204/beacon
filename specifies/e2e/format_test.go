package e2e_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/specifies/utils"
)

const (
	messyGoalPath = ".beacon/records/goals/bc-aaaa-sample.yaml"
	messyGoal     = "status: active\nsuccess: Sample\nstatement: Sample\n" +
		"author: 'human:jane'\ncreated: 2026-11-22\nid: bc-aaaa\n"
	cleanGoal = "id: bc-aaaa\ncreated: 2026-11-22\nauthor: \"human:jane\"\n" +
		"statement: Sample\nsuccess: Sample\nstatus: active\n"
)

func TestFormatRewritesAndIsIdempotent(t *testing.T) {
	root := utils.NewWorkspace(t)
	path := utils.WriteFile(t, root, messyGoalPath, messyGoal)

	first := utils.Beacon(t, root, "fmt", "--json")
	parsed := parseEnvelope(t, first.Stdout)

	if first.Code != 0 || !parsed.OK {
		t.Fatalf("result = %+v", first)
	}
	if !strings.Contains(string(parsed.Data), messyGoalPath) {
		t.Fatalf("data = %s", parsed.Data)
	}
	if got := utils.ReadFile(t, path); got != cleanGoal {
		t.Fatalf("got:\n%s", got)
	}

	second := utils.Beacon(t, root, "fmt", "--json")
	parsed = parseEnvelope(t, second.Stdout)

	if second.Code != 0 {
		t.Fatalf("result = %+v", second)
	}
	if strings.Contains(string(parsed.Data), messyGoalPath) {
		t.Fatalf("second run changed files: %s", parsed.Data)
	}
}

func TestDoctorHintLeadsToFormat(t *testing.T) {
	root := utils.NewWorkspace(t)
	utils.WriteFile(t, root, messyGoalPath, messyGoal)

	before := utils.Beacon(t, root, "doctor", "--json")
	parsed := parseEnvelope(t, before.Stdout)

	if before.Code != 1 || parsed.Hint == nil {
		t.Fatalf("result = %+v", before)
	}
	if !strings.Contains(*parsed.Hint, "beacon fmt") {
		t.Fatalf("hint = %q", *parsed.Hint)
	}

	if result := utils.Beacon(t, root, "fmt"); result.Code != 0 {
		t.Fatalf("fmt = %+v", result)
	}

	after := utils.Beacon(t, root, "doctor")
	if after.Code != 0 {
		t.Fatalf("doctor = %+v", after)
	}
}

func TestFormatSkipsInvalidFilesAndReportsThem(t *testing.T) {
	root := utils.NewWorkspace(t)
	messy := utils.WriteFile(t, root, messyGoalPath, messyGoal)
	breakGoal(t, root)
	broken := filepath.Join(root, filepath.FromSlash(brokenGoal))
	before := utils.ReadFile(t, broken)

	result := utils.Beacon(t, root, "fmt", "--json")
	parsed := parseEnvelope(t, result.Stdout)

	if result.Code != 1 || parsed.Error == nil {
		t.Fatalf("result = %+v", result)
	}
	if got := utils.ReadFile(t, broken); got != before {
		t.Fatalf("invalid file changed:\n%s", got)
	}
	if got := utils.ReadFile(t, messy); got != cleanGoal {
		t.Fatalf("valid file was not formatted:\n%s", got)
	}

	changed := parsed.Error.Details.Changed
	if len(changed) != 1 || changed[0] != messyGoalPath {
		t.Fatalf("changed = %v", changed)
	}
}

func TestFormatTextReportsChanges(t *testing.T) {
	root := utils.NewWorkspace(t)
	utils.WriteFile(t, root, messyGoalPath, messyGoal)

	result := utils.Beacon(t, root, "fmt")

	if result.Code != 0 {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(result.Stdout, messyGoalPath) {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

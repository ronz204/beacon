package deltas_test

import (
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

const lockout = "After 5 failed attempts the account locks for 15 minutes"

func TestNormalizeTextIgnoresCaseAndPunctuation(t *testing.T) {
	got := deltas.NormalizeText("  Locks,  after 5 tries!! ")

	if got != "locks after 5 tries" {
		t.Fatalf("normalized = %q", got)
	}
}

func TestNormalizeTextKeepsNumbersApart(t *testing.T) {
	if deltas.NormalizeText("1.5") == deltas.NormalizeText("15") {
		t.Fatal("1.5 and 15 must stay different")
	}
}

type levelCase struct {
	name     string
	kind     string
	before   *deltas.ChangedRecordInput
	after    *deltas.ChangedRecordInput
	level    int
	needsYes bool
}

func goalOf(fields map[string]string) *deltas.ChangedRecordInput {
	return recordOf("goal", "bc-0001", fields)
}

func configOf(fields map[string]string) *deltas.ChangedRecordInput {
	return recordOf("config", "", fields)
}

func plain(kind, id string) *deltas.ChangedRecordInput {
	return recordOf(kind, id, nil)
}

func with(
	record *deltas.ChangedRecordInput,
	key, value string,
) *deltas.ChangedRecordInput {
	cloned := *record
	cloned.Fields = map[string]string{key: value}
	for existing, text := range record.Fields {
		if _, set := cloned.Fields[existing]; !set {
			cloned.Fields[existing] = text
		}
	}

	return &cloned
}

func TestClassifyChangeFollowsTheLevelTable(t *testing.T) {
	active := promise("bc-c3d4", lockout, "active")
	retyped := promise("bc-c3d4", lockout+".", "active")
	tenTries := promise("bc-c3d4", "After 10 failed attempts", "active")
	recheck := active
	recheck.Check = `{"runner":"test-go","target":"b"}`
	soft := promise("bc-c3d4", lockout, "proposed")
	edited := promise("bc-c3d4", "Other", "proposed")
	more := promise("bc-d5e6", "Locks for admins", "proposed")
	moreOn := promise("bc-d5e6", "Locks for admins", "active")
	goal := goalOf(map[string]string{"statement": "Pay fast", "success": "9"})
	capability := caps(active)
	renamed := with(capability, "name", "x")
	narrowed := with(capability, "paths", "x")
	retitled := with(goal, "statement", "Pay more")
	goalTypo := with(goal, "statement", "pay fast!")
	dropped := with(goal, "status", "dropped")
	nongoals := with(goal, "nongoals", "x")
	decision := plain("decision", "bc-a1b2")
	rejected := with(decision, "rejected", "x")
	question := plain("question", "bc-e5f6")
	answered := with(question, "text", "x")
	task := plain("task", "bc-91cd.1")
	claimed := with(task, "owner", "x")
	work := plain("work", "bc-91cd")
	runnersA := configOf(map[string]string{"runners": "a"})
	runnersB := configOf(map[string]string{"runners": "b"})
	schemaA := configOf(map[string]string{"schema": "1"})
	schemaB := configOf(map[string]string{"schema": "2"})
	limitsA := configOf(map[string]string{"limits": "1"})
	limitsB := configOf(map[string]string{"limits": "2"})
	packsA := configOf(map[string]string{"packs": "a"})
	packsB := configOf(map[string]string{"packs": "b"})
	agentsA := configOf(map[string]string{"agents": "a"})
	agentsB := configOf(map[string]string{"agents": "b"})

	tests := []levelCase{
		{"active text", "modified", caps(active), caps(tenTries), 3, true},
		{"active check", "modified", caps(active), caps(recheck), 3, true},
		{"active typo", "modified", caps(active), caps(retyped), 1, false},
		{"active soft", "modified", caps(active), caps(soft), 3, true},
		{"active closed", "modified", caps(active, more), caps(more), 3, true},
		{"activated", "modified", caps(soft), caps(active), 2, true},
		{"soft edited", "modified", caps(soft), caps(edited), 1, false},
		{"soft gone", "modified", caps(soft, more), caps(more), 1, false},
		{"new soft", "modified", caps(active), caps(active, more), 2, false},
		{"new active", "modified", caps(active), caps(active, moreOn), 2, true},
		{"new capability", "added", nil, caps(soft), 2, false},
		{"capability deleted", "deleted", caps(active), nil, 3, true},
		{"renamed", "modified", capability, renamed, 1, false},
		{"narrowed", "modified", capability, narrowed, 1, false},
		{"new goal", "added", nil, goal, 3, true},
		{"goal statement", "modified", goal, retitled, 3, true},
		{"goal typo", "modified", goal, goalTypo, 1, false},
		{"goal status", "modified", goal, dropped, 3, true},
		{"goal nongoals", "modified", goal, nongoals, 1, false},
		{"new decision", "added", nil, decision, 2, false},
		{"rejected", "modified", decision, rejected, 1, false},
		{"question", "modified", question, answered, 1, false},
		{"new guide", "added", nil, plain("guide", "bc-b7c8"), 1, false},
		{"task edited", "modified", task, claimed, 1, false},
		{"work deleted", "deleted", work, nil, 1, false},
		{"runners", "modified", runnersA, runnersB, 3, true},
		{"schema", "modified", schemaA, schemaB, 3, true},
		{"limits", "modified", limitsA, limitsB, 1, false},
		{"packs", "modified", packsA, packsB, 3, true},
		{"agents", "modified", agentsA, agentsB, 1, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			change := changeOf(test.kind, test.before, test.after)
			level, needsYes := levelOf(t, change)

			if level != test.level || needsYes != test.needsYes {
				t.Fatalf("got %d/%v", level, needsYes)
			}
		})
	}
}

func TestClassifyChangeFlagsBaselineFiles(t *testing.T) {
	change, err := deltas.ParseStateChange(deltas.StateChangeInput{
		Path:     "e2e/shots/login.png",
		Kind:     "modified",
		Baseline: true,
		Runners:  []string{"visual"},
	})
	if err != nil {
		t.Fatal(err)
	}

	level, needsYes := levelOf(t, change)

	if level != 3 || !needsYes {
		t.Fatalf("got %d/%v", level, needsYes)
	}
}

func TestClassifyElementsNamesTheWeakenedPromise(t *testing.T) {
	active := promise("bc-c3d4", lockout, "active")
	tenTries := promise("bc-c3d4", "After 10 failed attempts", "active")
	change := changeOf(
		"modified",
		caps(active),
		caps(tenTries),
	)

	var asking []string
	for _, finding := range deltas.ClassifyElements(change) {
		if finding.NeedsYes {
			asking = append(asking, finding.Target.Value())
		}
	}

	if len(asking) != 1 || asking[0] != "bc-c3d4" {
		t.Fatalf("asking = %v", asking)
	}
}

func TestParseStateChangeRejectsUnknownKind(t *testing.T) {
	_, err := deltas.ParseStateChange(deltas.StateChangeInput{
		Path:   "x",
		Kind:   "renamed",
		Before: recordOf("task", "bc-91cd.1", nil),
	})

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestClassifyElementsSkipsUntouchedPromises(t *testing.T) {
	active := promise("bc-c3d4", lockout, "active")
	other := promise("bc-d5e6", "Locks for admins", "active")
	renamed := with(caps(active, other), "name", "Sign in")

	change := changeOf("modified", caps(active, other), renamed)

	findings := deltas.ClassifyElements(change)
	if len(findings) != 1 || findings[0].Target.Value() != "bc-7f3a" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestClassifyElementsQuotesWhatChangedInAPromise(t *testing.T) {
	active := promise("bc-c3d4", lockout, "active")
	tenTries := promise("bc-c3d4", "After 10 failed attempts", "active")

	findings := deltas.ClassifyElements(changeOf(
		"modified",
		caps(active),
		caps(tenTries),
	))

	got := findings[0]
	if got.Reason != deltas.ReasonPromiseText || got.Before != lockout {
		t.Fatalf("finding = %+v", got)
	}
	if got.After != "After 10 failed attempts" {
		t.Fatalf("finding = %+v", got)
	}
}

func TestAddedPromisesListsOnlyTheNewOnes(t *testing.T) {
	active := promise("bc-c3d4", lockout, "active")
	fresh := promise("bc-d5e6", "Locks for admins", "proposed")

	cases := []struct {
		name   string
		change deltas.StateChange
		want   []string
	}{
		{
			"one new",
			changeOf("modified", caps(active), caps(active, fresh)),
			[]string{"bc-d5e6"},
		},
		{
			"none new",
			changeOf("modified", caps(active), caps(active)),
			nil,
		},
		{
			"new record",
			changeOf("added", nil, caps(active, fresh)),
			[]string{"bc-c3d4", "bc-d5e6"},
		},
		{
			"deleted",
			changeOf("deleted", caps(active), nil),
			nil,
		},
	}

	for _, c := range cases {
		got := c.change.AddedPromises()

		if !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

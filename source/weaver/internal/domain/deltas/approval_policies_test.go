package deltas_test

import (
	"errors"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

func mustActor(t *testing.T, raw string) kernel.Actor {
	t.Helper()

	actor, err := kernel.ParseActor(raw)
	if err != nil {
		t.Fatal(err)
	}

	return actor
}

func mustDate(t *testing.T, raw string) kernel.CalendarDate {
	t.Helper()

	day, err := kernel.ParseCalendarDate(raw)
	if err != nil {
		t.Fatal(err)
	}

	return day
}

func TestStampApprovalForAPersonHasNoVia(t *testing.T) {
	stamp, err := deltas.StampApproval(
		mustActor(t, "human:jane"),
		"ignored",
		mustDate(t, "2026-11-26"),
	)

	if err != nil {
		t.Fatal(err)
	}
	if stamp.By != "human:jane" || stamp.Via != "" {
		t.Fatalf("stamp = %+v", stamp)
	}
}

func TestStampApprovalForAnAgentCarriesThePerson(t *testing.T) {
	stamp, err := deltas.StampApproval(
		mustActor(t, "agent:sample-agent"),
		"jane",
		mustDate(t, "2026-11-26"),
	)

	if err != nil {
		t.Fatal(err)
	}

	want := kernel.ApprovalInput{
		By:  "human:jane",
		At:  "2026-11-26",
		Via: "agent:sample-agent",
	}
	if stamp != want {
		t.Fatalf("stamp = %+v", stamp)
	}
}

func TestStampApprovalForAnAgentNeedsAName(t *testing.T) {
	_, err := deltas.StampApproval(
		mustActor(t, "agent:sample-agent"),
		"  ",
		mustDate(t, "2026-11-26"),
	)

	var failure deltas.ApprovalError
	if !errors.As(err, &failure) {
		t.Fatalf("err = %v", err)
	}
	if failure.Code != deltas.ApprovalIdentityMissing {
		t.Fatalf("code = %q", failure.Code)
	}
}

func approvedPromise(approved string) deltas.ChangedPromiseInput {
	input := promise("bc-c3d4", lockout, "active")
	input.Approved = approved

	return input
}

func promiseTarget(t *testing.T) deltas.ChangeTarget {
	t.Helper()

	id, err := kernel.ParseRecordID("bc-c3d4")
	if err != nil {
		t.Fatal(err)
	}

	return deltas.PromiseTarget(id)
}

func TestFreshApprovalNeedsAChangedApprovedOnThePromise(t *testing.T) {
	stamped := changeOf(
		"modified",
		caps(approvedPromise("")),
		caps(approvedPromise("jane@2026-11-26")),
	)
	untouched := changeOf(
		"modified",
		caps(approvedPromise("jane@2026-11-25")),
		caps(approvedPromise("jane@2026-11-25")),
	)
	target := promiseTarget(t)

	fresh := deltas.IsFreshlyApproved(target, []deltas.StateChange{stamped})
	stale := deltas.IsFreshlyApproved(target, []deltas.StateChange{untouched})

	if !fresh || stale {
		t.Fatalf("fresh = %v, stale = %v", fresh, stale)
	}
}

func TestFreshApprovalForConfigReadsItsApprovedKey(t *testing.T) {
	before := recordOf("config", "", map[string]string{"runners": "a"})
	after := recordOf("config", "", map[string]string{"runners": "b"})
	after.Approved = "jane@2026-11-26"
	change := changeOf("modified", before, after)

	fresh := deltas.IsFreshlyApproved(
		deltas.ConfigTarget(),
		[]deltas.StateChange{change},
	)

	if !fresh {
		t.Fatal("config approval was not seen")
	}
}

func TestFreshApprovalForAGoalReadsItsApprovedKey(t *testing.T) {
	before := recordOf("goal", "bc-0001", map[string]string{"status": "active"})
	after := recordOf("goal", "bc-0001", map[string]string{"status": "dropped"})
	after.Approved = "jane@2026-11-26"
	id, err := kernel.ParseRecordID("bc-0001")
	if err != nil {
		t.Fatal(err)
	}

	fresh := deltas.IsFreshlyApproved(
		deltas.RecordTarget(id),
		[]deltas.StateChange{changeOf("modified", before, after)},
	)

	if !fresh {
		t.Fatal("goal approval was not seen")
	}
}

func baselineChange(t *testing.T) deltas.StateChange {
	t.Helper()

	change, err := deltas.ParseStateChange(deltas.StateChangeInput{
		Path:     "e2e/shots/login.png",
		Kind:     "modified",
		Baseline: true,
		Runners:  []string{"test-go"},
	})
	if err != nil {
		t.Fatal(err)
	}

	return change
}

func TestFreshApprovalForABaselineNeedsAPromiseOfThatRunner(t *testing.T) {
	stamped := changeOf(
		"modified",
		caps(approvedPromise("")),
		caps(approvedPromise("jane@2026-11-26")),
	)
	target, err := deltas.BaselineTarget("e2e/shots/login.png")
	if err != nil {
		t.Fatal(err)
	}

	withApproval := []deltas.StateChange{baselineChange(t), stamped}
	withoutApproval := []deltas.StateChange{baselineChange(t)}

	fresh := deltas.IsFreshlyApproved(target, withApproval)
	stale := deltas.IsFreshlyApproved(target, withoutApproval)

	if !fresh || stale {
		t.Fatalf("fresh = %v, stale = %v", fresh, stale)
	}
}

func closureChange(t *testing.T, approved string) deltas.StateChange {
	t.Helper()

	change, err := deltas.ParseStateChange(deltas.StateChangeInput{
		Path:   ".beacon/archive/2026-11.jsonl",
		Kind:   "modified",
		Closed: []deltas.ClosureInput{{ID: "bc-c3d4", Approved: approved}},
	})
	if err != nil {
		t.Fatal(err)
	}

	return change
}

func TestFreshApprovalSeesTheStampOfAnArchivedClosure(t *testing.T) {
	target := promiseTarget(t)
	stamped := closureChange(t, `{"by":"human:jane","at":"2026-11-26"}`)
	bare := closureChange(t, "")

	fresh := deltas.IsFreshlyApproved(target, []deltas.StateChange{stamped})
	stale := deltas.IsFreshlyApproved(target, []deltas.StateChange{bare})

	if !fresh || stale {
		t.Fatalf("fresh = %v, stale = %v", fresh, stale)
	}
}

func TestClassifyChangeIgnoresAnArchiveSegment(t *testing.T) {
	level, needsYes := levelOf(t, closureChange(t, ""))

	if level != 1 || needsYes {
		t.Fatalf("got %d/%v", level, needsYes)
	}
}

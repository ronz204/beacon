package ledger_test

import (
	"testing"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
)

var sampleNow = time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

func sampleGoalInput() ledger.GoalInput {
	return ledger.GoalInput{
		ID:        must(ledger.ParseGoalID("bc-0001")),
		Author:    must(kernel.ParseActor("human:jane")),
		Statement: must(kernel.ParseRecordText("Checkout in two minutes")),
		Success:   must(kernel.ParseRecordText("Over 90% finish")),
		Nongoals: []kernel.RecordText{
			must(kernel.ParseRecordText("No recurring payments")),
		},
	}
}

func TestNewGoalStartsActiveAndDated(t *testing.T) {
	goal, err := ledger.NewGoal(sampleGoalInput(), sampleNow)

	if err != nil {
		t.Fatal(err)
	}
	if goal.Status() != ledger.GoalActive {
		t.Fatalf("status = %q", goal.Status())
	}
	if goal.Created().String() != "2026-11-22" {
		t.Fatalf("created = %q", goal.Created())
	}
}

func TestNewGoalRejectsMissingField(t *testing.T) {
	input := sampleGoalInput()
	input.Statement = kernel.RecordText{}

	_, err := ledger.NewGoal(input, sampleNow)

	requireCode(t, err, ledger.GoalFieldMissing)
}

func TestGoalProtectsItsNongoals(t *testing.T) {
	input := sampleGoalInput()
	goal := must(ledger.NewGoal(input, sampleNow))
	replacement := must(kernel.ParseRecordText("changed"))

	input.Nongoals[0] = replacement
	goal.Nongoals()[0] = replacement

	if goal.Nongoals()[0].String() != "No recurring payments" {
		t.Fatalf("nongoals = %v", goal.Nongoals())
	}
}

func TestRestoreGoalKeepsStateAndIdentity(t *testing.T) {
	goal := must(ledger.NewGoal(sampleGoalInput(), sampleNow))

	restored := ledger.RestoreGoal(ledger.GoalState{
		ID:     goal.ID(),
		Status: ledger.GoalDropped,
	})

	if restored.Status() != ledger.GoalDropped {
		t.Fatalf("status = %q", restored.Status())
	}
	if !restored.SameAs(goal) {
		t.Fatal("goals with the same id should be the same")
	}
}

func TestGoalApprovedIsOptional(t *testing.T) {
	goal := must(ledger.NewGoal(sampleGoalInput(), sampleNow))
	if _, found := goal.Approved(); found {
		t.Fatal("new goal should have no approval")
	}

	approval := must(kernel.ParseApproval(kernel.ApprovalInput{
		By: "human:jane",
		At: "2026-11-22",
	}))
	restored := ledger.RestoreGoal(ledger.GoalState{Approved: approval})

	if _, found := restored.Approved(); !found {
		t.Fatal("restored goal should keep its approval")
	}
}

package memory_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
)

func sampleDecisionInput() memory.DecisionInput {
	return memory.DecisionInput{
		ID:     must(memory.ParseDecisionID("bc-a1b2")),
		Author: must(kernel.ParseActor("agent:sample-agent")),
		What:   sampleText("Short sessions"),
		Why:    sampleText("Limits the damage of a leaked token"),
		Where: []kernel.PathGlob{
			must(kernel.ParsePathGlob("src/auth/**")),
		},
		Rejected: []kernel.RecordText{
			sampleText("Long sessions: a leaked token lasts days"),
		},
		Review: must(memory.ParseReviewSpan("6mo")),
	}
}

func TestNewDecisionStartsActiveAndDated(t *testing.T) {
	decision, err := memory.NewDecision(sampleDecisionInput(), sampleNow)

	if err != nil {
		t.Fatal(err)
	}
	if decision.Status() != memory.DecisionActive {
		t.Fatalf("status = %q", decision.Status())
	}
	if decision.Created().String() != "2026-11-22" {
		t.Fatalf("created = %q", decision.Created())
	}
}

func TestNewDecisionRejectsMissingField(t *testing.T) {
	input := sampleDecisionInput()
	input.Why = kernel.RecordText{}

	_, err := memory.NewDecision(input, sampleNow)

	requireCode(t, err, memory.DecisionFieldMissing)
}

func TestDecisionProtectsItsLists(t *testing.T) {
	input := sampleDecisionInput()
	decision := must(memory.NewDecision(input, sampleNow))
	glob := must(kernel.ParsePathGlob("changed/**"))

	input.Where[0] = glob
	input.Rejected[0] = sampleText("changed")
	decision.Where()[0] = glob
	decision.Rejected()[0] = sampleText("changed")

	if decision.Where()[0].String() != "src/auth/**" {
		t.Fatalf("where = %v", decision.Where())
	}
	if decision.Rejected()[0].String() == "changed" {
		t.Fatalf("rejected = %v", decision.Rejected())
	}
}

func TestDecisionOptionalFields(t *testing.T) {
	decision := must(memory.NewDecision(sampleDecisionInput(), sampleNow))

	if _, found := decision.Supersedes(); found {
		t.Fatal("supersedes should be absent")
	}
	if _, found := decision.Reviewed(); found {
		t.Fatal("reviewed should be absent")
	}
	if _, found := decision.Approved(); found {
		t.Fatal("approved should be absent")
	}
	if _, found := decision.Review(); !found {
		t.Fatal("review should be present")
	}
}

func TestRestoreDecisionKeepsStateAndIdentity(t *testing.T) {
	decision := must(memory.NewDecision(sampleDecisionInput(), sampleNow))

	restored := memory.RestoreDecision(memory.DecisionState{
		ID:       decision.ID(),
		Status:   memory.DecisionSuperseded,
		Reviewed: must(kernel.ParseCalendarDate("2026-11-23")),
	})

	if restored.Status() != memory.DecisionSuperseded {
		t.Fatalf("status = %q", restored.Status())
	}
	if !restored.SameAs(decision) {
		t.Fatal("decisions with the same id should be the same")
	}
	if _, found := restored.Reviewed(); !found {
		t.Fatal("reviewed should be kept")
	}
}
